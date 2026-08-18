---
id: HCR-0405
title: CI must invoke mise run check, not duplicate its tasks
scope: repo
regulates: maintainability
summary: "If a CI workflow inlines its own copies of the lint, corpus, and metaschema commands instead of delegating to mise run check, the same task logic then exists in two places — mise.toml and the CI workflow file. Nothing breaks the day that duplication is introduced; CI still runs whatever was copied in. The cost lands on the next change: an edit to a mise task (a new flag, a renamed step, a reordering) has no reason to also touch the CI file, so the two definitions silently diverge and CI keeps validating against a stale copy of what a contributor runs locally. That is docs/concepts.md's maintainability question exactly — breaks nothing today, taxes the next change."
state: enforced
enforced_by:
  - kind: grep
    file: ".github/workflows/ci.yml"
    require: ["mise run check"]
---
When you add .github/workflows/ci.yml, its steps must call
`mise run check` rather than reimplementing metaschema, corpus, shape,
or lint invocations inline. If CI ever needs to run a subset, add or
adjust a mise task and call that task by name — do not fork the command
out into the workflow file.
