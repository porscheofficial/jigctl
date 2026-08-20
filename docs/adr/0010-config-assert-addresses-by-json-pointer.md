---
status: proposed
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0010: Config assertions use JSON Pointer addresses

## Context

The normative corpus already expresses every `config-assert` path in
pointer syntax. None uses dotted addressing. The dotted path in
`.hcr/HCR-0409-adr-rationale-prefix-is-mapped.md` is therefore the outlier,
and correcting it is a correction rather than an adoption cost. This follows
the repository rule that fixtures are the specification.

The schema deliberately leaves the path expression grammar to the runtime.
A TOML key containing a dot is one pointer segment and needs no escaping,
whereas a dotted grammar has no specified answer for that key.

The schema also leaves comparison semantics open. Its expected value can be
a string, number, or boolean, while decoded configuration data may contain
other types. In particular, the TOML decoder represents a datetime as a
`time.Time`, which the declared expected-value types cannot express.

Some authored paths occur in service-scoped records. Their scope identifies
where a rule applies, but does not establish a filesystem base. The service
fixture's API record names `scripts/check-api-contracts.sh`; that file is at
the tree root, and no corresponding service-local scripts directory exists.

## Decision

`config-assert.path` is an RFC 6901 JSON Pointer. Escaping and lookup follow
that grammar without a dotted-path fallback.

`equals` compares values without coercion. Strings equal strings. Booleans
equal booleans. Numbers compare by numeric value. Operands from different
types do not equal, and the mismatch is surfaced as a diagnostic.

`gte` and `lte` accept only numeric actual and expected values. They compare
inclusively. A nonnumeric operand is a diagnostic and is never silently
excluded.

`matches` accepts only string actual and expected values. The expected value
is an unanchored regular expression. Anchoring must be explicit. A nonstring
operand or invalid expression is a diagnostic.

`absent` passes only when the target file was read successfully and the
pointer does not resolve. A missing file fails evaluation of the declared
subject and produces a diagnostic; it does not pass by technicality.

A TOML datetime is unsupported as a comparison operand. Encountering one at
the addressed value produces a diagnostic rather than string conversion or
silent exclusion.

YAML, TOML, and JSON are in scope. XML is deferred because attributes,
repeated siblings, text nodes, and namespaces have no inherent RFC 6901
representation. Properties files are deferred because they are flat mappings
from strings to strings and dots are part of their keys. Neither deferred
format is attested by a record or fixture.

Every authored path resolves against `Plan.Root`, the absolute tree root.
This applies to path-shaped `command.run`, `grep.file`,
`config-assert.file`, and path-shaped `exceptions[].scope`, in repo-scoped
and service-scoped records alike. There is one resolution base and no
per-target base.

This base preserves the rule already fixed for R-104. It also matches the
service fixture whose root script is named by a service-scoped record, and
the schema's statement that commands execute from the repository root.
Finally, the schema defines service scope as applicability to a service,
not as a path root.

For `command.run`, confinement applies only to the first token when that
token contains `/`, using the same path-shaped test as R-104. A bare first
token is a `PATH` lookup, is not an authored path, and is deliberately not
confined. Later tokens are not inspected as paths.

## Consequences

Configuration addressing has one grammar across the supported data formats.
Keys containing dots remain unambiguous. Pointer escaping handles slashes
and tildes without adding format-specific syntax.

Comparison failures remain visible. Type mismatches, unsupported datetimes,
invalid expressions, and missing target files cannot turn into passing or
partially evaluated assertions.

Supporting XML or properties files later requires a separate mapping
decision before either format can participate in pointer lookup.

Tree-root resolution keeps record scope independent from filesystem
interpretation. Path confinement can therefore use the same base for every
binding and exception scope, while ordinary executable lookup remains normal
`PATH` behavior.
