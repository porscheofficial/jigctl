---
id: HCR-0416
title: AGENTS.md must carry the hcr-author constraint-records section
scope: repo
regulates: architecture-fitness
summary: "AGENTS.md must reproduce every bullet of the hcr-author skill's constraint-records section, so the practice the skill installs elsewhere stays documented here."
state: enforced
enforced_by:
  - kind: command
    run: ".hcr/checks/check-agents-hcr-section.py"
---
Run `.hcr/checks/check-agents-hcr-section.py` after editing either the
`## Constraint Records` section of AGENTS.md or the
skill's `references/agents-section.md`. The check asserts containment,
not equality: the intro paragraph is this repo's own and a bullet may
carry an appended sentence, but every bullet the skill ships must appear
verbatim. When the two disagree, decide which side is right and change
that side — never loosen the check to admit both.

Containment rather than a digest of the kind `expectations.sha256` keeps
over the corpus. There the fixtures are the specification, so a silent
edit inverts what jigctl means and the re-freeze ceremony earns its
keep. This section is guidance whose wording should stay improvable; a
digest would put every rewrapped line behind a ceremony while catching
nothing containment misses.

Classified `architecture-fitness` rather than `maintainability`: nothing
in AGENTS.md reads as wrong once the section drifts, and the file stays
coherent to anyone who only reads it. What breaks is the relation
between two artifacts — the skill vendored at `.agents/skills/hcr-author/`
tells other repos to carry a section this repo has stopped carrying, so
jigctl no longer demonstrates the practice it ships.
