---
id: HCR-0406
title: HCR validation lives only in internal/hcr; only main may exit
scope: repo
regulates: architecture-fitness
summary: "Whether a stray os.Exit call anywhere in the module is a violation depends entirely on which file it sits in, a fact invisible from the call site itself — os.Exit(1) reads as ordinary Go whether it appears in cmd/jigctl/main.go or buried inside internal/hcr. The defect only exists in the relationship between that call and the one file the whole codebase has agreed owns process exit, exactly the class of problem docs/concepts.md assigns to architecture-fitness, where reading any single file in isolation shows nothing wrong and the violation only surfaces once you compare it against a second file entirely. Letting internal/hcr, or anything else, exit the process directly would also mean HCR validation had grown a second implicit entrypoint, defeating the point of a single, callable, testable validation package."
state: enforced
enforced_by:
  - kind: grep
    file: "**/*.go"
    forbid: ["os.Exit("]
exceptions:
  - scope: cmd/jigctl/main.go
    reason: "cmd/jigctl/main.go is the CLI's sole process entrypoint and the one place permitted to turn a validation result into a process exit code; every other file returns a value instead so the logic stays callable and unit-testable without spawning a process."
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
