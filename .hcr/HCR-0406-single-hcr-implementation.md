---
id: HCR-0406
title: HCR validation lives only in internal/hcr; only main may exit
scope: repo
regulates: architecture-fitness
summary: "HCR validation is implemented once, in internal/hcr, as a library that returns diagnostics and errors as data. cmd/jigctl/main.go is the only file permitted to call os.Exit, so validation stays callable and testable without spawning a process."
state: enforced
enforced_by:
  - kind: grep
    file: "**/*.go"
    forbid: ["os.Exit("]
exceptions:
  - scope: cmd/jigctl/main.go
    reason: "The CLI's sole process entrypoint, and the only place permitted to turn a validation result into a process exit code; every other file returns a value instead."
  - scope: internal/runner/grep_test.go
    reason: "Test file mocking a Go source file with an os.Exit call."
---
HCR validation is implemented exactly once, in internal/hcr, as a plain
library: it returns diagnostics and errors as data and never terminates
the process itself. cmd/jigctl calls into that library and is the only
code permitted to translate its result into a process exit code, and it
does so in exactly one place: cmd/jigctl/main.go. If you find yourself
reaching for os.Exit anywhere else — including inside internal/hcr, and
including anywhere you are tempted to short-circuit a validation path —
return an error or a diagnostic instead and let main.go decide what the
process does with it.

Classified `architecture-fitness`: `os.Exit(1)` reads as ordinary Go at
any call site, so whether it is a violation depends entirely on which
file it sits in — nothing is wrong in the calling file read alone. A
second exit path would also give HCR validation a second, implicit
entrypoint, defeating the point of a single callable package.
