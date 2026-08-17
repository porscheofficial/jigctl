---
id: HCR-2011
title: Require deprecation notices before removing a public field
scope: repo
regulates: behaviour
summary: A public API field must carry a deprecation notice for one release before removal.
state: enforced
---

Mark the field as deprecated in the schema and changelog at least one release before deleting it.
Give consumers a migration note describing the replacement field or behaviour.
Do not remove a field in the same release that first marks it deprecated.

<!-- jig:expect
valid: false
covers: [R-009]
diagnostics:
  - rule: R-009
    at: ""
-->
