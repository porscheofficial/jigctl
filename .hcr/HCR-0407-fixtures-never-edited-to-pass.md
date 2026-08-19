---
id: HCR-0407
title: Fixture expectations are never edited to make jigctl pass
scope: repo
regulates: architecture-fitness
summary: "corpus/ is normative. When jigctl disagrees with a fixture's declared valid, at or covers, fix jigctl — never the expectation. A fixture diff shows that an expectation changed, never whether the change was justified."
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

Classified `architecture-fitness`: deciding whether an edit is a
legitimate specification change or a cover-up for a validator regression
means comparing two artifacts that live apart — the fixture, which is
the spec, and internal/hcr's actual behaviour on it. Neither read alone
shows anything wrong.
