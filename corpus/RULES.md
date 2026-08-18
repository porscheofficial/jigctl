# RULES.md — the rule register

This register is normative for which rules exist. Completeness means every rule id is cited by at least one fixture: tier-1 rule ids via `covers`, deferred tier-2 rule ids via `deferred`, and enforced tier-2 rule ids via at least one `diagnostics` entry in a tree fixture's `expect.yaml` (see below). A rule with N independently falsifiable clauses gets N fixtures.

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

| id | rule | status | reason | fixture home |
|---|---|---|---|---|
| R-101 | `id` is unique across the tree, not scoped to a single service's effective set — `supersedes` and `// Implements [[HCR-NNNN]]` carry no service qualifier, so two sibling services defining the same id would break both references while no single effective set could ever observe the collision | enforced (M1) | requires-cross-file-resolution | multi-service, a repo-level record |
| R-102 | every `supersedes` target exists and is not dangling | enforced (M1) | requires-cross-file-resolution | the valid fixture carrying `supersedes` |
| R-103 | two records sharing a `ref` carry byte-identical `run` after YAML decoding (D5) | enforced (M1) | requires-cross-file-resolution | **both** multi-service service records — they carry the same `ref`, since a single record cannot express a two-record rule |
| R-104 | a path-shaped `run` resolves on disk (`handover:157`) | enforced (M1) | requires-filesystem | the valid `command` fixture |
| R-105 | `state: warn` downgrades a binding's severity to advisory at runtime (D7) | deferred | requires-cli | the valid `warn` + `severity: blocking` fixture |
| R-106 | severity/cadence defaults are injected per kind tier (D8) | deferred | requires-cli | the valid fixture that omits both |
| R-107 | an `exceptions[].until` in the past re-fires the constraint | deferred | requires-cli | the valid `exceptions` fixture, using a past date |
| R-108 | a service's effective set = repo records ∪ service records | enforced (M1) | requires-cross-file-resolution | multi-service, a service-level record |
| R-109 | a record's `scope` agrees with its directory location | enforced (M1) | requires-cross-file-resolution | multi-service, a service-level record |
| R-110 | `external.docs` anchors resolve (`handover:175`) | enforced (M1) | requires-filesystem | the valid `external` fixture |
| R-111 | a `rationale` reference resolves to an artifact that exists | deferred | requires-cli | the valid fixture carrying `rationale` |
| R-112 | a record's filename begins with its own `id` (`HCR-NNNN-<slug>.md`) | enforced (M1) | requires-filesystem | multi-service, a repo-level record — every record there is named this way |

### Enforced tier-2 rule algorithms

Every rule below runs only over records that have already cleared the JSON Schema layer: a record is decoded a second time into a typed record — and so becomes eligible for META evaluation at all — only if the schema layer produced zero diagnostics for it. A record that fails the schema layer is excluded from META evaluation entirely. R-102's target-lookup index is the one deliberate exception to this gating, described in its own paragraph below.

**R-101.** `id` uniqueness is checked across every record discovered anywhere in the tree — repo-scoped and every service-scoped directory together — never scoped to a single service's effective set. The algorithm collects the `id` of every record that passed the schema layer, groups records by `id`, and reports one diagnostic per record beyond the first in any group larger than one. Tree-wide scope is deliberate: `supersedes` and `// Implements [[HCR-NNNN]]` source references carry no service qualifier, so two sibling services independently choosing the same id would corrupt both kinds of reference while no single service's effective set would ever contain both colliding records to notice the clash.

**R-102.** Every `supersedes: HCR-NNNN` value must name the `id` of some record that exists in the tree. The target-lookup index is built by scanning every discovered record for an `id` matching `^HCR-[0-9]{4}$` — deliberately including records that fail the schema layer, the one exception to the gating rule above — so that whether a `supersedes` target is judged to exist never depends on the order in which unrelated records get their own schema defects repaired. Only the record making the `supersedes` claim must itself have passed the schema layer to be checked at all; the target side of the lookup is schema-agnostic.

**R-103.** Within an effective set, every pair of `enforced_by` bindings sharing the same `ref` must decode to identical `run` strings. The comparison happens once both values have been YAML-decoded: a plain `Unmarshal` discards quoting style, so `run: "make lint"` and `run: make lint` decode to the identical Go string and are equal, even though their source quoting differs. When a `ref` group disagrees, the algorithm emits one diagnostic per participating binding, not one for the group as a whole, so every offending binding is individually flagged.

**R-104.** A `command` binding's `run` string is tested for a path-shaped first token: the algorithm splits `run` on whitespace, takes only the *first* token, and resolves it against the filesystem only if that token contains a `/`. `run: tools/schema-shape.py` is checked, because its first (and only) token contains a `/`. `run: mise run lint` is skipped, because its first token, `mise`, contains no `/`. `run: uv run --script tools/schema-shape.py` is also skipped, even though a later token is a real path, because only the first token is ever inspected and `uv` contains no `/`.

**R-108.** A service's effective set is the union of every repo-scoped record and that service's own service-scoped records. This rule is a resolver rather than a diagnostic-producing check: a set union cannot fail, so R-108 never emits a diagnostic and has no negative fixture — every other META rule that reasons about "the effective set" (R-101 excepted, which is deliberately tree-wide) depends on this resolution having already happened.

**R-109.** A record's `scope` field must agree with where it was discovered: a record found under the tree root's `.hcr/` must declare `scope: repo`, and a record found under a service's `.hcr/` must declare `scope: service`. The algorithm compares the declared `scope` against the directory tier the record was discovered in and reports a diagnostic on any mismatch.

**R-110.** Every `external.docs` value must resolve. The algorithm treats `docs` as a filesystem path with an optional `#anchor` fragment: it checks that the path component exists on disk and, if an anchor is present, that the anchor exists within that file. No network I/O is performed — a `docs` value naming a URL is out of scope for this check, deliberately, so the rule can never fail because of network availability.

**R-112.** A record's filename must begin with its own `id`: `HCR-0042-some-slug.md` is a valid filename for a record whose frontmatter `id` is `HCR-0042`; any other filename prefix is a diagnostic.

## Fixture corpora and the `expect.yaml` format

Two independent corpora exist, deliberately validated by two different entry points — this is why `corpus/records/` never carries a META (tier-2) rule.

`corpus/records/*.md` (58 files) is the **record-shape corpus**: each file is one isolated HCR record, validated by a single-file, schema-only entry point that has no notion of a surrounding tree and therefore cannot evaluate any META rule. Its expectations live in the per-file `<!-- jig:expect -->` block already used above (`valid`, `covers`, `diagnostics`, `deferred`).

Any directory under `corpus/fixtures/` that contains a `jig.toml` is a **tree fixture** — for example `corpus/fixtures/multi-service/` — a small filesystem tree of one or more services, validated by a filesystem-aware entry point that runs the schema layer over every record in the tree *and* all eight META rules across the resolved effective sets. A tree fixture's expectations live in one `expect.yaml` file at the fixture root, not in per-file comments, because a META diagnostic is a property of the tree as a whole rather than of any single record.

`expect.yaml` holds a single `diagnostics:` list. Each entry is `{file, at, rule}`: `file` is the offending record's path relative to the fixture root (for example `services/api/.hcr/HCR-0303-api-contract-check.md`), `at` is the same JSON-pointer-shaped location tier-1 fixtures already use, and `rule` names the violated tier-2 rule id. The list is matched against the actual run as an exact sorted set: every expected diagnostic must appear, and no unexpected diagnostic may appear alongside it. `expect.yaml` deliberately has no `covers`, `valid`, or `deferred` key — those are `corpus/records/`'s per-file vocabulary, and none of them means anything for a tree-wide result.

## Allowed properties by binding kind

| kind | owns | prohibits |
|---|---|---|
| command | run ref select timeout_secs pattern | 12 |
| config-assert | file path op value | 13 |
| grep | file require forbid | 14 |
| external | tool docs | 15 |
| agent-review | prompt grounding model runs | 13 |
| inferential | — | 17 |
