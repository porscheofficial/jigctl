---
id: HCR-2009
title: Require load testing before capacity-sensitive releases
scope: service
regulates: performance
summary: Releases that change hot-path request handling must include a load test comparison.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-load-test-report.sh
---

Run the standard load test profile against the changed endpoints before release.
Compare latency and error-rate results against the last known-good baseline.
Attach the comparison report to the release checklist.

<!-- jig:expect
valid: false
covers: [R-005]
diagnostics:
  - rule: R-005
    at: /regulates
-->
