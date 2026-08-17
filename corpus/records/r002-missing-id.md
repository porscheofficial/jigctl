---
title: Require test coverage floor for new modules
scope: repo
regulates: reliability
summary: New modules must ship with a minimum line-coverage floor enforced in CI.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-coverage-floor.sh
---

When adding a new module, include tests that meet the coverage floor before merging.
Prefer testing behaviour at the boundary rather than internal implementation details.
Coverage gaps are flagged in review; do not suppress the check to merge faster.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
