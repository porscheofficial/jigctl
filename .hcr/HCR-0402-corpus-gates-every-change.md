---
id: HCR-0402
title: Corpus fixtures gate every change to schema or rules
scope: repo
regulates: architecture-fitness
summary: "corpus/RULES.md declares which rule ids exist and which fixture demonstrates each one, schema/hcr.schema.json defines what counts as valid, and every fixture separately declares — in its own expectation block — the verdict and rule it claims to prove. None of that is visible by reading any single one of those files: a fixture can look correct in isolation while quietly mismatching the rule it claims to cover, or the schema can drift out from under a fixture that used to agree with it, and nothing about either file's own content would show it. That is architecture-fitness by docs/concepts.md's own test — the defect exists only in the relationship between files that live apart, the same way a circular import is invisible until you look at two files at once."
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
