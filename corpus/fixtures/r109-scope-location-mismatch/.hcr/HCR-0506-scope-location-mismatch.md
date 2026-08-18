---
id: HCR-0506
title: Deploy notes are attached before every release
scope: service
regulates: maintainability
summary: >-
  Release notes must accompany every deploy; this record is deliberately
  filed at the tree root while declaring service scope.
state: enforced
enforced_by:
  - kind: inferential
---

This record lives directly inside the tree root's `.hcr/`, which requires
`scope: repo`, but it declares `scope: service`. That mismatch is the
fixture's single violation: R-109 requires a record's declared `scope` to
agree with the directory tier it was discovered in.
