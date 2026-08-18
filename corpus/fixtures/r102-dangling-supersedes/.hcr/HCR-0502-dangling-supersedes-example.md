---
id: HCR-0502
title: Repository lint gate stays green on every commit
scope: repo
regulates: maintainability
summary: >-
  The repository's lint gate must pass before merge; this record replaces an
  older lint policy that no longer exists in this tree.
state: enforced
supersedes: HCR-0599
enforced_by:
  - kind: inferential
---

`supersedes: HCR-0599` names an id that does not exist anywhere under this
fixture's tree. That dangling reference is the fixture's single violation:
R-102 requires every `supersedes` target to exist in the tree-global identity
index.
