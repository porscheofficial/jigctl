---
id: HCR-2003
title: Forbid direct database access from HTTP handlers
regulates: architecture-fitness
summary: HTTP handlers must call into a service layer rather than querying the database directly.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-handler-boundaries.sh
---

Route all data access through a service or repository layer, never straight from a handler.
This keeps handlers thin and makes the data-access logic testable in isolation.
Reviewers should flag any handler file that imports a database driver directly.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
