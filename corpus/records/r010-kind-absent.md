---
id: HCR-0102
title: "Every binding must declare its kind"
scope: repo
regulates: architecture-fitness
summary: "The binding omits the kind field entirely, leaving the harness unable to dispatch it."
state: enforced
enforced_by:
  - run: "scripts/lint.sh"
---

Every enforcement binding must state which sensor kind executes it. Add the
missing `kind` field rather than leaving the harness to guess from context.

<!-- jig:expect
valid: false
covers: [R-010]
diagnostics:
  - rule: R-010
    at: /enforced_by/0
-->
