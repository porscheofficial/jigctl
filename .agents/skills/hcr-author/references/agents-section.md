<!--
The constraint-records section every adopting repo carries in its root
AGENTS.md. Written at Step 4 of SKILL.md, after the approved records exist.

The heading and the bullets below are normative: reproduce them verbatim.
The intro paragraph is not — replace the placeholder with one or two
sentences about the repo at hand. A repo may append a sentence to a bullet
where it has something local to add; it must not reword what is here.
-->

## Constraint Records

`.hcr/` holds this repo's Harness Constraint Records — each one a rule and
the check that enforces it, in a single versioned file, so the two cannot
drift apart. Keeping them current is part of the change that affects them,
not follow-up work.

- Read the records frontmatter-first. Every `.hcr/*.md` opens with a YAML
  block carrying `title`, `scope`, `regulates`, `summary` and `state` —
  that block is the index. Scan every record's frontmatter, then open only
  the bodies that bear on the change in front of you; `summary` is capped
  at 25 words so the scan stays cheap. Where a record binds a `command`,
  its `run` is what decides the rule — run it rather than reasoning about
  whether the code complies.
- A change that introduces a constraint carries its record in the same
  commit. If a contributor could violate the new rule without noticing, it
  needs a record — a rule that lives only in a review comment, or only in
  this file, is enforced by nobody.
- A change that invalidates a record updates that record in the same
  commit. Renaming a task, moving a package or dropping a check moves the
  binding target with it, and a record pointing at something that no longer
  exists is worse than no record at all.
- Not every preference is a record. Two questions decide it: does it get
  violated in *this* repo, and can you name the concrete thing a check
  would bind to. Whatever fails either is a style note and belongs in this
  file instead.
- A new record is `state: draft`. Draft is how a rule's impact is measured
  before it gates anyone; promoting it to `warn` or `enforced` is a
  separate, deliberate change.
- Author records with the `hcr-author` skill rather than copying an
  existing one by hand — it owns the field limits, the `regulates`
  discriminators and the id and filename rules. When a record and the
  tooling disagree, fix the record: never the check it binds to.
