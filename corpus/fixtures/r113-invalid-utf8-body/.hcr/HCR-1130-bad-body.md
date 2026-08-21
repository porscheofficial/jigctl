---
id: HCR-1130
title: Body must be valid UTF-8
scope: repo
regulates: maintainability
summary: >-
  This record has a body that is not valid UTF-8.
state: enforced
enforced_by:
  - kind: inferential
---

This is some valid text, but followed by invalid UTF-8: ÿþ