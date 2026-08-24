package runner

import "github.com/porscheofficial/jigctl/internal/hcr"

// Progress observes a run as it happens. A run notifies it when a group of
// bindings starts executing and again when each of them has settled, which is
// what lets a live view show the work in flight instead of only its result.
//
// Bindings that share a Ref execute once and settle together, so a single
// Start can be followed by several Done calls.
type Progress interface {
	// Start reports that the binding is now executing.
	Start(id BindingIdentity)
	// Done reports the binding's settled verdict, which is nil when no
	// verdict could be produced for it.
	Done(id BindingIdentity, v *Verdict)
}

// EvaluatePlan evaluates all executable bindings in a plan, deduplicating
// executions for bindings that share a non-empty Ref.
func EvaluatePlan(plan hcr.Plan, authorized bool, cadence CadenceSet) []*Verdict {
	return EvaluatePlanWithProgress(plan, authorized, cadence, nil)
}

func notifyStart(p Progress, eligibleMembers []member) {
	if p == nil {
		return
	}
	for _, m := range eligibleMembers {
		b := m.ctx.binding
		p.Start(BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex})
	}
}

func notifyDone(p Progress, members []member, verdicts []*Verdict) {
	if p == nil {
		return
	}
	for _, m := range members {
		b := m.ctx.binding
		p.Done(BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex}, verdicts[m.ctx.originalIdx])
	}
}
