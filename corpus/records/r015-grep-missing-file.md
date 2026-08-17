---
id: HCR-0151
title: "Grep bindings must name a target file"
scope: repo
regulates: security
summary: "The grep binding omits the file field, leaving no target for the pattern search."
state: enforced
enforced_by:
  - kind: grep
    require: ["TODO(security-review)"]
---

A `grep` binding scans a specific set of files for required or forbidden
patterns. Add the missing `file` glob naming what to scan.

<!-- jig:expect
valid: false
covers: [R-015]
diagnostics:
  - rule: R-015
    at: /enforced_by/0
-->
