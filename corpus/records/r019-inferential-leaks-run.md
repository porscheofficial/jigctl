---
id: HCR-0196
title: "Inferential bindings must not carry a command-only field"
scope: repo
regulates: architecture-fitness
summary: "The inferential binding, which by definition executes nothing, also carries a run field."
state: enforced
enforced_by:
  - kind: inferential
    run: "echo check"
---

An `inferential` binding records a judgment call with no executable check at
all. `run` belongs to `command`; an inferential binding must not carry it.

<!-- jig:expect
valid: false
covers: [R-019]
diagnostics:
  - rule: R-019
    at: /enforced_by/0
-->
