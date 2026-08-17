---
id: HCR-0133
title: "Config assertions must name a comparison operator"
scope: repo
regulates: compliance
summary: "The config-assert binding omits the op field, leaving the comparison unspecified."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    path: "/replicas/minimum"
---

A `config-assert` binding must state how the value at the target path is
compared. Add the missing `op` field naming the comparison to perform.

<!-- jig:expect
valid: false
covers: [R-013]
diagnostics:
  - rule: R-013
    at: /enforced_by/0
-->
