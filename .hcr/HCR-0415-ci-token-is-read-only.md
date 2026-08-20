---
id: HCR-0415
title: CI jobs must run with a read-only token
scope: repo
regulates: security
summary: "Every job in .github/workflows/ci.yml must declare permissions with contents read, rather than inheriting the repository default token scope."
state: enforced
enforced_by:
  - kind: config-assert
    file: ".github/workflows/ci.yml"
    path: "/jobs/ci/permissions/contents"
    op: equals
    value: "read"
---
Declare `permissions:` on every job in .github/workflows/ci.yml and give
it the narrowest scope that job needs; for `ci`, that is
`contents: read`. Do not fall back on the repository-wide default, which
is an organisation setting that can be widened without anyone touching
this repository. A job that needs to write something needs its own
narrow grant, not a broader default for all of them.

The assertion names the `ci` job in its path rather than a
workflow-level block, because a workflow-level grant is inherited
silently: a job added later would be covered by whatever the top of the
file said, without ever stating what it needs. A second job therefore
means a second assertion in this record.

Classified `security` rather than `compliance`: the concern is blast
radius, not an external requirement to document one. Every action in the
job runs with the token it is handed, so a compromised or typosquatted
action is bounded by that grant and by nothing else.
