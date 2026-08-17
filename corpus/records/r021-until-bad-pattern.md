---
id: HCR-2016
title: Require feature flags for risky rollouts
scope: repo
regulates: reliability
summary: Changes to hot-path behaviour must ship behind a feature flag with a rollback switch.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-feature-flag-usage.sh
exceptions:
  - scope: services/billing-service
    reason: Billing service rollout tooling does not yet support flag-based rollback.
    until: "2026/01/01"
---

Wrap the new code path in a feature flag so it can be disabled without a redeploy.
Keep the old path intact until the new one has run clean in production for a full cycle.
Remove the flag and the old path together once the rollout is confirmed stable.

<!-- jig:expect
valid: false
covers: [R-021]
diagnostics:
  - rule: R-021
    at: /exceptions/0/until
-->
