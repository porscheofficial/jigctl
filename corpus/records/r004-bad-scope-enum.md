---
id: HCR-2008
title: Require ownership metadata on shared libraries
scope: team
regulates: maintainability
summary: Shared libraries must declare an owning group so support requests reach the right people.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-library-ownership.sh
---

Add an ownership entry to any library consumed by more than one service.
Keep the entry current when ownership changes hands.
A library with no declared owner is treated as unmaintained and flagged in review.

<!-- jig:expect
valid: false
covers: [R-004]
diagnostics:
  - rule: R-004
    at: /scope
-->
