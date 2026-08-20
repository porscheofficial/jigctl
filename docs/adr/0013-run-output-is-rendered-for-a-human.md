---
status: proposed
date: 2026-08-20
decision-makers: [Patrice Bouillet]
---

# ADR-0013: Run output is rendered for a human

## Context

`jigctl run` prints one line per binding in the shape `path: code:
message`. That shape is inherited from `jigctl validate`, where it is
right: every line a validator emits is a problem, and `path: code:
message` is what editors and greps already understand. A run is not a
validation. Most of its lines are not problems, and spending the
diagnostic shape on a record that passed spends the reader's attention on
the records that need none.

The result is that the one line that matters is the least visible thing
in the output. Against this repository the run prints ten lines, nine of
which are a pass. The tenth is unchecked — the outcome ADR-0012 exists
to make impossible to overlook — and it sits between two passes looking
exactly like them. There is no colour, no alignment, and no visual
hierarchy of any kind. Every distinction the run computes is present in
the text and absent to the eye.

Shortening the lines helped and did not solve this. The line is now
readable in isolation; ten of them still read as a wall.

Four existing constraints bound any answer. Rendering is hash-tested for
determinism, so five renders must produce identical bytes. Findings go to
stdout and operational errors to stderr, and the two never mix. No
unchecked outcome may render as a pass, and outcomes that share the
unchecked projection must stay distinguishable from each other. The
summary line states its unchecked counts whether or not anything was
found.

## Decision

The default rendering of `jigctl run` targets a human reader. Each
binding becomes a row with a status column, aligned across the run and
coloured when the destination is a terminal:

```
  pass       HCR-0404  go-source-passes-gates          3.60s  mise run lint
  unchecked  HCR-0407  fixtures-never-edited-to-pass          kind-not-executable
  violation  HCR-0406  single-hcr-implementation              1 finding
      cmd/jigctl/main.go: a single implementation of HCR validation must exist
```

The status column carries the projection, or for an unchecked or blocked
row the reason code. This is the vocabulary the run already speaks; no
glyph alphabet is introduced. Colour is redundant with the word and never
a substitute for it, so the output means the same thing to a reader who
cannot see it, to a pipe, and to a log.

Findings keep a `path: message` line, indented beneath the row that
produced them. A finding is the one thing a reader navigates to, so it
stays in the shape an editor can jump from. The rows above exist for
orientation, not navigation.

Colour is emitted when stdout is a terminal and never otherwise. A
`NO_COLOR` variable that is present and not empty, whatever its value,
disables it; this is the informal standard published at no-color.org and
honoured by ripgrep, fd, gh, jq, ruff, Nix, zig and deno. A `--no-color`
flag overrides the environment, following that standard's own guidance
that a per-invocation argument outranks a variable.

Whether stdout is a terminal is decided in `cmd/jigctl` and passed to the
renderer as a field alongside the existing normalisation flag. The
renderer writes to an `io.Writer` and must not learn what a terminal is.

## Consequences

The determinism test is unaffected. It renders to a buffer, which is not
a terminal, so colour is off and never enters a hash. For the same reason
`jigctl run . | grep` sees what it saw before colour existed. Only colour
switches on detection; the structure of the output does not. A tool whose
shape changes when piped is a tool whose output cannot be trusted in a
script.

ADR-0012's obligations survive intact. Putting the reason code in the
status column is what keeps `cadence-excluded` and `record-draft` from
collapsing into one another, and the summary line is untouched.

Aligning columns means the renderer must see every row before printing
the first. It already sorts all verdicts before printing, so this costs
nothing new.

One question is left open for the owner. Detecting a terminal from the
standard library alone means calling `Stat` on stdout and testing for a
character device, which reports `/dev/null` as a terminal and would
colour a redirect to it. Doing it correctly means `golang.org/x/term`,
maintained by the Go team, as a second direct dependency after
doublestar. The milestone added exactly one dependency and this would be
the second, so it is a decision to take deliberately rather than to
absorb.

Three alternatives were rejected. Switching the output *shape* on
terminal detection, as `ls` does, was rejected for the reason above.
`--color=always|auto|never` was rejected as three states for a two-state
question: automatic detection plus `--no-color` covers every case, and
the third state exists to force colour into a pipe, which nothing in this
repository's workflow wants. A progress indicator was rejected reluctantly
— bindings execute in sequence and some take seconds, so the case is real
— because live output and byte-determinism cannot be reconciled without
maintaining two rendering paths.
