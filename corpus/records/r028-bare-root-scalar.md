---
this is not a mapping, just a plain scalar string
---

This fixture deliberately breaks the frontmatter shape rather than any single field.
A schema without an explicit object-type check at the root would accept this silently.
The body still carries the required assertion block below.

<!-- jig:expect
valid: false
covers: [R-028]
diagnostics:
  - rule: R-028
    at: ""
-->
