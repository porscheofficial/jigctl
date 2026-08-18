---
id: HCR-0407
title: Fixture expectations are never edited to make jigctl pass
scope: repo
regulates: architecture-fitness
summary: "Whether an edit to a fixture's valid, at, or covers expectation is a legitimate specification change or a quiet cover-up for a validator regression is invisible from the fixture's own diff — the file by itself only shows that an expectation changed, never why. Answering that requires comparing two artifacts that live apart entirely — the fixture, which is the spec, and internal/hcr's actual behavior on it, which is the implementation — and asking whether the implementation was fixed to match a still-correct expectation, or the expectation was loosened to match a broken implementation. That comparison is exactly the property docs/concepts.md assigns to architecture-fitness: nothing wrong is visible in either file read alone, and the defect, if any, exists only in the relationship between them, the same shape as a circular import."
state: enforced
enforced_by:
  - kind: inferential
---
corpus/ is normative: fixtures are the spec, not a description of
whatever the tool currently does. If jigctl's output disagrees with a
fixture's declared valid, at, or covers, treat that as a bug in jigctl
until you have concrete evidence otherwise, and fix the implementation.
Do not edit the fixture's expectation to make the disagreement go away —
that silently redefines the spec to match a possibly-broken tool. A hash
of every expectation block is checked so a genuine edit is at least a
visible, deliberate act in review; that hash can show that an
expectation changed, never whether the change was justified, which is
why enforcement here is a human editorial judgement call and not
something a command can run.
