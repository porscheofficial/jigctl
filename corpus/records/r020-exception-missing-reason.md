---
id: HCR-2014
title: Require idempotency keys on payment-initiating requests
scope: service
regulates: reliability
summary: Any request that initiates a payment must accept and honour an idempotency key.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-idempotency-key.sh
exceptions:
  - scope: services/legacy-checkout
    until: "2026-09-30"
---

Accept an idempotency key header on payment-initiating endpoints and store it with the result.
Replaying the same key must return the original result rather than initiating a second payment.
New endpoints must implement this from day one; there is no default grace period.

<!-- jig:expect
valid: false
covers: [R-020]
diagnostics:
  - rule: R-020
    at: /exceptions/0
-->
