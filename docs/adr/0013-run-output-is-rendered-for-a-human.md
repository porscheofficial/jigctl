---
status: accepted
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

## Notes

### 2026-08-20

The Decision above puts each finding on its own `path:line` line. That
holds for one finding shape and not for the rest, which was not known
when this record was accepted.

A finding addresses a file and a pointer, and the pointer means a
different thing per binding kind. A `grep` forbid-match puts a line
number there and is the only shape an editor can open. A `grep`
missing-require names the glob pattern instead, because no single file
is the offender. A `config-assert` finding carries an RFC 6901 pointer,
which addresses a node rather than a line. A `command` finding has no
locus at all, because an exit code has no location.

What this refines is the syntax, not the decision. The detail block is
still the part a reader navigates from. Navigability is per-shape: a
line number renders as one only where a line number exists, and the
other shapes render what they actually address. Padding them into a
uniform syntax would send an editor to a line that is not there, which
is worse than declining to offer the jump.

Extending the locus with a line field was considered and rejected. A
pointer is not a degenerate line number but a different kind of address,
and flattening the two would cost the distinction that makes a
config-assert finding actionable.

One question this record did not consider is left open. A record bundles
the guidance an agent reads with the binding that verifies it, and the
detail block renders the summary but not the guidance body, which is
discarded at parse time. Whether that body belongs in the detail block,
and in what form, is a decision this record did not make and does not
foreclose.

### 2026-08-20 (second)

Two of the decisions above have been reversed, and one constant they left
implicit has been made a variable.

The run no longer closes with the fixed-field ledger. Its stated purpose was
to satisfy ADR-0012's unconditional-counts obligation by construction, and
that obligation has been narrowed to `--plain` in a Note on that record. On
a green run the ledger was five counts, four of them zero, restating what
the absence of a detail block already said. On a broken one it restated in
aggregate what the detail block had just said per record. The report now
ends on its last detail entry, and ends on the scan list when nothing failed.

A progress indicator is no longer rejected. The rejection assumed live
output and byte-determinism need two rendering paths. They do not, because
the live frames are not output. The view paints one line per binding into
the terminal's alternate use of the cursor, repaints them on a tick, and
on completion moves the cursor back over the block and erases it. `Render`
then prints the settled report into the space the block occupied. The bytes
that survive the run are produced by the same call as before, from the same
rows, in the same order — a live run and a piped run differ in what was
transiently on screen, not in what the terminal is left holding.

The live view is off unless stdout is a terminal, which the determinism
tests are not, and off additionally under `--plain`, under `TERM=dumb`,
below sixty columns, and when the block would not fit the window. It reads
its terminal size in `cmd/jigctl` and receives it as a number, so the
renderer still does not learn what a terminal is.

A binding that has not started is a dim `·`. A running one is a braille
spinner beside a timer that counts up in tenths, so a reader watching a
command that takes eight seconds can see it is progressing rather than
hung, and can see which record it is spending them on. A settled one is
replaced in place by the exact line the final report will carry, which
makes the transition to the settled block invisible rather than a redraw.
Braille was chosen over a clock face because emoji occupy two cells in some
terminals and one in others, and every column right of the timer would
shift by terminal.

The scan list's title column was a constant forty runes, which truncated
most real record titles. It is now sized from the longest title in the run,
bounded by what the terminal can give it after the fixed columns and a
sixteen-cell floor for the evidence column, and the detail block wraps to
the terminal's width rather than to seventy-two. A destination with no
measurable width — a pipe, a file, a test buffer — truncates nothing and
sizes the column purely by content, which is what keeps rendering a pure
function of the rows wherever a hash observes it.
