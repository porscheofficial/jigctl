---
id: HCR-0181
title: "Agent review bindings must stay advisory"
scope: repo
regulates: behaviour
summary: "The agent-review binding sets severity to blocking, but a non-deterministic review cannot gate a merge."
state: enforced
enforced_by:
  - kind: agent-review
    prompt: "Confirm the pull request includes a rollback plan before merge."
    severity: blocking
---

An `agent-review` check is inherently non-deterministic and must never block
a merge outright. Remove the `severity: blocking` override and let the
binding use its advisory default.

<!-- jig:expect
valid: false
covers: [R-018]
diagnostics:
  - rule: R-018
    at: /enforced_by/0/severity
-->
