---
id: HCR-2017
title: Require a single canonical config format per service
scope: service
regulates: maintainability
summary: A service must configure itself from exactly one canonical file format, not a mixture.
state: enforced
supersedes: HCR-42
enforced_by:
  - kind: command
    run: tools/check-config-format.sh
---

Pick one configuration format for the service and remove any competing legacy format.
Update every reader and writer of configuration to agree on the same file.
This record replaces an earlier, narrower version scoped to a single service only.

<!-- jig:expect
valid: false
covers: [R-022]
diagnostics:
  - rule: R-022
    at: /supersedes
-->
