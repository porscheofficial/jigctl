# AGENTS.md

This file is for agents and contributors working ON jigctl itself (this
repo), not for agents consuming HCRs in some other repo.

## Invariants

- The corpus (`corpus/`) is normative: fixtures ARE the spec. `corpus/RULES.md`
  governs fixture structure.
- Every invalid fixture isolates exactly ONE violation. Do not combine
  multiple violations in a single invalid fixture.
- Never weaken the schema (`schema/hcr.schema.json`) to make a fixture pass.
  If a fixture and the schema genuinely disagree, that is a defect to
  report, not something to hack around.
- Schema documentation goes in `description` (for consumers) and `$comment`
  (for maintainers) — never in comments, which strict JSON cannot carry.
  Cross-kind prohibitions stay literally `"<name>": false`: annotating one
  turns it into a schema that permits any value, silently removing the
  prohibition.
- The `jigctl` CLI is the product and accepts a path argument, replacing the
  contained Python test runner. However, `cmd/jigctl` is wiring only; all
  cross-file rules and validation logic must live in `internal/hcr`.
- No new binding `kind` and no new `cadence` value may be added without
  updating this file and the schema together, in the same change.
- A record file is named `HCR-NNNN-<slug>.md`, matching its own `id` field.
  This governs real records (see `corpus/fixtures/multi-service/`), NOT the
  negative fixtures in `corpus/records/`, which are named `r<NNN>-<slug>.md`
  after the rule they falsify — several carry a malformed id or none at all.
- Run `mise run check` before proposing any change
  and again before merging.

## Go Implementation

- Strict 250 pure-LOC per-file ceiling across the codebase.
- `cmd/jigctl/main.go` is ≤30 lines.
- `internal/hcr` is the only implementation of validation (two consumers: the CLI and `go test`).
- A diagnostic is data, not a Go `error` — this deliberately overrules the Go reference stack's `errors.Join` mandate for diagnostics; `errors.Join` and `%w` apply to operational errors only.
- Never edit a fixture's `valid`, `at` or `covers` values to make the tool pass — there is a SHA-256 backstop over the expectation blocks in `internal/hcr/testdata/expectations.sha256`, refreshed by `mise run expectations:freeze`.

## Structure

- Single, flat `AGENTS.md` at repo root. No nested `AGENTS.md` files exist
  or should be added at this milestone (M1) — there are no subdirectories
  that warrant per-directory overrides yet.
