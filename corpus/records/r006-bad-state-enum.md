---
id: HCR-2010
title: Require secrets to be loaded from the secret manager
scope: repo
regulates: security
summary: Application code must load secrets from the secret manager, never from plain environment files.
state: active
enforced_by:
  - kind: command
    run: tools/check-secret-loading.sh
---

Load credentials and tokens through the secret manager client at startup.
Never commit a `.env` file containing a real secret value, even temporarily.
If a secret leaks, rotate it immediately and open an incident record.

<!-- jig:expect
valid: false
covers: [R-006]
diagnostics:
  - rule: R-006
    at: /state
-->
