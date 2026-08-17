---
id: HCR-0152
title: "Grep bindings must assert at least one pattern"
scope: repo
regulates: security
summary: "The grep binding names a file to scan but asserts neither a required nor a forbidden pattern."
state: enforced
enforced_by:
  - kind: grep
    file: "**/*.go"
---

A `grep` binding with no `require` and no `forbid` pattern checks nothing and
is a silent no-op. Add at least one pattern under `require` or `forbid`.

<!-- jig:expect
valid: false
covers: [R-015]
diagnostics:
  - rule: R-015
    at: /enforced_by/0
-->
