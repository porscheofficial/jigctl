---
id: HCR-0301
title: Every release-affecting change updates the changelog
scope: repo
regulates: maintainability
summary: A pull request that changes runtime behaviour must add a corresponding entry under the Unreleased heading in the root changelog.
state: enforced
enforced_by:
  - kind: grep
    severity: blocking
    cadence: [on-change, ci]
    file: CHANGELOG.md
    require:
      - "## [Unreleased]"
---

If your change affects runtime behaviour, add a short bullet describing it
under the `## [Unreleased]` heading in `CHANGELOG.md` before requesting
review. Reviewers should bounce a PR that skips this back to the author
rather than adding the entry themselves.

<!-- jig:expect
valid: true
covers: [R-004, R-015]
deferred:
  - rule: R-101
    reason: requires-cross-file-resolution
  - rule: R-112
    reason: requires-filesystem
-->
