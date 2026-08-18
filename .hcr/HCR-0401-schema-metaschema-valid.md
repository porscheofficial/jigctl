---
id: HCR-0401
title: Schema must validate against its own metaschema
scope: repo
regulates: reliability
summary: "schema/hcr.schema.json is embedded directly into the jigctl binary via go:embed and compiled once at build time — nothing re-validates it at runtime. If it silently drifts out of conformance with the JSON Schema 2020-12 metaschema, that defect ships inside every copy of the tool until the next release, and a non-conforming schema leaves a compliant validator free to do anything with it — accept what it should reject, panic, or read differently than check-jsonschema reads the very same file elsewhere in this repo. Nothing is broken the moment a stray edit violates the metaschema; the odds of a production incident simply rise the instant that binary ships, which is exactly the reliability question docs/concepts.md asks, not an architecture-fitness one — the check compares this one file against a fixed external standard, not against another file in this repo."
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
