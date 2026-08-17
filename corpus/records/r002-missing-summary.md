---
id: HCR-2005
title: Require a threat model for new external integrations
scope: repo
regulates: security
state: draft
enforced_by:
  - kind: command
    run: tools/check-threat-model.sh
---

Before wiring up a new external integration, write a short threat model covering trust boundaries.
List what data crosses the boundary and what happens if the remote side is compromised.
Attach the threat model to the pull request that introduces the integration.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
