---
id: HCR-2019
title: Require contract tests between service boundaries
scope: repo
regulates: architecture-fitness
summary: Services that call each other over HTTP must share a contract test suite.
state: enforced
enforced_by:
  - just a plain string, not a binding object
---

Write a consumer-driven contract test alongside any new inter-service HTTP call.
Run the contract suite for both sides whenever either service changes its interface.
A broken contract test blocks the release of the side that changed first.

<!-- jig:expect
valid: false
covers: [R-028]
diagnostics:
  - rule: R-028
    at: /enforced_by/0
-->
