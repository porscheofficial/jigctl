---
id: HCR-0193
title: "Grep bindings must not carry a command-only field"
scope: repo
regulates: security
summary: "The grep binding is otherwise valid but also carries a timeout_secs field, which belongs only to command bindings."
state: enforced
enforced_by:
  - kind: grep
    file: "**/*.py"
    forbid: ["print("]
    timeout_secs: 30
---

A `grep` binding is a pattern scan, not a process with a runtime budget.
`timeout_secs` belongs to `command`; remove it from this binding.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
