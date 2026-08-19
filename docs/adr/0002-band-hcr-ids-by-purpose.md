---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0002: Band HCR ids by purpose

## Context

The repository has accumulated several distinct populations of Hard
Constraint Records (HCRs): real constraints enforced by jigctl, positive
fixtures in the test corpus, and deliberately broken negative fixtures.
The identifier namespace is already allocated into implicit bands: 00xx
through 02xx and 20xx for record-shape fixtures, 03xx for the
multi-service fixture, 04xx for the repository's own real records, and
05xx for the single-rule tree fixtures. How many ids each band currently
holds is deliberately not recorded here. That is inventory, it changes
whenever a fixture is added, and the decision does not rest on it.

The problem is that this allocation exists only in a gitignored plan file.
A new contributor or agent has no way to discover it and will allocate the
next id from whatever they happened to read last.

This lack of documentation creates a trap. The record-shape fixtures in
`corpus/records/` at least announce themselves: the falsifying ones are
named `r<NNN>-<slug>.md` and the valid ones `valid-<slug>.md`, and that
naming holds without exception. The tree fixtures have no such luxury.
R-112 requires a record's filename to agree with its own id, so a
fixture like `HCR-0502-dangling-supersedes-example.md` is named exactly
as a real record would be, carries `state: enforced`, and reads as a
plausible constraint in every frontmatter field. Only its path reveals
that it exists in order to be wrong. Searching for an id and reading
what comes back is therefore not enough to tell the two populations
apart, which is a worse trap for an agent than for a human.

## Decision

The bands named above are adopted as the allocation rule for future ids.
New real records added to `.hcr/` take the next free id in the 04xx
band. New fixtures take an id inside the band matching their corpus.

This is a naming convention enforced by review, not by jigctl. The CLI is
correct not to enforce it: id uniqueness (R-101) is scoped to a tree root
defined by `jig.toml`. The `indexTree` routine globs only
`<tree-root>/.hcr/*.md` plus one glob per `service_globs` match. Since
jigctl's own `jig.toml` sets `service_globs = []`, its tree is exactly the
files in `.hcr/`, and discovery never descends into `corpus/`. Each
fixture carries its own `jig.toml`, making it a separate tree root. Two
records in different trees are not in the same namespace, so R-101 cannot
and should not fire. Collapsing separate trees into one global namespace
to police a naming habit would be the wrong trade.

## Consequences

The band costs nothing to follow and buys unambiguous search: a real
record's id appears only where real records live, while a fixture id
appears only in the corpus. However, because nothing detects a
violation, the convention degrades silently over time unless reviewers
enforce it.

The convention will run out of room in a corpus band eventually. Widening
a band is a cheap editorial act because the ids are opaque and nothing
computes on them.

This convention could become a real constraint once the project can
express it, but there is no timeline or assigned rule number for that.
