---
id: HCR-0132
title: "Config assertions must name a target path"
scope: repo
regulates: compliance
summary: "The config-assert binding omits the path field, leaving no location inside the file to check."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    op: equals
    value: 3
---

A `config-assert` binding checks a value at a specific pointer inside the
target file. Add the missing `path` field naming the JSON Pointer to check.

<!-- jig:expect
valid: false
covers: [R-013]
diagnostics:
  - rule: R-013
    at: /enforced_by/0
-->
