---
id: HCR-0261
title: "Agent review run count must be a positive integer"
scope: repo
regulates: behaviour
summary: "The agent-review binding sets runs to zero, which asks the reviewing agent to execute zero times."
state: enforced
enforced_by:
  - kind: agent-review
    prompt: "Verify the changelog reflects every breaking API change."
    runs: 0
---

Asking a reviewing agent to run zero times is asking for nothing at all. Set
`runs` to a positive integer, or omit it entirely to use the default.

<!-- jig:expect
valid: false
covers: [R-026]
diagnostics:
  - rule: R-026
    at: /enforced_by/0/runs
-->
