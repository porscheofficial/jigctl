---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0004: Rationale references resolve through a per-prefix map

## Context

A record may carry an optional `rationale` field holding an opaque
prefixed identifier such as `ADR-0007`, `RFC-12` or `PRD-2026-11`. The
schema constrains its shape and deliberately not its prefix, because
teams record decisions in different systems and hard-coding one team's
convention into the framework is the failure mode this project exists to
avoid.

R-111 states that a rationale reference resolves to an artifact that
exists. It has stood deferred, filed as requiring the command executor.
That classification is wrong. R-104 checks that a path-shaped `run`
resolves on disk and R-110 checks that an `external.docs` path and its
anchor resolve; both are filesystem checks and both already ship. Asking
whether a referenced artifact exists is the same question.

What actually blocked R-111 is that an opaque identifier is not a path.
`ADR-0007` says nothing about which directory holds ADRs or how their
filenames are built. The record cannot supply that without becoming
path-aware, and the schema cannot supply it without constraining the
prefix it just refused to constrain.

## Decision

The tree declares where each prefix resolves, in `jig.toml`:

```toml
[rationale]
ADR = "docs/adr/{rest}-*.md"
```

The key is the prefix, meaning everything before the first hyphen of the
identifier. The value is a path pattern relative to the tree root. Two
substitutions are available: `{id}` expands to the whole identifier and
`{rest}` to everything after the first hyphen. The expanded pattern is
then matched as a glob, and R-111 reports when it matches nothing.

A prefix with no entry is skipped rather than reported. jigctl has not
been told where those artifacts live, and a reference it cannot locate is
not evidence that the record is wrong. This mirrors R-104, which inspects
only a `run` whose first token looks like a path and stays silent
otherwise.

## Consequences

R-111 moves to the filesystem tier and ships alongside R-104 and R-110
instead of waiting for a milestone it never needed.

The check is opt-in for each prefix. A tree that declares nothing keeps
the behaviour it has today, and a typo inside an unmapped prefix goes
unreported indefinitely. That is the cost of refusing to guess where a
team keeps its decisions, and it is paid in the direction of not
inventing conventions on their behalf.

Matching is a glob rather than an exact path, so the descriptive part of
a filename can change without editing the records that point at it. A
decision that keeps its number through a rename still resolves.

The pattern is written by the tree that is being checked, so a careless
mapping can match far more than intended and make the rule vacuous
without appearing to. This is not defended against: a tree that declares
`service_globs` is already trusted to describe itself honestly.
