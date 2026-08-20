---
id: HCR-0413
title: jigctl must validate its own records with execution authorized
scope: repo
regulates: reliability
summary: "The mise dogfood task must run jigctl over this repository with --allow-exec, so its own command bindings execute instead of being skipped."
state: draft
rationale: ADR-0012
enforced_by:
  - kind: config-assert
    file: "mise.toml"
    path: "/tasks/dogfood/run"
    op: equals
    value: "go run ./cmd/jigctl run . --allow-exec"
---
Keep `--allow-exec` on the `dogfood` task, and keep `dogfood` in the
`check` task's dependencies. Without the flag the runner still reports an
outcome for every command binding under .hcr/, but that outcome is that
the command was never executed — the tree stays green while nothing ran.
If you want a faster local pass, add a separate task; do not strip the
flag off this one.

The assertion pins the whole command rather than looking for the flag
alone, because retargeting the run at a fixture tree would satisfy any
check that only asked whether `--allow-exec` was present, and this
repository would quietly stop being subject to its own records.

Classified `reliability` rather than `architecture-fitness`: the failure
mode is a false green. Bindings that never execute report success to CI,
so a violation this repository already records is merged as though it had
been checked.
