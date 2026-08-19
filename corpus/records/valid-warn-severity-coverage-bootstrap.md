---
id: HCR-0009
title: Bootstrap coverage floor, blocking check in warn mode
scope: repo
regulates: reliability
summary: New coverage floor for changed packages; the binding is blocking but the record itself is still in warn state during rollout.
state: warn
enforced_by:
  - kind: config-assert
    file: "coverage/summary.json"
    path: "/coverage/minimum"
    op: gte
    value: 70
    severity: blocking
    cadence: [on-change, ci]
---
This coverage floor is a blocking check, but the record is still marked
`warn` while teams migrate existing gaps — a failure is surfaced, not
merge-blocking, until the record's state graduates to `enforced`. Do not
lower the threshold just to make a failure disappear.

<!-- jig:expect
valid: true
covers: [R-006, R-007]
-->
