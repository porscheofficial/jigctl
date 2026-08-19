---
id: HCR-0401
title: Schema must validate against its own metaschema
scope: repo
regulates: reliability
summary: "schema/hcr.schema.json must stay valid against the JSON Schema 2020-12 metaschema. It is embedded via go:embed at build time and never re-validated at runtime, so a non-conforming schema ships inside every copy of the binary."
state: enforced
enforced_by:
  - kind: command
    run: "mise run metaschema"
---
Run `mise run metaschema` (or let `mise run check` do it for you) any
time you hand-edit schema/hcr.schema.json. If check-jsonschema reports
the schema itself is invalid against the metaschema, fix
schema/hcr.schema.json — do not work around it by loosening a validator
elsewhere or by skipping the check.

Classified `reliability` rather than `architecture-fitness`: a
non-conforming schema leaves a compliant validator free to accept what
it should reject, panic, or read the file differently than
check-jsonschema reads it elsewhere in this repo. Nothing breaks the
moment a stray edit lands; the odds of an incident simply rise once that
binary ships. The check also compares one file against a fixed external
standard, not against another file in this repo.
