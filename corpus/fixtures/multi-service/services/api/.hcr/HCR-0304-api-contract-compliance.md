---
id: HCR-0304
title: API contract check doubles as compliance evidence
scope: service
regulates: compliance
summary: The same contract checker that guards api's request/response shapes is also the evidence artifact compliance sign-off relies on.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: api-contract-check
    run: scripts/check-api-contracts.sh
---

This record exists so compliance sign-off can point at one command and
know it is the same check architecture relies on for contract stability.
Do not fork this into a second script: if the check needs to change,
change `scripts/check-api-contracts.sh` once and both records benefit.

