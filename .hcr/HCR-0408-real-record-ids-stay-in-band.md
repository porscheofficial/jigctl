---
id: HCR-0408
title: A real record sits directly in .hcr/ and uses an 04xx id
scope: repo
regulates: maintainability
summary: "Every real record must sit directly in .hcr/ with an HCR-04xx id, so none is hidden from discovery or mistaken for a corpus fixture."
state: enforced
rationale: ADR-0002
enforced_by:
  - kind: command
    run: ".hcr/checks/check-record-ids.py"
---
Run `.hcr/checks/check-record-ids.py` (or `mise run check`) after adding
or renumbering a record under .hcr/. Choose the next free HCR-04xx id. Do
not allocate a fixture id to a real record or renumber corpus fixtures to
make room; each population owns its existing band.

The same script refuses a record nested below .hcr/, next to the checks
or anywhere else. `indexTree` globs `<root>/.hcr/*.md` and nothing
deeper, so a nested file is not a weak record or an invalid one — it is
not a record at all, and jigctl has nothing to report about it.

This binding names the script by path rather than routing through a mise
task, because R-104 resolves a path-shaped command and ignores everything
else. The two forms also fail differently: a missing script is reported
`blocked-unchecked`, while a renamed mise task starts fine, exits
non-zero and is reported as a violation of the rule itself —
infrastructure breakage wearing a constraint's name. Route a binding
through mise only when the task wraps something with no single path, the
way `lint` chains three separate tools; a lone script has a path, so it
uses it.

Classified `maintainability` for both halves. An out-of-band id does not
change jigctl's validation result, but it forces every later reader to
tell real records from fixtures by path instead of by id. A nested file
is not a rule that stopped working either — it never was one; what it
costs is the reader who believes it is.
