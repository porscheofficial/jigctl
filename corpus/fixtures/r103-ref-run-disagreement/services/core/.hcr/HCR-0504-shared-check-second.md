---
id: HCR-0504
title: Core service check, second binding
scope: service
regulates: architecture-fitness
summary: >-
  A second, independently authored sensor that claims to describe the same
  shared check as HCR-0503, but disagrees on how to run it.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: shared-check
    run: make check-b
---

This binding shares `ref: shared-check` with the binding in
`HCR-0503-shared-check-first.md`, but the two disagree on `run`
(`make check-b` here versus `make check-a` there). Neither `run` value is
path-shaped (no `/` in its first token), so this fixture stays clear of
R-104 and isolates exactly the R-103 violation.
