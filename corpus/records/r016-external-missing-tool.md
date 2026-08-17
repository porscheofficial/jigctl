---
id: HCR-0161
title: "External bindings must name the tool"
scope: repo
regulates: compliance
summary: "The external binding omits the tool field, leaving the check unattributed to any scanner."
state: enforced
enforced_by:
  - kind: external
    docs: "docs/policies/dependency-scan.md"
---

An `external` binding delegates a check to a tool the harness itself does not
run. Add the missing `tool` field naming what performs the check.

<!-- jig:expect
valid: false
covers: [R-016]
diagnostics:
  - rule: R-016
    at: /enforced_by/0
-->
