---
id: HCR-0302
title: Root Node.js engine version pin
scope: repo
regulates: reliability
summary: The root package manifest must pin an explicit Node.js engine range so every service builds against the same runtime.
state: enforced
enforced_by:
  - kind: config-assert
    severity: blocking
    cadence: [on-change, ci]
    file: package.json
    path: /engines/node
    op: equals
    value: "20.x"
---

Keep `engines.node` in the root `package.json` set to `20.x`. If a service
needs a newer runtime, open a repo-wide discussion first rather than
drifting one service's toolchain out from under this pin.

