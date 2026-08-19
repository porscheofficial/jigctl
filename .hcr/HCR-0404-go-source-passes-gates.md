---
id: HCR-0404
title: Go source must pass lint and static analysis before merge
scope: repo
regulates: maintainability
summary: "Go source must be gofumpt-clean and free of golangci-lint and nilaway findings before merge. None of those findings break anything the moment they appear; they raise the cost of every change made to that code afterwards."
state: enforced
enforced_by:
  - kind: command
    run: "mise run lint"
---
Run `mise run lint` (or `mise run check`) before opening a pull request
that touches any Go file. Fix what it reports rather than suppressing it
inline — a suppressed finding is still a maintainability tax, just a
hidden one.

Classified `maintainability` rather than `reliability`: the code still
compiles and runs exactly as before. Unformatted or overly complex code
slows every future review, and an unaddressed nil-safety finding is a
landmine the next person to touch that function inherits without being
told — cost on the next diff, not production-incident probability today.
