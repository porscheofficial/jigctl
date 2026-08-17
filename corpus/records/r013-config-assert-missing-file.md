---
id: HCR-0131
title: "Config assertions must name a target file"
scope: repo
regulates: compliance
summary: "The config-assert binding omits the file field, leaving no target for the assertion."
state: enforced
enforced_by:
  - kind: config-assert
    path: "/replicas/minimum"
    op: equals
    value: 3
---

A `config-assert` binding checks a value inside a specific configuration
file. Add the missing `file` field naming the repo-relative path to check.

<!-- jig:expect
valid: false
covers: [R-013]
diagnostics:
  - rule: R-013
    at: /enforced_by/0
-->
