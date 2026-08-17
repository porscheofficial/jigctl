---
id: HCR-0003
title: No debugger breakpoints left in service code
scope: service
regulates: maintainability
summary: Committed Python code must not contain interactive debugger breakpoints and must declare the future annotations import.
state: enforced
enforced_by:
  - kind: grep
    file: "src/**/*.py"
    require: ["from __future__ import annotations"]
    forbid: ["pdb.set_trace(", "breakpoint("]
    severity: blocking
    cadence: [on-change, ci]
---
Every new or modified Python module under `src/` must start with
`from __future__ import annotations`, and must never contain a
`pdb.set_trace()` or bare `breakpoint()` call. Use the project logger
instead of an interactive breakpoint, and add the missing future import
if the file lacks one.

<!-- jig:expect
valid: true
covers: [R-015]
-->
