---
id: HCR-0509
title: Rationale references resolve to decision artifacts
scope: repo
regulates: maintainability
summary: >-
  Every mapped rationale reference must resolve to an artifact in the
  repository tree.
state: enforced
enforced_by:
  - kind: inferential
rationale: ADR-0009
---

`rationale: ADR-0009` is this fixture's single violation. The mapped
`docs/adr/` directory contains a real artifact for ADR-0001, but no artifact
for ADR-0009. This ensures the fixture falsifies target resolution rather than
merely proving that a mapped directory is missing.
