---
id: HCR-0006
title: Manual sign-off for cross-service contract changes
scope: service
regulates: architecture-fitness
summary: A cross-service contract change needs a documented human sign-off before merge; nothing here executes automatically.
state: draft
enforced_by:
  - kind: inferential
---
If your change alters a contract another service depends on, name an
accountable reviewer and record their sign-off in the pull request
description before merging. Nothing here runs a check — this is a
judgement call that a human owns, not something a command can decide.

<!-- jig:expect
valid: true
covers: [R-010]
deferred:
  - rule: R-106
    reason: requires-cli
-->
