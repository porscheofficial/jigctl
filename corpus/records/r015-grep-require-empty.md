---
id: HCR-0153
title: "An empty require list is not a real assertion"
scope: repo
regulates: security
summary: "The grep binding declares a require list with zero entries, which checks nothing."
state: enforced
enforced_by:
  - kind: grep
    file: "**/*.go"
    require: []
---

An empty `require` list is indistinguishable from asserting nothing at all.
Add at least one pattern the file must contain, or remove the key entirely.

<!-- jig:expect
valid: false
covers: [R-015]
diagnostics:
  - rule: R-015
    at: /enforced_by/0/require
-->
