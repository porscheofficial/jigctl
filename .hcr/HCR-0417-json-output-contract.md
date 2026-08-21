---
id: HCR-0417
title: JSON output contract must be respected
scope: repo
regulates: behaviour
summary: The runner must emit exactly the JSON structure defined in the versioned schema when the json format is requested.
state: enforced
enforced_by:
  - kind: command
    run: ".hcr/checks/check-json-output.py"
---
This record governs the JSON output contract, ensuring backward compatibility for external consumers. It regulates `behaviour` because violating the contract breaks a consumer who changed nothing.

### Recursion Hazard

This rule binds to a command script (`.hcr/checks/check-json-output.py`). Because command bindings inherit `cmd.Dir` (the tree root) and the ambient environment (minus `PWD`), any environment variables present during the runner's invocation are passed down to the check. 

Since `JIGCTL_ALLOW_EXEC=1` authorizes the execution of command bindings, a script that invokes `jigctl run .` under an inherited allow-exec environment would trigger itself again, recursing without bound. To prevent this, the script explicitly unsets `JIGCTL_ALLOW_EXEC` before invoking the runner.

### Enforcement

This record now gates (`state: enforced`). It observed the command binding's behaviour in the harness for a period as `draft` before this deliberate promotion.

The JSON output contract is also enforced by a Go test (`internal/runner/json_contract_test.go`) which runs as part of `mise run test`; this record's command binding additionally gates `mise run check`.
