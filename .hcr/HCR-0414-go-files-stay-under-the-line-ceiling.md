---
id: HCR-0414
title: No Go file may exceed 250 pure lines
scope: repo
regulates: maintainability
summary: "Every Go file must stay at or under 250 pure lines, counting neither blank lines nor comment-only lines."
state: enforced
enforced_by:
  - kind: command
    run: ".hcr/checks/check-file-loc.py"
---
Run `.hcr/checks/check-file-loc.py` before proposing a change that grows
a Go file. A file at the ceiling is a signal to split
the unit along a seam that already exists, not to move code into a
sibling file to buy room. Test files count too: a 300-line test file is
as hard to read as a 300-line implementation.

Pure lines are source lines that are neither blank nor comment-only, so
neither documentation nor spacing pushes a file over. The count is
deliberately not `wc -l`, which would make a well-commented file look
worse than a dense one.

Classified `maintainability` rather than `architecture-fitness`: an
oversized file compiles, passes its tests, and misbehaves in no way at
all. It charges every later reader for the size, and it hides the seam
that the split would have made obvious — a cost on the next diff, not a
defect today.
