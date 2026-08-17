---
id: HCR-2020
title: Require review of new cross-region data transfers
scope: repo
regulates: compliance
summary: Any new data flow that crosses a regional boundary must be reviewed for residency rules.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-data-residency.sh
exceptions:
  - just a plain string, not an exception object
---

Document the source and destination region before enabling a new cross-region data flow.
Get sign-off from the compliance reviewer before the flow carries production data.
Residency exceptions are time-boxed and must be revisited before they lapse.

<!-- jig:expect
valid: false
covers: [R-028]
diagnostics:
  - rule: R-028
    at: /exceptions/0
-->
