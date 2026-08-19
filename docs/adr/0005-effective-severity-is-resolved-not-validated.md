---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0005: Effective severity is resolved, not validated

## Context

Two rules describe how a binding arrives at its severity and cadence.
R-106 says defaults are injected per kind tier: the mechanical kinds
default to blocking on change and in CI, while `agent-review` and
`inferential` default to advisory. R-105 says a record in `warn` state
downgrades its bindings to advisory, so a rollout can report without
gating.

Both have stood deferred, filed as requiring the command executor. As
with R-111, that classification confuses when a fact becomes useful with
when it can be determined. Neither rule needs anything to run. Both are
transformations over data already on disk.

The defaults are in fact already written down. The schema declares them
per kind as JSON Schema `default` keywords. But `default` is an
annotation rather than an assertion: a validator records it and applies
nothing, and no code in this repository reads `severity` or `cadence` at
all. The values are correct, documented, and inert.

## Decision

R-105 and R-106 resolve rather than report. Like R-108 they produce a
view and emit no diagnostic, because a transformation cannot fail. They
resolve in a single pass, because they compose: defaults are injected
first, and the `warn` downgrade then applies to the result. A binding
that declares `severity: blocking` inside a record in `warn` state
resolves to advisory, which is the case the corpus already carries.

The per-kind defaults are mirrored in Go rather than read out of the
embedded schema at run time, and a test walks the schema and asserts the
two agree. Reading them at run time would make drift impossible, but it
would also bind the resolver to the schema's conditional structure, where
a restructuring would yield no defaults at all and quietly mislabel every
binding. A mirror that drifts fails a test loudly instead. This is the
division `mise run shape` already draws: the product consumes the schema,
and separate checks police its shape.

## Consequences

R-105 and R-106 become enforced without the executor. They need only
resolution over a record the schema has already validated, and their
record-corpus deferral citations retire as R-111's did.

Severity and cadence gain a defined resolved value before anything
consumes them, so the executor milestone inherits settled semantics and
is left to decide only what running a binding means.

The mirror is a second copy of a fact the schema owns. It is safe only
while the test pinning it keeps passing, and a contributor changing a
default in one place has to be shown the other. That is what the test is
for, and it is the same bargain the frozen expectation digest already
makes.

Nothing consumes the resolved values yet, so the resolver is written
against a specification rather than a caller, which risks an interface
shaped for no one. R-108 was built the same way and has held. The
alternative is leaving semantics that are already decided sitting
unexecuted in a schema annotation, which is how they came to be filed
under the wrong milestone in the first place.
