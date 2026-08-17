---
id: HCR-0111
title: "Binding items must not carry unrecognised keys"
scope: repo
regulates: maintainability
summary: "The binding carries an extra key that belongs to no recognised kind."
state: enforced
enforced_by:
  - kind: command
    run: "scripts/check-style.sh"
    owner: "platform-guild"
---

Keep each binding limited to the fields its kind recognises. An extra key is
usually a typo or a leftover from a template and should be removed rather than
carried forward.

<!-- jig:expect
valid: false
covers: [R-011]
diagnostics:
  - rule: R-011
    at: /enforced_by/0
-->
