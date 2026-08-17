# jigctl

A constraint harness for polyglot monorepos — repo-wide and per-service.

An **HCR** (Harness Constraint Record) is a single, versioned, machine-checkable
rule about your codebase; `jigctl` is the reference implementation that reads
and enforces HCRs stored under a repo's `.hcr/` directory, configured by `jig.toml`.

Here's a complete HCR:

```yaml
---
id: HCR-0001
title: Static lint check must pass before merge
scope: repo
regulates: maintainability
summary: Every changed package must pass the linter with zero warnings before a pull request can merge.
state: enforced
enforced_by:
  - kind: command
    run: "make lint"
    timeout_secs: 120
    severity: blocking
    cadence: [on-change, ci]
---
Run `make lint` against any package you touch before opening a pull request.
Zero warnings are allowed — fix the code rather than suppressing the rule.
```

This repo currently contains the HCR schema (`schema/hcr.schema.json`) and a
fixture corpus of valid and invalid example records (`corpus/records/`).
There is no CLI yet.

Licensed under MIT.
