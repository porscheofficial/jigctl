---
id: HCR-0405
title: CI must invoke mise run check, not duplicate its tasks
scope: repo
regulates: maintainability
summary: "CI must invoke `mise run check` rather than inlining its own copies of the lint, corpus and metaschema commands. Duplicated task logic in mise.toml and the workflow file diverges the moment either side is edited alone."
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

Classified `maintainability`: nothing breaks the day the duplication is
introduced, since CI still runs whatever was copied in. The cost lands
on the next change — a new flag, a renamed step, a reordering in a mise
task has no reason to also touch the CI file, so CI keeps validating
against a stale copy of what a contributor runs locally.
