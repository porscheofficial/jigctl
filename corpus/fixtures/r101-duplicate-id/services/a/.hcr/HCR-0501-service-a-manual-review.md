---
id: HCR-0501
title: Service A requires a manual review before every deploy
scope: service
regulates: reliability
summary: >-
  A human reviewer must sign off on any deploy of service A until automated
  coverage for its deploy path exists.
state: enforced
enforced_by:
  - kind: inferential
---

This record and `services/b/.hcr/HCR-0501-service-b-manual-review.md` both
claim `HCR-0501`. That is the fixture: two independently authored records
under sibling services reusing one id. Neither record supersedes the other;
they are simply a collision, which is exactly what R-101 exists to catch
across the whole tree, not per service.
