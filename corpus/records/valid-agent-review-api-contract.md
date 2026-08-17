---
id: HCR-0005
title: Human review for public API contract changes
scope: service
regulates: architecture-fitness
summary: Any change to a service's public API contract needs an AI-assisted review pass before merge.
state: draft
enforced_by:
  - kind: agent-review
    prompt: "Review this diff for changes to the service's public API contract. Confirm backward compatibility is preserved, or that a version bump and migration note are included."
    grounding: ["docs/architecture/api-contract-guidelines.md"]
    model: "contract-review-model-v1"
    runs: 2
    severity: advisory
    cadence: [scheduled]
---
When a pull request touches a public API contract file, an automated
reviewer checks whether the change preserves backward compatibility.
Treat its findings as advisory input for the human reviewer, not a merge
blocker — use judgement about whether the flagged change is intentional.

<!-- jig:expect
valid: true
covers: [R-017, R-026]
-->
