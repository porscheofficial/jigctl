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
violation. `docs/policy-notes.md` does not exist under this fixture's tree
root and is not created here; that absence, and later the anchor mismatch
once the file is added, is exactly what R-110 exists to catch.
