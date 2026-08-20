---
id: HCR-0410
title: Invocation time must not be stored in a package variable
scope: repo
regulates: reliability
summary: "Go code must read time once per invocation and pass it as an argument, never expose a package-level swappable clock."
state: enforced
rationale: ADR-0006
enforced_by:
  - kind: grep
    file: "**/*.go"
    forbid: ["= time.Now"]
exceptions:
  - scope: internal/runner/command.go
    reason: "Runner measures duration of commands using time.Now."
  - scope: internal/runner/command_test.go
    reason: "Tests command duration measurement."
  - scope: internal/runner/live.go
    reason: "Live view measures elapsed time of a running binding using time.Now."
---
Read the current time once at the invocation boundary, convert it to the
UTC date used by validation, and thread that value through ordinary
arguments. If a test needs another date, pass it to the rule under test;
do not add `var nowFn = time.Now` or any equivalent package-level clock
that tests mutate.

The forbidden pattern is intentionally `= time.Now`, not `time.Now`.
It catches an assignment such as `var nowFn = time.Now` whether the
executor treats entries as literals or regular expressions, while it
leaves the legitimate `utcDateOf(time.Now())` invocation alone because
that call has no `= ` before `time.Now`. Do not broaden the pattern and
turn the invocation boundary itself into a violation.

Classified `reliability`: reading time repeatedly can split one run
across two dates, while a swappable package variable introduces shared
mutable state into shuffled tests. Both increase the probability of a
non-reproducible validation or test failure rather than merely making a
future edit harder.
