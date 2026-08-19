# Record format

Everything needed to author a schema-valid record. Read this at Step 3.

## File

A record is a markdown file with YAML frontmatter, living in a `.hcr/`
directory, named `HCR-NNNN-<slug>.md` where `NNNN` matches its own `id` field
and `<slug>` matches `[a-z0-9]+(-[a-z0-9]+)*`. The frontmatter is authoritative;
the filename exists so a record can be found by id without opening it.

## Root fields

| Field | Required | Value |
|---|---|---|
| `id` | yes | `HCR-NNNN`, four digits, unique across the tree |
| `title` | yes | one line, reads as a violation message |
| `scope` | yes | `repo` or `service` — must match the record's location |
| `regulates` | yes | one of the six below |
| `summary` | yes | one sentence, ≤25 words |
| `state` | yes | `draft`, `warn`, `enforced`, `deprecated` |
| `enforced_by` | yes | array, at least one binding |
| `rationale` | no | external reference id, e.g. `ADR-0007` |
| `supersedes` | no | `HCR-NNNN` of a record this replaces; the target must exist |
| `exceptions` | no | array of `{scope, reason, until?}` |

No other root keys are permitted. `severity` and `cadence` are **per-binding**
fields and belong inside `enforced_by` items, never at the root.

Everything below the frontmatter is the body: guidance for whoever hits the
rule, plus the reasoning behind the classification. The argument itself always
lives here. The `rationale` field above does not hold prose and does not
replace it — it points at a decision recorded outside the harness.

## Three orthogonal axes

Never collapse these into each other. Each answers a different question.

- **`state`** — is this rule active? `draft` is not yet active. `warn` reports
  without blocking. `enforced` gates. `deprecated` is retained for history and
  never deleted, so superseding records keep a resolvable target.
- **`severity`** — does a violation fail the run? `blocking` or `advisory`.
- **`cadence`** — when does the check run? `on-change`, `ci`, `scheduled`,
  `production`.

A `draft` record with a `blocking` binding is coherent and common: the rule is
not active yet, but when it activates it will block.

## Choosing `regulates`

Ask which statement is true of a violation:

| Value | A violation… |
|---|---|
| `maintainability` | makes a future change more expensive while breaking nothing today |
| `architecture-fitness` | is invisible in any single file; only the relationship between units is wrong |
| `behaviour` | breaks someone outside who changed nothing |
| `reliability` | raises the probability of a production incident |
| `security` | creates something an attacker can exploit |
| `compliance` | violates an obligation originating outside engineering |

The pairs that actually collide, resolved:

- function length is `maintainability`, but circular imports are
  `architecture-fitness`
- API compatibility is `behaviour`, but reversible migrations are `reliability`
- vulnerable dependencies are `security`, but licence allowlists are
  `compliance`

## Binding kinds

Six, and the set is closed. Each binding declares `kind` plus only the fields
that kind owns — a field belonging to another kind is a schema error, not a
harmless extra.

| Kind | Required | Also allowed | Default severity | Default cadence |
|---|---|---|---|---|
| `command` | `run` | `ref`, `select`, `timeout_secs`, `pattern` | `blocking` | `[on-change, ci]` |
| `config-assert` | `file`, `path`, `op` | `value` | `blocking` | `[on-change, ci]` |
| `grep` | `file` + at least one of `require`/`forbid` | the other of the two | `blocking` | `[on-change, ci]` |
| `external` | `tool`, `docs` | — | `blocking` | `[on-change, ci]` |
| `agent-review` | `prompt` | `grounding`, `model`, `runs` | `advisory` (fixed) | `[scheduled]` |
| `inferential` | — | — | `advisory` | none |

Notes that bite:

- **`command`** — `ref` names a shared check. Every binding across the tree that
  uses the same `ref` must have a byte-identical `run`.
- **`config-assert`** — `op` is `equals`, `gte`, `lte`, `matches` or `absent`.
  `value` is required for every operator except `absent`, and forbidden with it.
- **`grep`** — `require` and `forbid` are arrays; a present one may not be empty.
- **`external`** — for a rule enforced by something outside the repo. `docs`
  must point at where that enforcement is documented.
- **`agent-review`** — severity is fixed at `advisory`. Non-deterministic review
  may never gate a merge, so `blocking` is a schema error rather than a runtime
  downgrade.
- **`inferential`** — owns no fields at all. It records a rule that is real but
  not mechanically checkable yet. This is the honest choice when a rule matters
  and no check exists; inventing a `command` that does not run is not.

## Constraints the validator enforces

These fire regardless of `state`. A `draft` record is validated exactly as
strictly as an `enforced` one, so a draft binding must still resolve.

| Code | Cause | Fix |
|---|---|---|
| `R-101` | id used by another record in the tree | pick the next free id |
| `R-102` | `supersedes` names a missing record, or itself | point at a record that exists |
| `R-103` | two bindings share a `ref` but their `run` strings differ | make them byte-identical |
| `R-104` | a path-shaped `run` does not exist on disk | create the script, or bind to a task runner |
| `R-109` | `scope` disagrees with the directory the record is in | move the file, or change the scope |
| `R-110` | `external.docs` path or `#anchor` does not resolve | fix the path, or add the heading |
| `R-112` | filename does not start with the record's own `id` | rename the file |

**R-104 in detail**, because it is the one that surprises people. Only the first
whitespace-separated token of `run` is inspected. If that token contains a `/`,
it is resolved against the tree root and must exist. So:

- `mise run lint` — skipped, first token has no `/`
- `uv run --script tools/check.py` — skipped, first token is `uv`
- `tools/check.py` — **checked**, the file must exist

This is what makes it safe to author a draft record binding to a task-runner
target that has not been written yet, while a bare script path must be real.

**R-110 anchors** use jigctl's own slug rule, not GitHub's: strip leading `#`,
trim, lowercase, collapse whitespace runs to a single `-`, then delete every
character outside `[a-z0-9-]`. Duplicate headings are not disambiguated, so
GitHub's `#api-1` does not resolve here.

## Exceptions

`scope` and `reason` are both mandatory: an exemption without a stated reason is
indistinguishable from an oversight. Optional `until` is `YYYY-MM-DD`; after
that date the rule fires again, so a temporary waiver cannot quietly become
permanent. Add one only when you know of a real case that must be exempt — a
speculative exception is a hole.

## Worked example

```markdown
---
id: HCR-0042
title: Every Go source file passes the configured linters
scope: repo
regulates: maintainability
summary: "All Go source passes gofumpt, golangci-lint and nilaway with the checked-in configuration."
state: draft
enforced_by:
  - kind: command
    run: "mise run lint"
---
Run `mise run lint` before opening a change. It runs gofumpt for formatting,
golangci-lint for the rule set in `.golangci.yml`, and nilaway for nil-pointer
analysis. Fix findings at the source rather than adding a suppression comment;
if a rule is genuinely wrong for this codebase, change the configuration so the
decision is visible in review.

Classified `maintainability` rather than `reliability`: a lint finding costs
future readers time and makes the next change more expensive, but nothing that
ships today is broken by it. A nil dereference that reaches production would be
`reliability` — that is a property of the runtime, not of the lint gate.
```

The body has two parts, and both are load-bearing: imperative guidance for
whoever hits the rule, then a paragraph defending the `regulates` choice against
its nearest neighbour. Match whatever equivalent convention the target repo's
existing records already use.
