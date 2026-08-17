---
id: HCR-0141
title: "Config assertions checking absence must not carry a value"
scope: repo
regulates: compliance
summary: "The config-assert binding checks for absence but also carries a value to compare against."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    path: "/feature/flag"
    op: absent
    value: true
---

An `op: absent` assertion checks that a key is missing entirely; there is
nothing to compare it against. Remove the `value` field rather than leaving a
comparison that can never run.

<!-- jig:expect
valid: false
covers: [R-014]
diagnostics:
  - rule: R-014
    at: /enforced_by/0
-->
