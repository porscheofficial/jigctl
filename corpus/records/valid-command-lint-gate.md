---
id: HCR-0001
title: Static lint check must pass before merge
scope: repo
regulates: maintainability
summary: Every changed package must pass the linter with zero warnings before a pull request can merge.
state: enforced
enforced_by:
  - kind: command
    run: "make lint"
    timeout_secs: 120
    severity: blocking
    cadence: [on-change, ci]
---
Run `make lint` against any package you touch before opening a pull request.
Zero warnings are allowed — treat a warning as a build failure, not a
suggestion. If the linter flags something that looks wrong, fix the code
rather than suppressing the rule inline.

<!-- jig:expect
valid: true
covers: [R-012, R-025]
deferred:
  - rule: R-104
    reason: requires-filesystem
-->
