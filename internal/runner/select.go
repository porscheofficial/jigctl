package runner

import (
	"github.com/porscheofficial/jigctl/internal/hcr"
)

// Select determines whether a binding should execute in the current invocation,
// enforcing gating on record state, cadence rules, and executable kinds.
//
// Cadence logic: A bare jigctl run invocation (with no --cadence flag)
// is executed identically by local developers and CI. Thus, a binding
// is selected if its resolved Cadence contains either "on-change" or "ci".
// Bindings with only "scheduled", "production", or empty cadence are excluded.
// If selected for execution, Select returns nil.
func Select(report *VerdictReport, binding *hcr.ExecutableBinding) *Verdict {
	if binding.State == "draft" {
		return NewNotAttemptedVerdict(report, ReasonRecordDraft)
	}

	if binding.State == "deprecated" {
		return NewNotAttemptedVerdict(report, ReasonRecordDeprecated)
	}

	switch binding.Kind {
	case "external", "agent-review", "inferential":
		return NewNotAttemptedVerdict(report, ReasonKindNotExecutable)
	}

	hasCadence := false
	for _, c := range binding.Cadence {
		if c == "on-change" || c == "ci" {
			hasCadence = true
			break
		}
	}
	if !hasCadence {
		return NewNotAttemptedVerdict(report, ReasonCadenceExcluded)
	}

	return nil
}
