---
id: HCR-0192
title: "Config-assert bindings must not carry a command-only field"
scope: repo
regulates: compliance
summary: "The config-assert binding is otherwise valid but also carries a run field, which belongs only to command bindings."
state: enforced
enforced_by:
  - kind: config-assert
    file: "config/service.yaml"
    path: "/replicas/minimum"
    op: gte
    value: 2
    run: "echo check"
---

A `config-assert` binding compares a value inside a file; it does not execute
anything, so it must not carry `run`, which belongs to `command`. Remove it.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
