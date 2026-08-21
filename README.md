# jigctl

A constraint harness for polyglot monorepos — repo-wide and per-service.

An **HCR** (Harness Constraint Record) is a single, versioned, machine-checkable
rule about your codebase. `jigctl` is the reference implementation CLI that reads
and validates HCRs stored under a repo's `.hcr/` directory, configured by `jig.toml`.

Here is a complete HCR:

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

## Install

Build the CLI from source inside a repository checkout. There are no published releases yet.

```bash
go build -o jigctl ./cmd/jigctl
```

## Quick Start

Validate your repo's records for well-formedness:

```bash
./jigctl validate .
```

## Machine Contract

`jigctl run` provides a JSON contract via `--format=json` for tooling integration.

- **Schema**: Validates against [`schema/run-output-v1.schema.json`](schema/run-output-v1.schema.json).
- **Versioning**: The `schema_version` field guarantees compatibility. Additive changes (new fields) are non-breaking; consumers should ignore unknown fields.
- **Channel Contract**: A completed run always emits valid JSON on `stdout` (even with 0 records). An invocation failure (e.g. bad format) or operational crash emits an empty `stdout`, writes the error to `stderr`, and exits with code `2`.

**Example:**

```bash
./jigctl run . --format=json
```

```json
{
  "schema_version": 1,
  "command": "run",
  "root": "/path/to/repo",
  "exit_code": 0,
  "diagnostics": [],
  "summary": {
    "records": 1,
    "bindings": 1,
    "bindings_by_projection": {
      "pass": 1,
      "violation": 0,
      "expected-unchecked": 0,
      "blocked-unchecked": 0,
      "operational": 0,
      "invalid": 0
    },
    "unwaived_findings": 0,
    "files_with_unwaived_findings": 0
  },
  "records": []
}
```

## Status

This repo contains the HCR schema and the `jigctl` CLI.

Licensed under MIT.
