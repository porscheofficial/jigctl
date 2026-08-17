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
- `tools/corpus-runner.py` is test infrastructure, not a product. It must
  NEVER grow cross-file rules and must NEVER accept a path argument.
- No new binding `kind` and no new `cadence` value may be added without
  updating this file and the schema together, in the same change.
- A record file is named `HCR-NNNN-<slug>.md`, matching its own `id` field.
  This governs real records (see `corpus/fixtures/multi-service/`), NOT the
  negative fixtures in `corpus/records/`, which are named `r<NNN>-<slug>.md`
  after the rule they falsify — several carry a malformed id or none at all.
- Run `mise run check` before proposing any change
  and again before merging.

## Structure

- Single, flat `AGENTS.md` at repo root. No nested `AGENTS.md` files exist
  or should be added at this milestone (M0) — there are no subdirectories
  that warrant per-directory overrides yet.
