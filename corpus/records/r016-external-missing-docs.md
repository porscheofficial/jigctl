---
id: HCR-0162
title: "External bindings must link supporting docs"
scope: repo
regulates: compliance
summary: "The external binding omits the docs field, leaving the check with no explanation for a reader."
state: enforced
enforced_by:
  - kind: external
    tool: "dependency-audit"
---

An `external` binding needs a docs pointer so a reader understands what the
delegated tool checks and why. Add the missing `docs` field.

<!-- jig:expect
valid: false
covers: [R-016]
diagnostics:
  - rule: R-016
    at: /enforced_by/0
-->
