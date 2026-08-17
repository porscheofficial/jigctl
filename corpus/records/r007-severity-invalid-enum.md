---
id: HCR-0071
title: "Migration verification severity must be a recognised value"
scope: repo
regulates: reliability
summary: "The migration verification binding declares a severity outside the closed enum."
state: enforced
enforced_by:
  - kind: command
    run: "scripts/verify-migrations.sh"
    severity: mandatory
---

Run the migration verification script before merging any schema change, and
treat a failing run as a signal the migration is not safe to ship. If this
check fires, roll the migration back rather than patching around it.

<!-- jig:expect
valid: false
covers: [R-007]
diagnostics:
  - rule: R-007
    at: /enforced_by/0/severity
-->
