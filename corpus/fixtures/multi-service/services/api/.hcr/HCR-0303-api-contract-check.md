---
id: HCR-0303
title: API service enforces its request/response contract check
scope: service
regulates: architecture-fitness
summary: The api service must run the shared contract checker against its OpenAPI spec before merge.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: api-contract-check
    run: scripts/check-api-contracts.sh
---

Run `scripts/check-api-contracts.sh` before merging any change touching
the api service's request or response shapes. It diffs the live OpenAPI
spec against the last published contract and fails on a breaking change
that was not explicitly versioned.

