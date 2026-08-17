---
id: HCR-2018
title: Require data retention limits on audit logs
scope: repo
regulates: compliance
summary: Audit logs containing personal data must be deleted or anonymised after the retention window.
state: enforced
rationale: retention-policy
enforced_by:
  - kind: command
    run: tools/check-log-retention.sh
---

Configure the log pipeline's retention window to match the documented policy.
Anonymise fields that identify a person rather than deleting the whole record where possible.
Review the retention window whenever the policy document changes.

<!-- jig:expect
valid: false
covers: [R-023]
diagnostics:
  - rule: R-023
    at: /rationale
-->
