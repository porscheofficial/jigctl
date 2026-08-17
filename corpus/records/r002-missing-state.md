---
id: HCR-2006
title: Require licence allowlist checks for new dependencies
scope: repo
regulates: compliance
summary: New third-party dependencies must be checked against the approved licence allowlist.
enforced_by:
  - kind: command
    run: tools/check-licence-allowlist.sh
---

Run the licence checker before adding a new dependency to the manifest.
If a dependency's licence is not on the allowlist, get compliance sign-off before merging.
Document the approved exception in the pull request description.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
