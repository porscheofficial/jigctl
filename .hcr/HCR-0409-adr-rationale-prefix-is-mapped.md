---
id: HCR-0409
title: ADR rationale references must have a repository mapping
scope: repo
regulates: architecture-fitness
summary: "jig.toml must map ADR rationale identifiers to docs/adr artifacts so R-111 can verify every cited repository decision exists."
state: enforced
rationale: ADR-0004
enforced_by:
  - kind: config-assert
    file: jig.toml
    path: /rationale/ADR
    op: equals
    value: "docs/adr/{rest}-*.md"
---
Keep `/rationale/ADR` in jig.toml equal to
`docs/adr/{rest}-*.md`. When a record cites `ADR-NNNN`, place the accepted
decision under docs/adr/ with `NNNN-` at the start of its filename. Fix a
broken citation or restore its decision artifact rather than broadening
the glob until an unrelated file happens to match.

Classified `architecture-fitness`: the identifier is valid in a record
and the ADR is valid on disk when each is read alone. The defect exists
only when jig.toml fails to connect the prefix used by one artifact to
the location of the other, leaving R-111 unable to inspect that
relationship.
