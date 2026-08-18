---
id: HCR-0507
title: Data retention policy is documented and current
scope: repo
regulates: compliance
summary: >-
  The repository's data retention window must be documented; this record
  points at an anchor that this fixture never supplies.
state: enforced
enforced_by:
  - kind: external
    severity: blocking
    cadence: [scheduled]
    tool: docs-anchor-check
    docs: docs/policy-notes.md#retention-window
---

`docs: docs/policy-notes.md#retention-window` is this fixture's single
violation. `docs/policy-notes.md` exists and contains headings (`# Policy notes`,
`## Data handling`), none of which slug to `retention-window`, so the violation
is an unresolvable anchor rather than a missing path. This matters because if
the file were absent the rule would take the path-missing branch and the slug
algorithm would never execute while `expect.yaml` still went green.
