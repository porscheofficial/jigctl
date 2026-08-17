---
id: HCR-0002
title: Coverage floor with a retired legacy threshold
scope: service
regulates: reliability
summary: Overall test coverage must stay at or above the agreed floor, and a retired legacy threshold key must no longer be present.
state: enforced
enforced_by:
  - kind: config-assert
    file: "coverage/summary.json"
    path: "/legacy/deprecatedThreshold"
    op: absent
    severity: blocking
    cadence: [on-change, ci]
  - kind: config-assert
    file: "coverage/summary.json"
    path: "/coverage/minimum"
    op: gte
    value: 80
    severity: blocking
    cadence: [on-change, ci]
---
This service's coverage report must show at least 80% overall coverage, and
the old per-module threshold key must be fully removed rather than left at
zero. If either assertion fails, raise coverage or delete the stale config
key — do not lower the floor just to make the check pass.

<!-- jig:expect
valid: true
covers: [R-013, R-014, R-027]
-->
