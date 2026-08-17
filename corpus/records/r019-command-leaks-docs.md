---
id: HCR-0191
title: "Command bindings must not carry an external-only field"
scope: repo
regulates: reliability
summary: "The command binding is otherwise valid but also carries a docs field, which belongs only to external bindings."
state: enforced
enforced_by:
  - kind: command
    run: "make lint"
    docs: "docs/policies/lint.md"
---

Fields belonging to one binding kind must not leak onto another. A `command`
binding runs something directly; it does not carry a `docs` pointer, which is
`external`'s field. Remove it.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
