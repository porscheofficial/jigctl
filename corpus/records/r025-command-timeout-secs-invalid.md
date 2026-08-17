---
id: HCR-0251
title: "Command timeout must be a positive integer"
scope: repo
regulates: reliability
summary: "The command binding sets timeout_secs to zero, which cannot bound a running process."
state: enforced
enforced_by:
  - kind: command
    run: "scripts/integration-test.sh"
    timeout_secs: 0
---

A timeout of zero never lets the command run at all. Set `timeout_secs` to a
positive integer that reflects how long the check may reasonably take.

<!-- jig:expect
valid: false
covers: [R-025]
diagnostics:
  - rule: R-025
    at: /enforced_by/0/timeout_secs
-->
