---
id: HCR-0101
title: "Binding kind must be one of the six recognised sensors"
scope: repo
regulates: architecture-fitness
summary: "The binding declares a kind that is not one of the six recognised sensor types."
state: enforced
enforced_by:
  - kind: monitoring-scan
    severity: blocking
    cadence: [ci]
---

Every enforcement binding must use a recognised sensor kind so the harness
knows how to execute it. An unrecognised kind means the binding cannot run at
all, which is worse than a failing check.

<!-- jig:expect
valid: false
covers: [R-010]
diagnostics:
  - rule: R-010
    at: /enforced_by/0/kind
-->
