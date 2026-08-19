---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0003: Execution authorization is per-invocation

## Context

A record's `command` binding names a shell command that proves a
constraint holds, for example `mise run lint`. Milestone M1 validates
records but never runs anything; `exec` appears nowhere in the Go source
today. A future milestone will run these bindings, at which point jigctl
executes strings that arrive from a repository it has just read. The
question this ADR settles is whether a decision to trust such a string
can ever be remembered.

## Decision

Authorization to run command bindings is supplied at invocation time by
the caller and is never persisted. It is not stored in the repository,
not in a user-level state directory, and not keyed by a hash of
anything.

First, there is nothing durable to hash. The string `mise run lint` is
stable but its meaning lives in `mise.toml`, which can be rewritten at
any time. An approval keyed to the string survives a change to what the
string does. A concrete attack is to get a legitimate-looking binding
approved, then change the task definition it resolves to. The approval
still matches and nothing re-prompts. Hashing the record instead is
worse in both directions, as it re-prompts when the surrounding prose is
edited and stays silent when the target is weaponised.

Second, the direnv model does not transfer here. The tool direnv hashes
`.envrc` and remembers the result. That works because `.envrc` is the
content that gets executed. A jigctl record is a pointer to content that
lives elsewhere. The same mechanism applied to a pointer gives the
failure described above. This is a rejected alternative.

The obvious repair is to follow the pointer: resolve `mise run lint` to
the task definition it names, hash that, and re-prompt when the
resolution changes. This is not feasible, because resolution has no
bottom. jigctl would have to understand mise, make, npm, task, just,
cargo and whatever a team adopts next, which is an unbounded parser
surface. Even a correct resolution lands on scripts that can change
without the resolution changing. The asymmetry with direnv is that
`.envrc` is a leaf, so hashing it terminates. A record is never a leaf.

Third, trust stored in the repository is self-certifying. The tool mise
shipped a vulnerability of exactly this shape where a repository-local
config could declare its own config path trusted. That satisfied the
trust gate before it ran. The fix was to strip trust-controlling fields
from any non-global config. The durable principle is that anything that
decides whether a prompt happens must be unreachable from inside the
thing being gated. The tool git applies the same rule by never cloning
hooks at all.

A prompt is not the boundary either, because approval prompts train
reflexive assent. This decision is about what is not stored, not a claim
that prompting makes execution safe.

## Consequences

Every invocation that runs command bindings must carry its own
authorization. The unit is the invocation, not the binding: an
interactive run asks once and covers every binding it executes, so the
cost does not grow with the number of records. In automation the caller
states its policy explicitly. There is no path where the answer is
remembered, and that is the point.

The cost is real but bounded. This is more friction than direnv-style
remembered trust, and users accustomed to that model will ask for it.
The answer is that per-invocation scope already removes most of what
they want, and the rest cannot be given back without reintroducing the
referent problem.

How authorization is expressed is deliberately not decided here. Flag
names, environment variables, defaults for interactive versus
non-interactive use, argv handling, and timeouts belong to a later
decision. Those belong to the milestone that builds the executor, and
they can change without reopening this decision.

This decision is necessary but may not be sufficient. If jigctl is ever
run against an unreviewed pull-request checkout in CI holding privileged
credentials, invocation-level authorization alone will not contain it,
and that context will need an out-of-tree policy or must stay
non-executing. Whether jigctl supports that context at all is a separate
decision belonging to the executor milestone. It cannot be settled
before the executor exists, and its answer does not change what is
decided here: under either outcome, remembered trust is the thing an
attacker would target.
