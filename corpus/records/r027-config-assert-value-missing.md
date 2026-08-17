---
id: HCR-0271
title: "Config assertions with a comparison must carry a value"
scope: repo
regulates: compliance
summary: "The config-assert binding compares with gte but omits the value to compare against."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    path: "/replicas/minimum"
    op: gte
---

Any comparison other than `absent` needs something to compare against. Add
the missing `value` field, or switch `op` to `absent` if presence is truly
all that matters.

<!-- jig:expect
valid: false
covers: [R-027]
diagnostics:
  - rule: R-027
    at: /enforced_by/0
-->
