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
- Real records in `.hcr/` take ids in the `04xx` band; corpus fixtures
  take ids outside it. This keeps a search for a real record's id from
  landing in a deliberately broken fixture. See ADR-0002 for reasoning.
- Run `mise run check` before proposing any change
  and again before merging.

## Constraint Records

This repo is jigctl's first consumer: `.hcr/` is our own harness and the
only place the tool runs against rules somebody actually has to live with.
Keeping it current is part of the change, not follow-up work.

- Read the records frontmatter-first. Every `.hcr/*.md` opens with a YAML
  block carrying `title`, `scope`, `regulates`, `summary` and `state` —
  that block is the index. Scan every record's frontmatter, then open only
  the bodies that bear on the change in front of you; `summary` is capped
  at 25 words so the scan stays cheap. Where a record binds a `command`,
  its `run` is what decides the rule — run it rather than reasoning about
  whether the code complies.
- A change that introduces a constraint carries its record in the same
  commit. If a contributor could violate the new rule without noticing, it
  needs a record — a rule that lives only in a review comment, or only in
  this file, is enforced by nobody.
- A change that invalidates a record updates that record in the same
  commit. Renaming a task, moving a package or dropping a check moves the
  binding target with it, and a record pointing at something that no longer
  exists is worse than no record at all.
- Not every preference is a record. Two questions decide it: does it get
  violated in *this* repo, and can you name the concrete thing a check
  would bind to. Whatever fails either is a style note and belongs in this
  file instead.
- A new record is `state: draft`. Draft is how a rule's impact is measured
  before it gates anyone; promoting it to `warn` or `enforced` is a
  separate, deliberate change.
- Author records with the `hcr-author` skill rather than copying an
  existing one by hand — it owns the field limits, the `regulates`
  discriminators and the id and filename rules. When a record and the
  tooling disagree, fix the record: never the check it binds to. The skill
  is vendored here at `.agents/skills/hcr-author/`, and the canonical text
  of this section is its `references/agents-section.md`.

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
