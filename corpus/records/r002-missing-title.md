---
id: HCR-2002
scope: repo
regulates: architecture-fitness
summary: Modules in the domain layer must not import from the infrastructure layer directly.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-layer-imports.sh
---

Keep dependencies pointing inward: domain code must not know about infrastructure details.
Introduce an interface in the domain layer and implement it in infrastructure instead.
If you find yourself importing a driver or client library from domain code, stop and refactor.

<!-- jig:expect
valid: false
covers: [R-002]
diagnostics:
  - rule: R-002
    at: ""
-->
