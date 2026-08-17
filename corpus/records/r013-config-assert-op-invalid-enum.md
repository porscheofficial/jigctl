---
id: HCR-0134
title: "Config assertion operator must be a recognised comparison"
scope: repo
regulates: compliance
summary: "The config-assert binding names a comparison operator outside the closed enum."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    path: "/replicas/minimum"
    op: startswith
    value: 3
---

The `op` field must be one of the recognised comparisons. Replace it with a
supported operator rather than inventing a comparison the harness cannot
execute.

<!-- jig:expect
valid: false
covers: [R-013]
diagnostics:
  - rule: R-013
    at: /enforced_by/0/op
-->
