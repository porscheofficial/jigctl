---
id: HCR-0503
title: Core service check, first binding
scope: service
regulates: architecture-fitness
summary: >-
  A sensor for the core service's shared check, declared from this record.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: shared-check
    run: make check-a
---

This binding shares `ref: shared-check` with the binding in
`HCR-0504-shared-check-second.md`, but the two disagree on `run`
(`make check-a` here versus `make check-b` there). That disagreement is the
fixture's single violation: R-103 requires every pair of bindings sharing a
`ref` to carry byte-identical `run` values, and the two records living in the
same service directory (rather than one record with two bindings) proves the
comparison crosses record boundaries.
