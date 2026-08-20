---
id: HCR-0408
title: Real HCR ids must stay in the 04xx band
scope: repo
regulates: maintainability
summary: "Every real record under .hcr must use an HCR-04xx id so searches cannot confuse it with a deliberately broken corpus fixture."
state: enforced
rationale: ADR-0002
enforced_by:
  - kind: command
    run: "tools/check-record-ids.py"
---
Run `mise run ids` (or `mise run check`) after adding or renumbering a
record under .hcr/. Choose the next free HCR-04xx id. Do not allocate a
fixture id to a real record or renumber corpus fixtures to make room;
each population owns its existing band.

This binding names the script by path rather than the `mise run ids` task
that wraps it, because R-104 resolves a path-shaped command and ignores
everything else. A `mise run` reference would survive its task being
renamed or deleted without any rule noticing. Route a binding through
mise only when the task wraps something with no single path, the way
`lint` chains three separate tools; a lone script has a path, so it uses
it.

Classified `maintainability`: an out-of-band id does not change jigctl's
validation result, but it makes every later search and review more
expensive by forcing readers to distinguish real records from fixtures
by path instead of by id.
