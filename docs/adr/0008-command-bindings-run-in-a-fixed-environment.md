---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0008: Command bindings run in a fixed environment

## Context

ADR-0003 deferred argv handling and timeouts to the milestone that builds
the executor. The schema already says that `run` is executed from the
repository root, while making `timeout_secs` optional. Execution still
needs one complete process contract so that an omitted field or ambient
terminal cannot change whether a check finishes.

The repository has direct script bindings whose executable files use an
`env` shebang, accept no arguments, and read paths relative to the tree
root. Other bindings invoke `mise` through `PATH`. The contract must
support both forms unchanged. It must also account for task runners that
start further processes, because killing only their direct child can leave
tools running and holding shared caches after a timeout.

## Decision

The child process working directory is the absolute tree root used to
build the execution plan. This applies to every binding, including a
service-scoped binding, because scope controls applicability rather than
path resolution and the schema defines `run` from the repository root.

The child inherits every variable from the invoking process except
`PWD`. All inherited `PWD` entries are removed and one `PWD` entry holding
the absolute tree root is added; no other variable is stripped, changed,
or added. Inheriting `PATH` preserves bare command lookup and `env`
shebangs, while replacing `PWD` makes the advertised directory agree with
the process working directory.

The child's standard input is connected to the operating system null
device opened for reading. It is never inherited, so a command that tries
to prompt receives end of file instead of waiting for an interactive
caller that may not exist.

An absent `timeout_secs` resolves to a wall-clock timeout of 120 seconds.
A declared value replaces that default for the binding.

Standard output and standard error are captured separately and are not
streamed while the command runs. Each capture is capped at 1,048,576
bytes. Reaching either cap kills the process group, waits for it to be
reaped, and reports that the output limit was reached rather than keeping
an unbounded transcript in memory.

Each command starts as leader of a new process group. On timeout, jigctl
sends `SIGKILL` to the negative child process id, thereby addressing the
whole group, and waits for the direct child to be reaped before returning.
Killing the group rather than only the direct child prevents a task runner
from leaving its tools alive after the binding has stopped being observed.

The `run` string is split at each maximal run of Unicode whitespace, empty
fields are discarded, and the resulting first field is the executable
with every remaining field passed as one argument, without quote,
escape, expansion, or metacharacter processing. This matches R-104's
existing whitespace tokenisation of the executable and invokes direct
scripts without inserting another program between jigctl and the file.

For bindings sharing a `ref`, the effective timeout is the maximum
resolved timeout among execution-eligible participants. Resolution means
the declared `timeout_secs` when present and 120 seconds otherwise. A
group with no execution-eligible participant does not execute and has no
effective timeout. Using only declared values could let a participant
declaring a short timeout shorten another participant that silently took
the larger default. A shorter timeout is a preference, not a safety
property, so it cannot constrain another record's shared execution.

## Consequences

Commands see the caller's installed tools and credentials. The process
contract is reproducible about directory, input, limits, and tokenisation;
it is not a hermetic environment or a security boundary.

Commands that require quoting or redirection must move that behavior into
an executable script or task-runner entrypoint. This keeps command
interpretation independent of a user's interactive configuration.

Captured output can be reported after completion without interleaving
with jigctl's own output. The cap makes memory use finite, at the cost of
stopping a command that emits more output than the report can retain.

Process-group signalling relies on operating-system process-group
semantics. A descendant that deliberately creates a different session is
outside the original group and cannot be guaranteed to receive its
signal.
