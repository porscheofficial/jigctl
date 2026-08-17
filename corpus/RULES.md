# RULES.md — the rule register

This register is normative for which rules exist. Completeness means every rule id is cited by at least one fixture: tier-1 rule ids via `covers`, tier-2 rule ids via `deferred`. A rule with N independently falsifiable clauses gets N fixtures.

## Tier 1 — schema-enforceable today, asserted mechanically

| id | rule | fixtures | asserted `at` |
|---|---|---|---|
| R-001 | `id` matches `^HCR-[0-9]{4}$` | 1 | `/id` |
| R-002 | every required root field present — one fixture per omitted field: `id`, `title`, `scope`, `regulates`, `summary`, `state` | 6 | `""` (root object) |
| R-003 | an unknown frontmatter key is rejected | 1 | `""` (root object) |
| R-004 | `scope` ∈ {repo, service} | 1 | `/scope` |
| R-005 | `regulates` ∈ {maintainability, architecture-fitness, behaviour, reliability, security, compliance} | 1 | `/regulates` |
| R-006 | `state` ∈ {draft, warn, enforced, deprecated} | 1 | `/state` |
| R-007 | `severity` ∈ {blocking, advisory} when present (binding-level) | 1 | `/enforced_by/0/severity` |
| R-008 | every `cadence` item ∈ {on-change, ci, scheduled, production} (binding-level) | 1 | `/enforced_by/0/cadence/0` |
| R-009 | `enforced_by` present, `minItems: 1` — the schema expression of the falsification test at `handover:213` | 2 | absent → `""`; empty array → `/enforced_by` |
| R-010 | `enforced_by[].kind` present and ∈ the closed set of six | 2 | bogus value → `/enforced_by/0/kind`; absent → `/enforced_by/0` |
| R-011 | an unknown key inside an `enforced_by` item is rejected | 1 | `/enforced_by/0` |
| R-012 | `command` requires `run` | 1 | `/enforced_by/0` |
| R-013 | `config-assert` requires `file`, `path`, `op`; `op` ∈ {equals, gte, lte, matches, absent} | 4 | 3 × `/enforced_by/0`; bad enum → `/enforced_by/0/op` |
| R-014 | `config-assert` with `op: absent` forbids `value` | 1 | `/enforced_by/0` |
| R-015 | `grep` requires `file`, and at least one **non-empty** `require` / `forbid` (otherwise the binding is a legal no-op) | 3 | missing `file` → `/enforced_by/0`; neither key → `/enforced_by/0`; `require: []` → `/enforced_by/0/require` |
| R-016 | `external` requires `tool` and `docs` | 2 | both → `/enforced_by/0` |
| R-017 | `agent-review` requires `prompt` | 1 | `/enforced_by/0` |
| R-018 | `agent-review` forbids `severity: blocking` (D7) | 1 | `/enforced_by/0/severity` |
| R-019 | each kind rejects fields belonging to another kind — **one fixture per kind**, six in total (D2's prohibition mechanism; the full ~90-case matrix is covered structurally by F1, not by fixtures) | 6 | `/enforced_by/0` for each |
| R-020 | `exceptions[]` requires both `scope` and `reason` | 2 | both → `/exceptions/0` |
| R-021 | `exceptions[].until` deserialises as a string and matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` (D9) | 2 | a bare non-string YAML scalar, e.g. an unquoted integer (fails `type: string`) → `/exceptions/0/until`; `"2026/01/01"` (fails `pattern`) → `/exceptions/0/until` |
| R-022 | `supersedes` is a string matching `^HCR-[0-9]{4}$` | 1 | `/supersedes` |
| R-023 | `rationale` matches `^[A-Z][A-Z0-9]*-[A-Za-z0-9._/-]+$` | 1 | `/rationale` |
| R-025 | `timeout_secs` is an integer ≥ 1 | 1 | `/enforced_by/0/timeout_secs` |
| R-026 | `runs` is an integer ≥ 1 | 1 | `/enforced_by/0/runs` |
| R-027 | `config-assert` with `op != absent` requires `value` (D10) | 1 | `/enforced_by/0` |
| R-028 | `type: object` holds at all three levels — a non-object frontmatter root, `enforced_by` item, or `exceptions` item is rejected (D2) | 3 | `""` (root); `/enforced_by/0`; `/exceptions/0` |

**Total tier-1 invalid fixtures: 49 across 27 rules.**

**R-024 is retired and its id is not reused.** It governed `config-assert.ratchet`, which was removed before any release because ratcheting requires a generated baseline outside the authored tree and that state format is undecided. Ids are not renumbered on removal: recycling `R-024` would make any surviving reference to it resolve silently to a different rule.

## Tier 2 — expressible now, executable only once the CLI exists

| id | rule | reason | fixture home |
|---|---|---|---|
| R-101 | `id` is unique across the effective set | requires-cross-file-resolution | multi-service, a repo-level record |
| R-102 | every `supersedes` target exists and is not dangling | requires-cross-file-resolution | the valid fixture carrying `supersedes` |
| R-103 | two records sharing a `ref` carry byte-identical `run` (D5) | requires-cross-file-resolution | **both** multi-service service records — they carry the same `ref`, since a single record cannot express a two-record rule |
| R-104 | a path-shaped `run` resolves on disk (`handover:157`) | requires-filesystem | the valid `command` fixture |
| R-105 | `state: warn` downgrades a binding's severity to advisory at runtime (D7) | requires-cli | the valid `warn` + `severity: blocking` fixture |
| R-106 | severity/cadence defaults are injected per kind tier (D8) | requires-cli | the valid fixture that omits both |
| R-107 | an `exceptions[].until` in the past re-fires the constraint | requires-cli | the valid `exceptions` fixture, using a past date |
| R-108 | a service's effective set = repo records ∪ service records | requires-cross-file-resolution | multi-service, a service-level record |
| R-109 | a record's `scope` agrees with its directory location | requires-cross-file-resolution | multi-service, a service-level record |
| R-110 | `external.docs` anchors resolve (`handover:175`) | requires-filesystem | the valid `external` fixture |
| R-111 | a `rationale` reference resolves to an artifact that exists | requires-cli | the valid fixture carrying `rationale` |
| R-112 | a record's filename begins with its own `id` (`HCR-NNNN-<slug>.md`) | requires-filesystem | multi-service, a repo-level record — every record there is named this way |

## Allowed properties by binding kind

| kind | owns | prohibits |
|---|---|---|
| command | run ref select timeout_secs pattern | 12 |
| config-assert | file path op value | 13 |
| grep | file require forbid | 14 |
| external | tool docs | 15 |
| agent-review | prompt grounding model runs | 13 |
| inferential | — | 17 |
