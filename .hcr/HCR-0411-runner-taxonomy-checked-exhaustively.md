---
id: HCR-0411
title: Runner outcome taxonomy must be checked exhaustively
scope: repo
regulates: architecture-fitness
summary: "golangci-lint must keep exhaustive enabled with default-signifies-exhaustive false, so every Reason value is named in the switches and maps over it."
state: enforced
rationale: ADR-0012
enforced_by:
  - kind: config-assert
    file: ".golangci.yml"
    path: "/linters/settings/exhaustive/default-signifies-exhaustive"
    op: equals
    value: false
---
Leave `default-signifies-exhaustive: false` in place when you edit
.golangci.yml. Adding a `default:` arm to a switch over `Reason`, or a
fallback branch around `reasonData`, is not a substitute for naming a new
value: this setting is what makes the linter reject the fallback and
demand the case. If a new `Reason` genuinely has nothing to report, say
so at the value, not by relaxing the setting.

The assertion targets the setting rather than the `enable` list, because
`exhaustive` can be enabled and still approve everything: with
`default-signifies-exhaustive: true` one `default:` arm silences it for
the whole switch. That single key is the difference between a linter that
runs and a linter that checks.

Classified `architecture-fitness` rather than `reliability`: no single
file is wrong when this breaks. A new value on `Reason` in
internal/runner/verdict.go is correct read alone, and `reasonData` in
internal/runner/report.go is correct read alone. Only the relationship
between the two is broken, and the compiler has nothing to say about it.
