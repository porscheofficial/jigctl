---
id: HCR-0403
title: Schema's per-kind prohibitions must match the rule register
scope: repo
regulates: architecture-fitness
summary: "The per-kind prohibition sets in schema/hcr.schema.json must match the 'Allowed properties by binding kind' table in corpus/RULES.md. Both files stay self-consistent while disagreeing with each other about what a given kind prohibits."
state: enforced
enforced_by:
  - kind: command
    run: "tools/schema-shape.py"
---
Run `tools/schema-shape.py` directly any time you add a kind, add a
field, or edit a per-kind prohibition in schema/hcr.schema.json or
corpus/RULES.md. This record's `run` is deliberately a bare path rather
than an indirect `uv run --script` invocation, so it doubles as proof
that a path-shaped command really does resolve on disk — do not route
this particular binding through mise or through a wrapping interpreter.

Classified `architecture-fitness`: the schema parses and the table
renders no matter which of the two is wrong, so a disagreement about,
say, whether `grep` prohibits 13 fields or 14 is invisible in either
file read alone and exists only in the relationship between them.
