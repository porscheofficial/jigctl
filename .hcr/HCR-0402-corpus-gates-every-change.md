---
id: HCR-0402
title: Corpus fixtures gate every change to schema or rules
scope: repo
regulates: architecture-fitness
summary: "Every change to schema/hcr.schema.json, corpus/RULES.md, or any file under corpus/ must pass the full corpus run. A fixture can look correct in isolation while disagreeing with the schema or with the rule it claims to prove."
state: enforced
enforced_by:
  - kind: command
    run: "mise run corpus"
---
Run `mise run corpus` (or `mise run check`) before proposing any change
that touches schema/hcr.schema.json, corpus/RULES.md, or any file under
corpus/. Do not hand-verify a subset of fixtures and call it good — the
whole point of this gate is that a fixture passing in isolation proves
nothing about whether it still agrees with the schema and the rule
register.

Classified `architecture-fitness`: corpus/RULES.md declares which rule
ids exist and which fixture demonstrates each, schema/hcr.schema.json
defines what counts as valid, and each fixture declares its own expected
verdict and covered rule. A mismatch between those three lives only in
the relationship between files that sit apart — invisible in any one of
them read alone, the same way a circular import is.
