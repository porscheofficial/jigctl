---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0009: Pattern Syntax Is Declared per Surface

## Context

Several binding fields describe text to find, but the schema does not give
them one shared syntax. The `require` and `forbid` fields describe patterns
in files. A command's `pattern` is explicitly a regular expression over its
combined standard output and standard error. The `matches` operation also
needs a syntax for comparing a configuration value.

Treating every such field as a regular expression contradicts attested
values. A Go measurement of `regexp.Compile("os.Exit(")` returns
`error parsing regexp: missing closing ): os.Exit(`. The same compilation
fails for `pdb.set_trace(`, `breakpoint(`, and `print(`. Those values occur
in repository records, and the latter three occur in the normative corpus.
The repository rules state that its fixtures are the specification.

Pattern syntax therefore has to be declared per surface rather than
inferred from the word "pattern". These decisions change together as one
pattern-language contract. ADR-0001's calendar test consequently places
them in one record rather than three records with artificial lifecycles.

## Decision

The elements of `require` and `forbid` are literal substrings. A required
substring must occur in at least one matched file. A forbidden substring
must occur in no matched file. Regular-expression metacharacters have no
special meaning on this surface.

The command `pattern` field is out of scope for M2 and is not evaluated.
Supporting it would force unbounded command-output capture, require a rule
for interleaving standard output and standard error, and add a second
verdict axis that can disagree with the exit code. No precedence rule for
that disagreement exists, and no record currently uses the field.

For `config-assert`, `op: matches` uses a Go regular expression. Matching is
unanchored: a successful match may cover any substring of the compared
string. Authors who need whole-string matching must declare anchors in the
expression. Applying this operation to a non-string value is a type error,
not a string conversion.

A future schema decision may add separate `require_regex` and
`forbid_regex` keys as an additive escape hatch. Those keys do not exist
today. The reverse change, narrowing regex fields after records depend on
them, would not preserve existing meaning.

## Consequences

Existing grep values containing unmatched parentheses remain valid and are
searched exactly as written. Implementations can use bounded, streaming
literal search rather than compiling record values or loading whole files.

Configuration assertions retain an explicit regular-expression facility.
Their unanchored default follows search semantics while leaving exact
matching available through authored anchors.

Command output assertions remain honest about being unimplemented instead
of silently ignoring the modifier or inventing capture and precedence
semantics during execution. Adding output-pattern support requires another
decision that settles those semantics before implementation.

Literal grep syntax cannot express alternatives, character classes, or
other regular-expression operations. Adding that capability requires new
schema keys and a coordinated schema and decision change.
