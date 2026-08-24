---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0012: A check that did not run is not a pass

## Context

A runner can finish while leaving a declared check unevaluated. Treating
every such case alike creates a partial false green: other checks pass, one
binary is missing, and the invocation exits successfully. The absent check
is indistinguishable from verified compliance.

Kyverno maintainers rejected this behavior because it would silently pass
and be hard to explain. A 2026 `cargo-audit` incident demonstrated the same
failure in practice. The binding set is closed, so the runner can classify
every outcome.

Unchecked work has two different meanings. Some checks are deliberately not
attempted because their kind, cadence, or record state says that they do not
run in this invocation. Other checks were selected and expected to run but
could not finish. Only the former is expected non-execution.

## Decision

An outcome has two axes. Its completion state is `completed`, `blocked`,
`not-attempted`, or `operational`. Its reported projection is `pass`,
`violation`, `expected-unchecked`, `blocked-unchecked`, or `operational`.
Findings remain a collection carried beside those axes; each retains its
locus, resolved severity, and `waived_by` exception identities.

The unchecked reason is always reported, using the existing diagnostic shape
`path:pointer: code: message`. The summary always prints expected-unchecked
and blocked-unchecked counts, including when each count is zero. Counts are
over verdict instances. Raw and waived findings are counted separately.

The runner taxonomy is exhaustive:

| No. | Outcome | Completion | Projection |
| ---: | --- | --- | --- |
| 1 | Command exits non-zero | `completed` | `violation` |
| 2 | Executable is absent from `PATH` | `blocked` | `blocked-unchecked` |
| 3 | Executable permission is denied | `blocked` | `blocked-unchecked` |
| 4 | Timeout expires | `blocked` | `blocked-unchecked` |
| 5 | Process is killed by a signal | `completed` | `violation` |
| 6 | Execution authorization is absent | `blocked` | `blocked-unchecked` |
| 7 | Grep glob matches no files | `blocked` | `blocked-unchecked` |
| 8 | Configuration data file is missing | `blocked` | `blocked-unchecked` |
| 9 | `absent` pointer does not resolve | `completed` | `pass` |
| 10 | Data is unreadable or malformed | `blocked` | `blocked-unchecked` |
| 11 | Kind cannot execute | `not-attempted` | `expected-unchecked` |
| 12 | RFC 6901 pointer is malformed | `blocked` | `blocked-unchecked` |
| 13 | `matches` cannot compile | `blocked` | `blocked-unchecked` |
| 14 | Scope has no valid shape | `blocked` | `blocked-unchecked` |
| 15 | Grep glob syntax is invalid | `blocked` | `blocked-unchecked` |
| 16 | Command argv cannot be split | `blocked` | `blocked-unchecked` |
| 17 | Format is unsupported | `blocked` | `blocked-unchecked` |
| 18 | Binding has `pattern` or `select` | `blocked` | `blocked-unchecked` |
| 19 | Other process-start failure | `operational` | `operational` |
| 20 | Output or read limit exceeded | `operational` | `operational` |
| 21 | Invocation is cancelled | `operational` | `operational` |

Row 11 includes `inferential`, `external`, and `agent-review` bindings.
Expected-unchecked also covers a binding excluded by cadence and a record in
`draft` or `deprecated`. These reasons are distinguished in output even
though they share a projection.

Row 19 includes a process-start failure other than not-found or denied
permission. It also includes an authored path that escapes the tree root.

A live exception is not unchecked. The binding completes, each suppressed
finding records its waiver metadata, and the projection is `pass` when no
unwaived finding remains. This reports both what was verified and what was
waived instead of understating one while overstating skipped work.

The precedence rule is: a binding that did not complete is never reported as
a pass. A grep that finds a forbidden literal and then cannot read another
file is `blocked` and carries its partial finding. Both reach the report;
the exit code takes the worse result.

Every runner outcome lands in exactly one of the five projections. Discovery
of a sixth projection during implementation falsifies this decision; the
implementation taxonomy test must reject that omission.

A valid tree with no records yields no bindings. It aggregates to exit 77
because no check ran. A tree that fails validation never reaches the runner;
its diagnostics produce exit 1, including R-104 and R-107 diagnostics.

| Exit | Meaning |
| ---: | --- |
| 0 | At least one real result, with no gating failure |
| 1 | Validation, blocked work, or a gating violation |
| 2 | An operational failure of jigctl or the invocation |
| 77 | Every verdict is expected-unchecked, or there are no verdicts |

Exit 2 retains the operational meaning already assigned by the command root.
An advisory violation with `warn` severity is reported but does not gate and
does not by itself produce exit 1. Blocked-unchecked is a failure by
default and does produce exit 1.

Exit 77 is returned only when every verdict is expected-unchecked and none
is blocked, including the vacuous case of no verdicts; one real result makes
77 unreachable.

Strict mode promotes expected-unchecked to a gating failure and exit 1. It
does not alter blocked or operational classification. The `dogfood` task
must not enable strict mode because the deliberately inferential
fixture-policy record would otherwise make that task fail permanently.

## Consequences

Expected omission remains visible without making ordinary mixed runs fail.
Failure to perform selected work can no longer borrow success from completed
checks. Operational failures remain separate from defects in a record or its
declared environment.

The report can show a partial finding and incomplete evaluation together.
Consumers must preserve both axes rather than reducing a verdict to one
scalar state.

Waivers remain evidence of evaluation rather than evidence of skipping.
Summary counts expose unchecked work even when there are no findings, and
strict callers may reject deliberate omissions without changing their normal
meaning.

## Notes

### 2026-08-20

The Decision above requires the summary to print its expected-unchecked and
blocked-unchecked counts including when each count is zero. That obligation
was written before a report had any other place to state unchecked work.
ADR-0013 gave it one: the detail block repeats every non-passing binding
individually, and an unchecked binding is not a passing one, so it is named
there with its reason code whether or not anything else failed.

The counts have therefore been dropped from the human-facing report, which
now ends on its last detail entry. What the obligation protected is
unaffected. A count of zero told the reader nothing they could act on, and
a count above zero told them less than the entry that names the record.

The obligation stands unchanged for `--plain`, which has no detail block and
still prints both counts unconditionally. A consumer that parses counts
parses that output, which is what ADR-0013 made it for.
