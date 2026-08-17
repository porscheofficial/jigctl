---
id: HCR-2015
title: Require pagination on list endpoints
scope: repo
regulates: reliability
summary: Any endpoint returning a list must support pagination rather than returning an unbounded set.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-list-pagination.sh
exceptions:
  - scope: services/reporting-api
    reason: Reporting API predates the pagination convention and is being migrated incrementally.
    until: 20260101
---

Add a page-size limit and a cursor or offset parameter to any endpoint returning a collection.
Document the default and maximum page size in the endpoint's summary.
An unbounded list endpoint risks timing out or exhausting memory under real data volume.

<!-- jig:expect
valid: false
covers: [R-021]
diagnostics:
  - rule: R-021
    at: /exceptions/0/until
-->
