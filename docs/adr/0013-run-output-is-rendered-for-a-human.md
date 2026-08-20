---
status: proposed
date: 2026-08-20
decision-makers: [Patrice Bouillet]
---

# ADR-0013: Run output is rendered for a human

## Context

`jigctl run` prints one line per binding in the shape `path: code:
message`, inherited from `jigctl validate`. That shape is right for a
validator, where every line is a problem. A run is not a validation. Most
of its lines are a pass, and spending the diagnostic shape on a record
that passed spends the reader's attention on the records that need none.
The one line that matters — the unchecked record ADR-0012 exists to make
impossible to overlook — sits between two passes looking exactly like
them.

Ten comparable tools were read for their actual terminal output before
choosing a shape. The finding that mattered most was negative: the
obvious answer, a name bound to a status word across leader dots, is
exactly what `pre-commit` does, and it is flat for the same reason this
output is flat. It carries no glyph, gives a passing item the same
visual weight as a failing one, and never repeats a failure at the end,
so the reader scrolls back up to find what broke.

The tools that read well combine four separate mechanisms. `vitest`
achieves the densest healthy line by putting a glyph first and pushing
counts and timing to the right. `terraform plan` prints a legend of its
symbols once before using them, and hides unchanged detail rather than
dimming it. `pytest` and `dbt` repeat every failure in a block after the
list, so the reader never scrolls. `dbt` closes with a fixed-field ledger
whose zero counts are still printed, so the shape never reflows. Two
tools go further and make success nearly silent — `go test` reduces a
package to `ok pkg 0.364s`, `cargo nextest` prints failures only — which
is the right instinct for a test runner and the wrong one here, because
a harness is run precisely to see that its constraints are enforced.

The Command Line Interface Guidelines settle the question this leaves
open. Human-readable output is paramount, and where it conflicts with
machine-readable output the answer is a `--plain` flag rather than
refusing to restructure. They also require four colour-disabling
triggers, not the two an earlier draft of this record assumed.

Four existing constraints bound any answer. Rendering is hash-tested for
determinism. Findings go to stdout and operational errors to stderr. No
unchecked outcome may render as a pass, and outcomes sharing the
unchecked projection must stay distinguishable. The summary states its
counts whether or not anything was found.

## Decision

Default output has two parts: a scan list of one line per binding, and a
detail block repeating only the bindings that did not pass.

```
  ✓  HCR-0404  go-source-passes-gates             2.85s  mise run lint
  ✓  HCR-0405  mise-owns-every-entrypoint
  ✗  HCR-0406  single-hcr-implementation                 1 finding
  ○  HCR-0407  fixtures-never-edited-to-pass             kind-not-executable
  ✓  HCR-0408  real-record-ids-stay-in-band        124ms  tools/check-record-ids.py

  ✗  HCR-0406  single-hcr-implementation
       A single implementation of HCR validation must exist, so that the
       CLI and the test suite cannot disagree about what is valid.
       cmd/jigctl/main.go:6

  Done. PASS=8 VIOLATION=1 UNCHECKED=1 BLOCKED=0 ERROR=0
```

A glyph opens each line because it is the only element a reader can scan
without reading. Five are needed, since a blocked binding gates the run
and an expected-unchecked one does not, and collapsing them would undo
ADR-0012. Five glyphs are more than anyone should be asked to memorise,
so a legend is printed once above the list — but only on a run that
actually uses a glyph other than the passing one, since a green run needs
no vocabulary.

The reason code stays in the right-hand column, in words, because the
glyph deliberately does not distinguish `cadence-excluded` from
`record-draft`. Colour is redundant with both glyph and word and never
carries meaning alone.

Durations are unit-scaled and rounded to roughly three significant
figures, right-aligned in a fixed column. Nanosecond precision claims an
accuracy that does not exist and churns between runs.

The detail block repeats each non-passing binding with its summary and
its findings, each finding on its own `path:line` line so an editor can
jump to it. This is the only part of the output a reader navigates from,
so it is the only part that keeps the diagnostic shape.

The run closes with a fixed-field ledger printing every count including
the zeroes, which satisfies ADR-0012's unconditional-counts obligation by
construction rather than by remembering to.

`--plain` restores one record per line in the current shape, with no
glyph, colour, alignment or detail block, for grep and awk. Colour is
disabled when stdout is not a terminal, when `NO_COLOR` is present and
non-empty whatever its value, when `TERM` is `dumb`, when `--no-color` is
passed, and when `JIGCTL_NO_COLOR` is set. The decision is made in
`cmd/jigctl` and passed to the renderer, which writes to an `io.Writer`
and must not learn what a terminal is.

## Consequences

The determinism test is unaffected: it renders to a buffer, which is not
a terminal, so colour is off and never enters a hash. Rounded durations
make it more stable, not less.

Piping now changes the shape, not only the colour, which an earlier draft
of this record forbade. That prohibition was wrong. A pipe is not a
reader, and the guidelines are explicit that the human case wins and the
machine case gets a flag; `bats` forks its rendering on the same
detection and is no less trustworthy for it. `--plain` is what a script
pins itself to, and unlike the flags the previous milestone excluded it
exists to serve composability rather than to add a second format.

Every non-passing binding is now printed twice. On a healthy run this
costs nothing, since nothing qualifies. On a broken one it is the point.

Aligning columns means the renderer must see every row before printing
the first; it already sorts before printing, so this costs nothing new.

One question is left open. Detecting a terminal from the standard library
alone means testing stdout for a character device, which reports
`/dev/null` as a terminal and would colour a redirect to it. Doing it
correctly means `golang.org/x/term`, maintained by the Go team, as a
second direct dependency after doublestar. The milestone added exactly
one, so the second is worth taking deliberately.

Two alternatives were rejected. Printing only the failures, as `go test`
and `cargo nextest` do, was rejected because a constraint harness is run
to confirm its constraints ran, and a silent pass cannot distinguish an
enforced rule from a forgotten one. A progress indicator was rejected
reluctantly, since bindings execute in sequence and some take seconds,
because live output and byte-determinism cannot be reconciled without
maintaining two rendering paths.
