---
id: HCR-2012
title: Require rate limiting on newly exposed endpoints
scope: service
regulates: reliability
summary: Any newly exposed HTTP endpoint must be registered with a rate limit before release.
state: enforced
enforced_by: []
---

Register the endpoint's rate limit alongside its route definition, not as an afterthought.
Pick a conservative default limit and loosen it only with traffic data to justify the change.
An endpoint with no registered limit is treated as unprotected and blocked at review.

<!-- jig:expect
valid: false
covers: [R-009]
diagnostics:
  - rule: R-009
    at: /enforced_by
-->
