---
id: HCR-42
title: Require changelog entries for public API changes
scope: repo
regulates: maintainability
summary: Every pull request that touches a public API must include a changelog entry describing the change.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-changelog-entry.sh
---

Add a changelog entry whenever you touch a file under a public API surface.
Describe what changed and why, in one or two sentences a consumer would understand.
If the change is purely internal refactoring with no observable effect, say so explicitly.

<!-- jig:expect
valid: false
covers: [R-001]
diagnostics:
  - rule: R-001
    at: /id
-->
