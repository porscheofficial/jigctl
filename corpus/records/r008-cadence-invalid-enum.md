---
id: HCR-0081
title: "Migration verification cadence must use recognised values"
scope: repo
regulates: reliability
summary: "The migration verification binding schedules a cadence value outside the closed enum."
state: enforced
enforced_by:
  - kind: command
    run: "scripts/verify-migrations.sh"
    cadence: [nightly]
---

Run the migration verification script on every schema change so a broken
migration is caught before it reaches a shared environment. Prefer fixing the
migration itself over silencing this check.

<!-- jig:expect
valid: false
covers: [R-008]
diagnostics:
  - rule: R-008
    at: /enforced_by/0/cadence/0
-->
