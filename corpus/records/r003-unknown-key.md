---
id: HCR-2007
title: Require design review sign-off for new public endpoints
scope: repo
regulates: architecture-fitness
summary: New public HTTP endpoints must be reviewed by the architecture group before release.
state: enforced
visibility: public
enforced_by:
  - kind: command
    run: tools/check-endpoint-review.sh
---

Open a design review request before merging a new public endpoint.
Include the request/response shape and expected call volume in the review.
Do not ship the endpoint behind a flag as a way to skip the review.

<!-- jig:expect
valid: false
covers: [R-003]
diagnostics:
  - rule: R-003
    at: ""
-->
