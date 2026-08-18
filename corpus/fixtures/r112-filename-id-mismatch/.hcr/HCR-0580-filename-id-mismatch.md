---
id: HCR-0508
title: Filenames stay in sync with their own id
scope: repo
regulates: maintainability
summary: >-
  This record's filename was mistyped when it was renamed and now begins
  with a transposed id.
state: enforced
enforced_by:
  - kind: inferential
---

This file is named `HCR-0580-filename-id-mismatch.md`, but its `id` field is
`HCR-0508` (the two middle digits transposed) -- a realistic typo. The
filename does not begin with the record's own id, which is the fixture's
single violation: R-112 requires `filepath.Base(path)` to start with the
declared `id`.
