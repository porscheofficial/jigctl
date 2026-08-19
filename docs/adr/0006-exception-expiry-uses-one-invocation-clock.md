---
status: accepted
date: 2026-08-19
decision-makers: [Patrice Bouillet]
---

# ADR-0006: Exception expiry uses one invocation clock

## Context

R-107 says that an `exceptions[].until` date in the past re-fires the
constraint it waived. The schema states the purpose directly: after the
date passes the rule fires again, so that a temporary waiver cannot
quietly become permanent.

Every rule enforced so far reads only the tree. Given the same bytes they
return the same diagnostics on any machine on any day, and that is what
lets a corpus fixture stand as a stable assertion. R-107 cannot work that
way. Its verdict depends on when it is asked, and that single fact is the
whole of what this decision has to settle.

The repository already contains the hazard. A record in the corpus carries
an `until` date that was in the future when the record was written and is
in the past now. Nothing broke, because that corpus asserts record shape
only and never runs tree rules. But it is a worked example of what happens
to any assertion that a waiver has not yet expired: it holds, until it
silently stops holding.

## Decision

An expired `until` is reported as a diagnostic on the offending
exception's `until` pointer. Reporting is the whole of the rule at this
milestone. Whether a runtime that evaluates constraints should also treat
an expired waiver as absent belongs to the executor and is not decided
here.

Expiry is strictly after the date. A waiver dated today is still in force
today and fires tomorrow, which is what the schema's wording says.

Dates are compared in UTC. A local comparison would let the same tree
produce different verdicts for a developer in one timezone and a runner in
another on the day a waiver turns over. That is the same class of problem
as filename case, which R-112 settled by comparing bytes rather than
asking the filesystem what it thinks two names mean.

The current date is read once per invocation and passed to the rule as an
ordinary argument. It is not read inside the rule, and it is not a package
variable that tests replace. Reading it per record would let a long run
straddle midnight and return two different answers for two identical
records. A package variable would be shared mutable state, which this test
suite runs shuffled specifically to catch.

The expired case gets a corpus fixture, because a date in the past stays
in the past. The unexpired case never does. It is asserted by a unit test
that supplies its own date, so that the assertion is about the comparison
and not about when the test happens to run.

## Consequences

R-107 is the first rule whose output is not a function of the tree alone.
The same commit validates clean one day and reports a finding the next.
That is the rule working rather than failing, but it does mean a green run
stops being evidence that the next run will be green, and that a bisect
across a date boundary will not reproduce.

Threading the clock as an argument keeps the dependency visible in every
signature it reaches. A rule that takes a date is time-dependent and a
rule that does not is not, and neither fact has to be remembered.

A waiver written with a far-future date will pass while defeating the
purpose of the field. Neither the schema nor this rule can prevent that.
It is a review concern, and naming it here is the only defence this
milestone offers.
