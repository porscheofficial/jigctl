---
id: HCR-0412
title: Go tests must run with race detection and shuffled ordering
scope: repo
regulates: reliability
summary: "The mise test task must invoke go test with -race and -shuffle=on, so shared mutable state and inter-test ordering coupling fail the run."
state: draft
enforced_by:
  - kind: config-assert
    file: "mise.toml"
    path: "/tasks/test/run"
    op: matches
    value: "-race"
  - kind: config-assert
    file: "mise.toml"
    path: "/tasks/test/run"
    op: matches
    value: "-shuffle=on"
---
Keep both flags on the `test` task in mise.toml. `-race` is what turns a
data race from an occasionally wrong answer into a failed run, and
`-shuffle=on` is what stops one test from silently depending on state
another left behind. If you add a task that runs part of the suite, give
it the same two flags rather than treating them as a slow path that only
`test` pays for.

Both assertions read `/tasks/test/run` rather than searching mise.toml
for the flags. A file-wide search would still pass with the flags deleted
from `test` and left on `corpus`, which is exactly the edit this record
exists to catch.

Classified `reliability` rather than `maintainability`: a race or an
ordering dependency that the suite never exercises does not make the next
change more expensive. It makes the eventual failure non-reproducible,
and it lets that failure reach a user instead of a contributor.
