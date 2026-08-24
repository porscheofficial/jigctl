package runner

import (
	"github.com/porscheofficial/jigctl/internal/hcr"
)

// Select determines whether a binding should execute in the current invocation,
// enforcing gating on record state, cadence rules, and executable kinds.
//
// Cadence logic: Selection checks the binding's cadence against the requested set.
// If it matches, the binding executes (returns nil). If it misses and the invoker
// explicitly specified the cadence, it yields ReasonCadenceDeselected. If it misses
// under the implicit default cadence, it yields ReasonCadenceExcluded.
func Select(report *VerdictReport, binding *hcr.ExecutableBinding, requested CadenceSet) *Verdict {
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

	if requested.Selects(binding) {
		return nil
	}
	if requested.Explicit() {
		return NewNotAttemptedVerdict(report, ReasonCadenceDeselected)
	}
	return NewNotAttemptedVerdict(report, ReasonCadenceExcluded)
}
