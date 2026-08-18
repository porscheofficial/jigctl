---
id: HCR-0403
title: Schema's per-kind prohibitions must match the rule register
scope: repo
regulates: architecture-fitness
summary: "schema/hcr.schema.json declares, per binding kind, exactly which fields are prohibited; corpus/RULES.md's 'Allowed properties by binding kind' table separately declares, per kind, what that same prohibition set should be. Each file can look entirely self-consistent on its own — the schema parses, the table renders — while silently disagreeing with the other about, say, whether grep prohibits 13 fields or 14. That disagreement is invisible in either file read alone and only exists in the relationship between them, which is exactly what docs/concepts.md means by architecture-fitness: reading any one file in isolation shows nothing wrong."
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
