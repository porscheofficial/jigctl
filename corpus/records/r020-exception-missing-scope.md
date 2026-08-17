---
id: HCR-2013
title: Require structured logging in request handlers
scope: repo
regulates: maintainability
summary: Request handlers must emit structured log entries rather than free-form text.
state: enforced
enforced_by:
  - kind: command
    run: tools/check-structured-logging.sh
exceptions:
  - reason: Legacy handler predates the structured logging library and is scheduled for rewrite.
    until: "2026-09-30"
---

Use the shared logging helper so every entry carries a consistent set of fields.
Avoid string-concatenating variables into the log message body.
Structured fields make the logs searchable during an incident; free text does not.

<!-- jig:expect
valid: false
covers: [R-020]
diagnostics:
  - rule: R-020
    at: /exceptions/0
-->
