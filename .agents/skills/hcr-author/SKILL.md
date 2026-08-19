---
name: hcr-author
description: Discover a repository's Harness Constraint Records (HCRs), propose five new records that fit it, and author the ones the user approves. Use when asked to propose, add or create HCRs, to adopt or bootstrap jigctl in a repo, to set up .hcr records or jig.toml, or to turn existing linters, CI gates, architecture rules or contributor-doc conventions into machine-checkable records. Triggers: HCR, jigctl, harness constraint record, .hcr, jig.toml, constraint harness, propose records, what should this repo enforce.
---

# HCR author

An HCR is one versioned markdown file holding both a rule for humans and the
machine check that enforces it, so guidance and enforcement cannot drift apart.
This skill finds the records a repository already has, proposes five it is
missing, and writes the ones the user approves.

## Workflow

| Step | Action | Output |
|---|---|---|
| 1 | [Discover](#step-1--discover) the authoritative record set | existing records, or none |
| 2 | [Propose](#step-2--propose-exactly-five) exactly five candidates | five cards, then **HARD STOP** |
| 3 | [Author](#step-3--author-the-approved-records) only what was approved | files in `.hcr/` |
| 4 | [Validate](#step-4--validate) | a green tree |

Step 2 ends the turn. Authoring a record the user has not picked is a failure of
this skill, not a shortcut.

## Step 1 — Discover

Resolve the record set from the tree configuration. Do not glob for `*.md` and
guess — that pulls in test corpora, fixtures and vendored examples.

1. Walk up from the working directory to the nearest `jig.toml`. That directory
   is the **tree root**. If there is none anywhere, read
   [references/bootstrap.md](references/bootstrap.md) and follow it instead.
2. Read `service_globs` from `jig.toml`. Expand each glob. Keep a match only if
   it contains a `.hcr/` directory.
3. The authoritative record set is exactly:
   - `<root>/.hcr/*.md` — these are `scope: repo`
   - `<service>/.hcr/*.md` for each kept match — these are `scope: service`
4. Everything else is out of scope by construction, and this is load-bearing:
   - a `.md` file **not** inside a `.hcr/` directory is not a record
   - a subdirectory carrying its **own** `jig.toml` is a separate tree — never
     merge its records into this one
   This is what keeps a project's own fixtures and example trees from being
   mistaken for its live rules. Apply it literally rather than pattern-matching
   on directory names.
5. If the resolved set is empty, read
   [references/bootstrap.md](references/bootstrap.md) and follow it instead.

Now read every record you found. You need three things from them: which ids are
taken, which constraints are already covered, and the repo's house style for
`summary` and body prose.

Run `jigctl validate <root>` if a binary or source checkout is available. Never
start authoring into a tree that is already failing — report the findings and
let the user decide.

## Step 2 — Propose exactly five

### Scan for candidates first

Look for constraints that already exist in the repo but are not yet recorded.

**Enforced but not recorded** — a check runs, but nothing states what it protects:

- task runners: `Makefile`, `mise.toml`, `justfile`, `Taskfile.yml`,
  `package.json` scripts, `pyproject.toml`, `Cargo.toml`
- linters and formatters: `.golangci.yml`, `.eslintrc*`, `biome.json`,
  `ruff.toml`, `.rubocop.yml`, `clippy.toml`, `.editorconfig`
- CI: `.github/workflows/*`, `.gitlab-ci.yml`, `.circleci/config.yml`
- types, tests, coverage: `tsconfig.json`, `mypy.ini`, `pytest.ini`, jest and
  vitest config, coverage thresholds
- supply chain: lockfiles, `dependabot.yml`, `renovate.json`, `SECURITY.md`,
  `LICENSE`

**Recorded but not enforced** — prose rules nobody can mechanically check:

- `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, ADRs, `docs/`, README constraints
- layering and ownership rules implied by the directory tree
- review conventions repeated in git history or PR templates

Rank candidates by two questions: does this get violated in *this* repo, and is
the check cheap. Drop anything that fails either.

### Card format

Emit five cards and nothing else. No preamble, no file contents, no
implementation detail.

```
### N. <title, phrased as it would read in a violation message>
- Regulates: <value> — <at most 10 words on why this value, not its neighbour>
- Scope: repo | service:<name>
- Cadence: [<values>]
- Binding: <kind> → <the concrete thing it binds to>
<One sentence: what must become true, and what exists today.>
```

### Rules for the set of five

- No duplicates or near-duplicates of an existing record.
- Prefer a mix: three or four that bind to something already present in the repo
  and are true today, one or two that define something genuinely new.
- Every proposal must be checkable exactly as worded. If you cannot name the
  binding target, the proposal is too vague — rewrite it or drop it.
- If the tree resolved no services, every proposal is `scope: repo`. A
  `scope: service` record has nowhere legal to live without a service directory.
- Pick `regulates` with the discriminator questions in
  [references/record-format.md](references/record-format.md), not by vibe. It is
  the field reviewers argue about.

Then stop and wait. The user will accept, reject or refine.

## Step 3 — Author the approved records

Read [references/record-format.md](references/record-format.md) for fields,
binding kinds and the constraints the validator enforces.

Four rules override anything you infer from existing records:

- **`state: draft`, always.** A new record is an observation, not yet a gate.
  Draft is how its impact gets measured before it blocks anyone. Promoting it to
  `warn` or `enforced` is a separate decision the user makes later.
- **`summary` is one sentence, at most 25 words**, naming the artifact and what
  must be true about it. No rationale, no history, no hedging. Agents read
  frontmatter to decide whether a record applies at all; every wasted word is
  context the reader pays for and does not need.
- **`exceptions[].reason` is at most 15 words** — what is exempt, and why. Name
  the specific thing. "Legacy" and "special case" are not reasons.
- **Reasoning belongs in the body**, never the frontmatter. The optional
  `rationale` field is the one exception, and it is not prose: it holds an
  opaque reference to a decision recorded elsewhere, such as `ADR-0007`. Set
  it when the rule implements a decision that is already written down, and
  keep the argument for that decision in the body regardless.

For everything else, match the house style: read two existing records and follow
their body shape, heading structure and voice.

Identity, filename and location:

- **id** is the next free `HCR-NNNN`, unique across the tree. Scan every id in
  the whole repository, not just the tree — colliding with a fixture id is legal
  but confusing. If live records occupy a numeric block, continue that block.
- **filename** is `HCR-NNNN-<slug>.md` where the id matches the record's own
  `id` field and the slug is `[a-z0-9]+(-[a-z0-9]+)*`.
- **location** must agree with scope: `repo` → `<root>/.hcr/`, `service` →
  `<service>/.hcr/`.

## Step 4 — Validate

Run `jigctl validate <root>`, then the repo's own gate if it has one
(`mise run check`, `make check`, `npm run check`). Report the result honestly,
including a tree that was already red before you touched it.

When a record and the tooling disagree, fix the record. Never edit the schema, a
fixture, or the check being bound to in order to make a new record pass — that
inverts the entire point of the harness. If they genuinely conflict, say so and
stop.
