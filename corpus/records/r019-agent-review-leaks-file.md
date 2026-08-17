---
id: HCR-0195
title: "Agent-review bindings must not carry a file-scanning field"
scope: repo
regulates: behaviour
summary: "The agent-review binding is otherwise valid but also carries a file field, which belongs only to grep and config-assert bindings."
state: enforced
enforced_by:
  - kind: agent-review
    prompt: "Confirm migration scripts are reversible."
    file: "migrations/*.sql"
---

An `agent-review` binding judges something via a prompt, not by scanning a
named file directly. `file` belongs to `grep` and `config-assert`; remove it
here.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
