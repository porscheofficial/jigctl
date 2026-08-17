---
id: HCR-0171
title: "Agent review bindings must carry a prompt"
scope: repo
regulates: behaviour
summary: "The agent-review binding omits the prompt field, leaving the reviewing agent nothing to evaluate."
state: enforced
enforced_by:
  - kind: agent-review
    grounding: ["docs/architecture.md"]
---

An `agent-review` binding asks a reviewing agent to judge something specific.
Add the missing `prompt` field stating exactly what the agent must check.

<!-- jig:expect
valid: false
covers: [R-017]
diagnostics:
  - rule: R-017
    at: /enforced_by/0
-->
