---
id: HCR-0305
title: Billing calculation changes get an agent review pass
scope: service
regulates: behaviour
summary: Any change to money-rounding or invoice-total logic in the billing service gets an automated agent review before merge, in addition to a human reviewer.
state: enforced
enforced_by:
  - kind: agent-review
    prompt: >-
      Check whether this change alters how invoice totals or currency
      rounding are computed. Flag any new rounding mode, precision change,
      or reordering of arithmetic operations for a human to confirm.
    grounding:
      - services/billing/README.md
---

An agent review here is advisory, not blocking: it exists to surface a
rounding or precision change to a human reviewer, not to gate the merge
on its own. Treat a flagged diff as a prompt to double-check the math,
not as a defect report to argue with.

