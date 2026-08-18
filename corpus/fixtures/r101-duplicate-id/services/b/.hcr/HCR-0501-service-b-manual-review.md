---
id: HCR-0501
title: Service B requires a manual review before every deploy
scope: service
regulates: reliability
summary: >-
  A human reviewer must sign off on any deploy of service B until automated
  coverage for its deploy path exists.
state: enforced
enforced_by:
  - kind: inferential
---

This record and `services/a/.hcr/HCR-0501-service-a-manual-review.md` both
claim `HCR-0501`. Service A's effective set is repo union A; service B's is
repo union B; the two never meet under a per-effective-set reading, so only
a tree-global uniqueness check can detect this collision.
