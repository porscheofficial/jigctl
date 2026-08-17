# Concepts

This document is the model's rationale — the *why* behind jigctl's design. The schema (`schema/hcr.schema.json`) and the corpus (`corpus/`) are normative: they say what is valid. This file explains why they say it, and never restates what they already say.

## What an HCR is, and why guidance and enforcement live together

A Harness Constraint Record (HCR) answers two questions in one artifact: what should an agent do here, and how do we know it did it? Most teams split those across two artifacts — a style guide or wiki page for guidance, a linter config or CI job for enforcement — and the split is exactly where they drift apart: the prose says one thing, the check says another, and nobody notices until a review catches the gap. An HCR keeps guidance (the summary and body an agent reads before writing code) and enforcement (the binding that verifies the result) in the same versioned record, so the two cannot silently diverge from each other.

Guidance without enforcement is a suggestion nobody follows. Enforcement without guidance is a check nobody understands until it fails and someone has to reverse-engineer intent from a broken pipeline. Neither half is useful alone, which is the entire reason they are one artifact instead of two.

## Three orthogonal axes

A binding's behavior is fixed by three independent questions, each living in its own field so that answering one never constrains the answer to another. `state` answers *is this rule active right now?* `severity` answers *does a violation fail the run?* `cadence` answers *when does this run?* These are different questions with different owners: whether a rule is active yet is a rollout decision, whether a violation blocks is a risk decision, and when a check runs is a scheduling decision.

The three stay orthogonal because collapsing any two of them makes some real situation inexpressible. A rule mid-rollout needs to run in CI as a non-blocking signal while the record itself is still provisional — if activation implied severity, or severity implied cadence, that shape could not be written down at all. Keeping the axes separate means a rule's lifecycle, its risk, and its schedule can each change independently without touching the other two.

## Why the six binding kinds are a closed set

`command`, `config-assert`, `grep`, `external`, `agent-review`, and `inferential` are closed — not an open enum a team can extend by adding a seventh. Every mechanism jigctl can express for checking a constraint reduces to one of these six: run something and inspect the result, assert a value inside a config file, require or forbid a text pattern, defer to an external tool's own verdict, ask a model to judge something no parser can, or name a constraint no tool can check yet.

Adding a seventh kind was rejected on the grounds that it would either be a rename of one of the six under new branding, or a checking mechanism this milestone has no diagnostic story for — no defined shape for what a validator reports when it fails. Closing the set means anything that consumes a binding — a validator today, a runner, a future dashboard — can exhaustively handle every kind and never meet one it doesn't recognize.

## Why scope resolves as repo ∪ service

The rules that actually apply to a given service are **repo ∪ service** — repo union service: every repo-wide record, plus every record scoped to that specific service. Union, not override and not intersection. A repo-wide rule ("every package lints clean") should not disappear because a service defines rules of its own, and a service-specific rule should not need to be duplicated at the repo level just to take effect.

Override semantics would let a service silently opt out of a repo-wide constraint just by defining a same-shaped local one. Intersection would mean that adding any service-level rule could shrink what's enforced elsewhere. Union is the one resolution where adding a record can only ever add a constraint and never remove one — a safety property neither of the other two options has.

## Why the corpus is the specification

There is deliberately no prose specification of what makes an HCR valid. The corpus is that specification: every valid shape is a fixture that must pass, every invalid shape is a fixture that must fail with exactly one diagnostic pointing at the one thing wrong with it. A prose spec is read once and drifts from the schema the moment someone edits one without the other — nobody notices until an agent trusts a sentence the schema no longer honors.

A corpus runs on every change instead of being read once: `corpus/RULES.md` records which rule each fixture is proving, but it is the fixtures, executed against the schema, that fail loudly the instant they disagree with it. The fixtures are the spec. This document and `corpus/RULES.md` are commentary on that spec, never a replacement for running it.

## Why a record's filename carries its id

A record file is named `HCR-NNNN-<slug>.md` — the id first, a human-readable slug after. Both halves earn their place. Without the id, the only way to answer "where does `HCR-0042` live?" is to read the frontmatter of every record in the tree, which is exactly the lookup a diagnostic, a `supersedes` chain, or an agent following a reference needs to do constantly. Without the slug, a directory listing is a wall of numbers that tells a human nothing about what the rules cover.

The alternative — topic-only names, with the id living solely inside the file — was rejected because it makes the id findable only by search, and a records directory grows monotonically: records are deprecated but never deleted, so the directory only ever gets longer and harder to navigate. A convention that reads fine at ten records has to still work at three hundred.

Note that the id in the filename is a navigation aid, never the source of truth. The `id` field in the frontmatter is authoritative; the filename is required to agree with it. That is a rule a tool can check, and a validator does — which is why it is a rule rather than a style preference. It does not apply to this repo's own negative test fixtures under `corpus/records/`, which are named after the rule they falsify: several of them deliberately carry a malformed id or no id at all, so there is nothing well-formed to name them after.

## What `regulates` means

Six values classify *why* a constraint exists, because the reason it exists determines who reviews a violation, how urgently it must be fixed, and whether it can ever be waived. Each value answers a single question that only it answers affirmatively:

`maintainability` — does violating it make a *future change* more expensive while breaking nothing today? Function length, dead code, and duplication qualify: nothing breaks the moment the rule is violated, but the next person to touch that code pays a tax for it.

`architecture-fitness` — is the violation *invisible in any single file*, detectable only in the relationship between two or more units? Layering direction, circular imports, and coupling qualify: reading any one file in isolation shows nothing wrong.

`behaviour` — does violating it break *someone outside* who changed nothing? API backward compatibility, semver, and schema evolution qualify: a caller who touched none of your code breaks anyway.

`reliability` — does violating it raise the *probability of a production incident*? Coverage floors, reversible migrations, and instrumentation qualify: nothing is broken yet, but the odds of an incident just went up.

`security` — does violating it create something an *attacker* can exploit? Secret handling, vulnerable dependencies, base images, and IAM qualify: the violation is a foothold for someone hostile, not an accident waiting to happen.

`compliance` — does the requirement *originate outside engineering*, carrying legal or contractual exposure? Licence allowlists, data residency, and audit retention qualify: the obligation exists because of a contract or a regulator, not an engineering judgment call.

### Boundary rules for pairs that collide

A fuzzy enum produces inconsistent tagging across teams, and inconsistent tagging destroys the one thing this field exists to produce: a coverage metric people can trust. Three pairs look alike enough that they need a fixed side, decided once here rather than re-litigated per pull request.

Function length is `maintainability`, but circular imports are `architecture-fitness` — one shows up in a single file's diff, the other only in the shape of the dependency graph.

API compatibility is `behaviour`, but reversible migrations are `reliability` — one breaks an external caller directly and immediately, the other only raises the odds of an incident if something else also goes wrong.

Vulnerable dependencies are `security`, but licence allowlists are `compliance` — one is exploitable by an attacker, the other is an obligation that exists whether or not anyone ever attacks anything.

## Rejected alternatives

A hand-curated TypeScript registry was rejected because it cannot reach Go, Python, or JVM teams — an HCR must be readable and checkable from whatever language a monorepo happens to contain, not only the language the tool itself was written in.

A separate `sensors.yaml` enumerating every check was rejected instead of trusting the records and the corpus as the single source of truth: a second hand-maintained registry drifts from what it describes, and nobody notices the drift until it is load-bearing.

Tombstone-by-deletion for deprecated records was rejected in favor of the `state: deprecated` value and the `supersedes` field, rather than hand-rolling deletion tracking on top of what version control already does, worse, for this exact purpose.

A maintained central index of active records was rejected: an index like that should be a generated view computed from the records themselves, not a hand-kept artifact a second author has to remember to update every time an underlying record changes.

Using `config-assert` to validate a DSL build script was rejected because that kind only works against data files it can path into and compare a value from — a build script is executable logic, not a value at a path, so this would misuse a kind for something it was never built to check.

Dispatching on binding kind the way a bare exclusive-choice enum would was rejected in favor of per-kind conditional branching: an exclusive-choice failure names no matching branch, so a validator can only report that none of six alternatives matched, which leaves tier-2 diagnostics unattributable — invalid, but not invalid *as* any particular kind.

Companion to the schema, never a duplicate: if a question here is answered by reading `schema/hcr.schema.json` or `corpus/RULES.md`, that answer belongs there, not in a second copy here.
