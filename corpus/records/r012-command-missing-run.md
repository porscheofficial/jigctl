---
id: HCR-0121
title: "Command bindings must declare what to run"
scope: repo
regulates: reliability
summary: "The command binding omits the run field, so there is nothing for the harness to execute."
state: enforced
enforced_by:
  - kind: command
    timeout_secs: 120
---

A `command` binding exists to execute something concrete. Add the missing
`run` field naming the script or command that performs the check.

<!-- jig:expect
valid: false
covers: [R-012]
diagnostics:
  - rule: R-012
    at: /enforced_by/0
-->
