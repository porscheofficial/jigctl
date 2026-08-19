---
id: HCR-0510
title: Temporary exceptions expire
scope: repo
regulates: maintainability
summary: Temporary exceptions re-fire their constraint after their declared expiry date.
state: enforced
enforced_by:
  - kind: inferential
exceptions:
  - scope: services/legacy-worker
    reason: "The migration window represented by this fixture has elapsed."
    until: "2000-01-01"
  - scope: services/permanent-adapter
    reason: "This permanent waiver proves exceptions without an expiry are skipped."
---

The first exception's past date is this fixture's single violation: its temporary
waiver has expired and the constraint must fire again. The second exception has
no `until` value so the fixture also proves permanent waivers are skipped rather
than over-reported.
