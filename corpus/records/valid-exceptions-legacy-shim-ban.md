---
id: HCR-0007
title: Ban the deprecated logging shim in service code
scope: service
regulates: maintainability
summary: Forbid importing the legacy logging shim now that the standard logger wraps its behaviour directly.
state: enforced
enforced_by:
  - kind: grep
    file: "src/**/*.py"
    forbid: ["from legacy_shim import logger"]
    severity: blocking
    cadence: [on-change, ci]
exceptions:
  - scope: services/reporting-worker
    reason: "Reporting worker still depends on a transitive package pinned to the old shim; tracked for removal."
    until: "2024-01-15"
---
Do not import the legacy logging shim in new or modified code — use the
standard logger directly. The reporting worker carries a temporary, dated
exception for a transitive dependency it does not yet control; every
other service must stay clean.

<!-- jig:expect
valid: true
covers: [R-015, R-020, R-021]
-->
