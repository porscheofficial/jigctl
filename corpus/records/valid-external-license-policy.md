---
id: HCR-0004
title: Third-party dependency licenses must be pre-approved
scope: repo
regulates: compliance
summary: New dependencies must be checked against the approved license list before they are added to the project.
state: enforced
enforced_by:
  - kind: external
    tool: license-scanner
    docs: "docs/policies/dependency-licensing.md#approved-licenses"
    severity: blocking
    cadence: [on-change, ci]
---
Before adding a new third-party dependency, run it through the license
scanner and confirm its license appears on the approved list. If the
scanner flags an unapproved license, do not add the dependency — request
an exception from the platform team instead of silently vendoring the code.

<!-- jig:expect
valid: true
covers: [R-016]
deferred:
  - rule: R-110
    reason: requires-filesystem
-->
