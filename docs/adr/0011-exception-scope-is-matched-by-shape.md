---
status: proposed
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0011: Exception scope is matched by shape

## Context

The schema gives `exceptions[].scope` one string field and describes two
meanings: a path glob or a service name. It supplies no discriminator. A
matcher must distinguish those meanings before it can decide which findings
a waiver suppresses.

The distinction matters because binding results have different locations. A
`grep` finding names a file. A `config-assert` finding names a file and a
pointer. A `command` exit code has no file locus. The other binding kinds
do not execute in this milestone.

The repository's own waiver names `cmd/jigctl/main.go` and has no `until`.
It permits the one `os.Exit(` call while leaving all other matching files in
the search set. Corpus waivers also attest service directory names.

## Decision

Scope shape is determined mechanically. First, a scope equal byte-for-byte
to a discovered service's tree-relative directory is service-shaped. If that
test fails, a scope containing any of `/`, `*`, `?`, `[`, `{`, or `\` is
path-shaped. Service equality takes precedence when both tests could apply.
A scope satisfying neither test has an invalid scope shape. ADR-0012
classifies the resulting outcome; this decision defines that condition.

A path-shaped scope is a glob interpreted by
`github.com/bmatcuk/doublestar/v4`. Matching is rooted at the tree root,
does not descend into `.git/`, and does not follow symlinks. Authored-path
resolution and confinement follow ADR-0010, Config assertions use JSON
Pointer addresses; they are not redefined here.

Applicability is fixed by binding kind:

- A `command` finding has no locus, so no exception scope suppresses it.
- A `grep` finding is matched against the path of its file locus.
- A `config-assert` finding is matched against the file component of its
  `(file, pointer)` locus; the pointer does not participate.
- An `external` binding is not run and produces no suppressible finding.
- An `agent-review` binding is not executed and produces no suppressible
  finding.
- An `inferential` binding is not executed and produces no suppressible
  finding.

A service-shaped scope matches findings whose locus falls under that
service's directory. This rule also applies when the exception belongs to a
repo-scoped record: the binding evaluates the whole tree, and only findings
carry a location that permits selective suppression. A finding without a
locus cannot be selected by service and is not suppressed. A command exit
code is the material case.

Exceptions suppress findings after evaluation. They never remove inputs from
the evaluated set. Suppression is the only choice that can report how many
findings a waiver absorbed. Silently shrinking the input set has the same
shape as the unchecked-rendering-as-pass failure forbidden by the verdict
model.

Each finding carries `waived_by`, a list of exception identities. The list
necessary because the schema imposes no uniqueness constraint on exceptions,
so more than one may suppress a finding. An exception identity is the pair
`(record path, exception index)`. It is unmatched exactly when that identity
appears in no finding's `waived_by`. The report names every unmatched
exception
so an author can see that the waiver did not apply.

That test is deliberate. A path-shaped scope may match files that were
evaluated yet suppress no finding. The output cannot distinguish that case
from a scope that matched no file, and no reachable data separates them.

Expiry is resolved before matching. An expired `until` produces an R-107
validation diagnostic, and `ExecutionPlan` returns no plan when any
diagnostic exists. The matcher therefore never receives an expired exception
and must not test expiry. An absent `until` means the exception does not
expire. The repository's own waiver has no `until`, so only that case is
exercised by the dogfood set. This agrees with ADR-0006, Exception expiry
uses
one invocation clock.

## Consequences

Waivers remain observable because evaluation still visits every declared
input and reports suppressed findings separately. Duplicate waivers remain
individually accountable through their identities.

Service waivers select located findings from repo-scoped bindings without
multiplying executions. They cannot imply that an unlocated command result
was selectively checked or suppressed.

Invalid scope shapes and unmatched exceptions remain visible rather than
quietly broadening or narrowing a waiver. Runtime matching has no second
clock or expiry policy.
