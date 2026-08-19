---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0001: Record architecture decisions

## Context

This repository already has a decision-record culture, but it is in the
wrong places. Locked decisions live as D1 through D15 inside plan files
under the `.omo/` directory, which is gitignored and therefore not in the
repository at all. Other rationale is scattered across the rejected
alternatives section of `docs/concepts.md` and `corpus/RULES.md`. A reader
of the repository cannot find why anything was chosen.

jigctl's own schema has a `rationale` field on records, and the deferred
rule R-111 will check that the artifact it names actually exists. The
project therefore needs a decision corpus with stable, resolvable ids.
This is not only documentation hygiene, it is a prerequisite for a rule
the project has already written down.

The constraint the repository owner set is that ADRs must not become
bloated, so that they do not have to be updated or superseded constantly.

## Decision

One ADR records one independently-lifecycled decision. The test is
calendar-based: if a decision could plausibly change on a different
schedule than the others in the document, it belongs in its own ADR. A
record bundling several decisions must be superseded in full whenever any
one of them changes, which multiplies churn rather than reducing it. The
one exception is a tight corollary that has no independent life without
its parent.

The format is Nygard's structure reduced to Context, Decision, and
Consequences, plus MADR-style YAML frontmatter for the machine-readable
fields. Sections for Decision Drivers, Considered Options, Pros and Cons
of the Options, and Confirmation are explicitly rejected as mandatory.
Those sections are how a one-page record becomes a multi-page one.

The lifecycle has three tiers. Frontmatter fields such as status,
supersedes, and superseded-by are always mutable. A dated Notes section
may be appended when new information refines a decision without reversing
it. The Context, Decision, and Consequences body is frozen once the status
is accepted. Changing what was decided means a new ADR that supersedes
this one. Editorial-only fixes to the frozen body are permitted and
recorded in git history. There is no status body section. Status lives in
frontmatter only because the body is frozen after acceptance while status
is not, making "the body is frozen" mechanically true.

Supersession is two-sided. The new ADR names what it supersedes and the
old one is flipped to superseded-by in the same commit. This repository
already enforces exactly this shape for records through R-102, and a
one-sided supersession is the most common integrity defect in real ADR
directories.

Status values are proposed, accepted, rejected, and superseded. An agent
proposes; a human accepts.

The layout is `docs/adr/NNNN-slug.md`, referenced in prose and in
rationale fields as ADR-NNNN. This repository needs the numeric form
because it makes resolving a reference a glob rather than a search, which
is what R-111 will need.

There is no index file for now. At this corpus size a hand-maintained
index is a drift liability; the directory listing is the index. An index
may be added when something generates it, and it will then be a derived
cache, never hand-edited. No external ADR tooling is adopted.

## Consequences

Rule R-111 becomes implementable once a prefix-to-location mapping exists,
without needing the M2 executor.

The frozen-body tier is currently a convention, not a gate. This
repository already owns the mechanism that would enforce it, which is the
SHA-256 backstop over corpus expectations at
`internal/hcr/testdata/expectations.sha256`. Adopting it here is
deliberately left for later rather than claimed now.

The granularity rule produces more, smaller files, and requires discipline
at the moment a decision is written.

The existing D1 through D15 decisions are not promoted into ADRs as a
batch. Most are implementation-level rather than architecturally
significant, they remain readable in their plan files, and a bulk
promotion is the exact failure mode this ADR's granularity rule exists to
prevent. A decision earns an ADR when it is genuinely architecturally
significant, at the moment it is locked.
