---
id: HCR-2004
title: Require rollback plans for schema migrations
scope: service
summary: Every migration must document a tested rollback path before it ships.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-migration-rollback.sh
---

Write the down-migration alongside the up-migration and run both in CI.
If a migration is genuinely irreversible, document why and get a second reviewer.
Never merge a migration whose rollback has not been exercised at least once.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
