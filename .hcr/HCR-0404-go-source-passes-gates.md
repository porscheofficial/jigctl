---
id: HCR-0404
title: Go source must pass lint and static analysis before merge
scope: repo
regulates: maintainability
summary: "None of gofumpt's formatting drift, golangci-lint's style and complexity findings, or nilaway's nil-safety findings break anything the moment they appear — the code still compiles and runs exactly as before. What they cost is the next change: unformatted or overly complex code slows every future review, and an unaddressed nil-safety finding is a landmine the next person to touch that function inherits without being told. That is docs/concepts.md's maintainability question almost verbatim — does violating it make a future change more expensive while breaking nothing today — not reliability, because nothing here raises an active production-incident probability; it raises the cost of the next diff."
state: enforced
enforced_by:
  - kind: command
    run: "mise run lint"
---
Run `mise run lint` (or `mise run check`) before opening a pull request
that touches any Go file. Fix what it reports rather than suppressing it
inline — a suppressed finding is still a maintainability tax, just a
hidden one.
