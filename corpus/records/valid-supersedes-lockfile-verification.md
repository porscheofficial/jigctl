---
id: HCR-0008
title: Lockfile verification replaces manual dependency pinning
scope: repo
regulates: reliability
summary: Automated lockfile verification supersedes the older manual pinning guidance for dependency versions.
state: enforced
enforced_by:
  - kind: command
    run: "scripts/verify-lockfile.sh"
    severity: blocking
    cadence: [on-change, ci]
supersedes: HCR-0042
rationale: ADR-0007
---
Run the lockfile verification script whenever a dependency manifest
changes. It replaces the old manual pinning checklist — do not resurrect
that checklist in a code review comment; point reviewers at this record
and the referenced decision instead.

<!-- jig:expect
valid: true
covers: [R-022, R-023]
-->
