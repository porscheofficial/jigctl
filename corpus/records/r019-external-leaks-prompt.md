---
id: HCR-0194
title: "External bindings must not carry an agent-review-only field"
scope: repo
regulates: compliance
summary: "The external binding is otherwise valid but also carries a prompt field, which belongs only to agent-review bindings."
state: enforced
enforced_by:
  - kind: external
    tool: "dependency-audit"
    docs: "docs/policies/dependency-scan.md"
    prompt: "Summarise the audit findings."
---

An `external` binding delegates to a named tool; it does not ask a reviewing
agent anything. `prompt` belongs to `agent-review`; remove it here.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
