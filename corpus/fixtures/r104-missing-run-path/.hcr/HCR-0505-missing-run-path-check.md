---
id: HCR-0505
title: Repository runs its packaging check before release
scope: repo
regulates: reliability
summary: >-
  A packaging sensor must run before every release; the script it names does
  not exist in this fixture, which is the point.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    run: scripts/does-not-exist.sh
---

`run: scripts/does-not-exist.sh` is path-shaped (its first whitespace token
contains `/`) but `scripts/does-not-exist.sh` does not exist anywhere under
this fixture's tree root, and never will: this fixture's whole purpose is
that absence. That is the single violation R-104 exists to catch.
