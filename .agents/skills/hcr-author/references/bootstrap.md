# Bootstrap — first records in a repository

Follow this path when Step 1 found no `jig.toml`, or found one but resolved zero
records. The user has probably never seen an HCR. Explain the model, set the
tree up, and propose five records so self-evident that the user recognises their
own repo in them.

## 1. Explain the model, briefly

Say roughly this, adapted to what the repo actually is. Keep it short — a wall
of theory is how adoption dies.

> A Harness Constraint Record is one markdown file holding both a rule and the
> check that enforces it. Today those live apart: the reasoning is in
> `CONTRIBUTING.md` or an ADR, the enforcement is a line in a CI config, and
> nobody notices when they stop agreeing. Putting them in one versioned file
> means changing the rule and changing its check are the same edit.
>
> Records live in `.hcr/` — one at the repo root for rules that apply
> everywhere, and one per service for rules that apply to just that service.
> `jigctl validate` checks they stay well-formed and consistent.

Be accurate about what the installed `jigctl` does. If the version in use
validates records but does not yet execute their bindings, say so plainly. The
record is still the one place the rule and its check are written down together,
and the binding is machine-readable for when execution arrives.

## 2. Detect the tree shape

Decide whether this is a single project or a monorepo, from evidence:

- multiple directories under `services/`, `packages/`, `apps/` or `cmd/`, each
  with its own manifest (`go.mod`, `package.json`, `pyproject.toml`,
  `Cargo.toml`) → monorepo
- one manifest at the root → single project

Do not guess from directory names alone. A `packages/` directory with one
manifest at the root is a single project.

## 3. Create the tree configuration

At the repo root, write `jig.toml`:

```toml
# Single project — all records are repo-scoped.
service_globs = []
```

```toml
# Monorepo — each match that contains a .hcr/ directory contributes its records.
service_globs = ["services/*"]
```

Then create `.hcr/` at the root. Create a service's `.hcr/` only when you are
actually writing a service-scoped record into it — a glob match without a
`.hcr/` directory is simply skipped, so empty ones add nothing.

A check that a record binds by path belongs in `.hcr/checks/`, beside the
records that invoke it. Discovery globs `<root>/.hcr/*.md` and never descends,
so the subdirectory is invisible to jigctl — which cuts the other way too: a
record file nested anywhere below `.hcr/` is silently not a record, and is
worth a check of its own.

A service's effective rule set is the **union** of repo-scoped and
service-scoped records. Service records add constraints; they never relax or
override a repo record. Mention this if the user asks how the two interact.

The third piece of scaffolding is the `## Constraint Records` section of
[agents-section.md](agents-section.md), which Step 4 of `SKILL.md` writes into
the root `AGENTS.md`. A `.hcr/` directory nobody is told to maintain collects
exactly the records that were there the day it was created.

## 4. Propose the first five

The first records have one job: make the model obvious. So prefer rules that are
**already true today**. A first batch that validates green teaches the shape of
the thing; a first batch that surfaces forty violations teaches the user that
this tool is a chore.

Read the repo and match against this catalogue. Take the five with the
strongest evidence — an actual config file, an actual task, an actual CI step.

| If the repo has | Propose | Regulates | Binding |
|---|---|---|---|
| a lint task or linter config | source passes the configured linter | `maintainability` | `command` → the lint task |
| a test task | the test suite passes on every change | `reliability` | `command` → the test task |
| a formatter | source is formatted by the configured formatter | `maintainability` | `command` → the format-check task |
| a type checker | the codebase type-checks with no suppressed errors | `maintainability` | `command` → the typecheck task |
| CI plus a local gate task | CI runs the same gate contributors run locally | `architecture-fitness` | `grep` → CI file must contain the gate command |
| a pinned toolchain or runtime version | the toolchain version is pinned, not floating | `reliability` | `config-assert` → the version field |
| a committed lockfile | dependencies resolve from the committed lockfile | `reliability` | `config-assert` or `grep` |
| a vulnerability scanner | dependencies carry no known critical advisory | `security` | `command` or `external` |
| a `LICENSE` and third-party deps | dependencies stay within the licence allowlist | `compliance` | `external` |
| documented layering or ownership rules | the documented layering holds | `architecture-fitness` | `inferential`, until a check exists |
| a coverage threshold | coverage stays at or above the configured threshold | `reliability` | `config-assert` → the threshold |
| an `AGENTS.md` (any repo adopting this skill) | AGENTS.md carries the constraint-records section | `architecture-fitness` | `grep` → `AGENTS.md` must contain `## Constraint Records` |

Two rules for picking:

- **Name the real thing.** Not "the linter must pass" but the actual task, the
  actual config file, the actual CI job. A record that could belong to any repo
  belongs to none.
- **Do not invent enforcement.** If a rule matters but nothing checks it, that
  is `kind: inferential`, which owns no fields and claims nothing. Authoring a
  `command` binding to a script that does not exist fails validation outright.

Use the card format from `SKILL.md`, emit exactly five, then stop and wait.

## 5. Author and validate

Return to Steps 3 to 5 of `SKILL.md`. Every new record is `state: draft`,
including these — draft is what lets the user watch a rule's impact before it
gates anyone.

After validating, tell the user the two things they will want next: how to
promote a record from `draft` to `warn` to `enforced`, and that adding the next
record is just another file in `.hcr/` — there is no index or registry to keep
in sync.
