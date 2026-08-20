---
status: proposed
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0007: Execution authorization is explicit

## Context

ADR-0003 made authorization per invocation and left its expression to the
executor milestone: "Flag names, environment variables, defaults for
interactive versus non-interactive use, argv handling, and timeouts belong
to a later decision." The authorization gate now needs a complete interface
that both a person and an unattended caller can use.

The repository's check task is an unattended caller. It is also invoked by a
workflow on `pull_request`, where the checkout may contain changes to a
record, the command it names, or the task through which that command is
reached. Supplying authorization from the checkout cannot distinguish a
reviewed command from a hostile replacement.

ADR-0003 therefore also left open whether an unreviewed pull-request
checkout with privileged credentials should execute at all. That question
must be answered before the executor becomes part of the check task.

## Decision

The authorization flag is `--allow-exec`. The corresponding environment
variable is `JIGCTL_ALLOW_EXEC`. Supplying the flag authorizes execution.
The environment variable authorizes execution only when its value is exactly
`1`; an unset variable or any other value does not authorize execution.

When a terminal is present and neither mechanism supplied authorization,
jigctl asks once whether to execute all command bindings in the invocation.
The interactive default is denial: an empty response or any response other
than an explicit `yes` does not authorize execution. One affirmative
response authorizes every command binding in that invocation and is not
remembered.

When no terminal is present, the default is denial and jigctl does not
prompt. The caller must supply `--allow-exec` or set `JIGCTL_ALLOW_EXEC=1`.
Without authorization, each selected command binding is reported as a
blocked-unchecked verdict. It contributes exit code 1 rather than being
reported as a pass or as an operational failure.

Granting authorization grants arbitrary code execution with the runner's
credentials. Authorization is caller consent and is legitimately in-tree.
The check task may therefore carry `--allow-exec` in `mise.toml`, as the
per-invocation rule already permits. It is not and cannot be a security
boundary, because anything a pull request can edit it can also authorize.

The workflow permission model is least-privilege exposure and blast-radius
limitation. Branch protection, required reviews, and repository-level runner
and secret settings sit outside the tree and limit what an executed revision
can reach. They are not containment either. Neither caller consent nor that
permission model prevents a hostile revision running, and nothing in this
milestone claims to do so.

A workflow handling a fork-origin `pull_request` must refuse to execute
command bindings from that checkout. Neither the in-tree flag nor the
environment variable overrides this ruling. Such a workflow may validate the
records without execution, but enabling fork execution requires a later,
explicit out-of-tree policy decision.

A portable sandbox is not designed or promised here. Containing arbitrary
repository commands across supported platforms is a different milestone.

## Consequences

Interactive use has one denial-by-default question per invocation rather
than one question per binding. Automation remains predictable because the
non-interactive default is denial and no prompt can block an unattended run.

The check task must carry the authorization flag when it intentionally runs
commands. Its continued operation does not make the checkout trusted; the
workflow must independently minimize credentials and runner capability.

Missing authorization is visible and fails the run with exit code 1. A
caller cannot mistake an execution it declined to permit for a successful
check.
