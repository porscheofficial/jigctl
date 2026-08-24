package runner

import (
	"sort"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

type bindingCtx struct {
	target      hcr.Target
	binding     *hcr.ExecutableBinding
	originalIdx int
}

// EvaluatePlanWithProgress evaluates all executable bindings in a plan,
// deduplicating executions for bindings that share a non-empty Ref, and
// reports each group's start and settlement to p. A nil p evaluates silently.
func EvaluatePlanWithProgress(plan hcr.Plan, authorized bool, cadence CadenceSet, p Progress) []*Verdict {
	all := make([]bindingCtx, 0)
	idx := 0
	for _, t := range plan.Targets {
		for i := range t.Bindings {
			all = append(all, bindingCtx{
				target:      t,
				binding:     &t.Bindings[i],
				originalIdx: idx,
			})
			idx++
		}
	}

	if len(all) == 0 {
		return []*Verdict{}
	}

	verdicts := make([]*Verdict, len(all))
	allGroups := groupBindings(all)

	for _, groupIdxs := range allGroups {
		sortGroup(all, groupIdxs)
		members, eligibleMembers, maxResolvedTimeout := filterGroup(all, groupIdxs, cadence)

		notifyStart(p, eligibleMembers)
		evaluateGroup(plan, authorized, members, eligibleMembers, maxResolvedTimeout, verdicts)
		notifyDone(p, members, verdicts)
	}

	return verdicts
}

func groupBindings(all []bindingCtx) [][]int {
	var independentGroups [][]int
	refGroups := make(map[string][]int)

	for _, ctx := range all {
		if ctx.binding.Ref == "" {
			independentGroups = append(independentGroups, []int{ctx.originalIdx})
		} else {
			refGroups[ctx.binding.Ref] = append(refGroups[ctx.binding.Ref], ctx.originalIdx)
		}
	}

	refKeys := make([]string, 0, len(refGroups))
	for k := range refGroups {
		refKeys = append(refKeys, k)
	}
	sort.Strings(refKeys)

	allGroups := make([][]int, 0, len(independentGroups)+len(refKeys))
	allGroups = append(allGroups, independentGroups...)
	for _, k := range refKeys {
		allGroups = append(allGroups, refGroups[k])
	}
	return allGroups
}

type member struct {
	ctx        bindingCtx
	verdict    *Verdict
	eligible   bool
	resolvedTO time.Duration
}

func evaluateGroup(
	plan hcr.Plan,
	authorized bool,
	members []member,
	eligibleMembers []member,
	maxResolvedTimeout time.Duration,
	verdicts []*Verdict,
) {
	var sharedVerdict *Verdict
	if len(eligibleMembers) > 0 {
		sharedVerdict = executeGroup(plan, authorized, eligibleMembers[0].ctx, maxResolvedTimeout)
	}

	for _, m := range members {
		if m.eligible && sharedVerdict != nil {
			verdicts[m.ctx.originalIdx] = finalizeVerdict(&m, sharedVerdict, maxResolvedTimeout)
		} else {
			verdicts[m.ctx.originalIdx] = m.verdict
		}
	}
}

func sortGroup(all []bindingCtx, groupIdxs []int) {
	sort.Slice(groupIdxs, func(i, j int) bool {
		bi := all[groupIdxs[i]].binding
		bj := all[groupIdxs[j]].binding
		if bi.RecordPath != bj.RecordPath {
			return bi.RecordPath < bj.RecordPath
		}
		return bi.BindingIndex < bj.BindingIndex
	})
}

func filterGroup(
	all []bindingCtx,
	groupIdxs []int,
	cadence CadenceSet,
) (members, eligibleMembers []member, maxResolvedTimeout time.Duration) {
	for _, idx := range groupIdxs {
		ctx := all[idx]
		b := ctx.binding
		m := member{ctx: ctx}
		targetProv := TargetProvenance{Name: ctx.target.Kind, Path: ctx.target.Path}
		report := VerdictReport{
			Identity: BindingIdentity{RecordPath: b.RecordPath, BindingIndex: b.BindingIndex},
			Target:   targetProv,
			Kind:     b.Kind,
			Severity: b.Severity,
		}

		if v := Select(&report, b, cadence); v != nil {
			m.verdict = v
			members = append(members, m)
			continue
		}

		if b.Kind == "command" && (b.Pattern != "" || b.Select != "") {
			m.verdict = NewBlockedVerdict(&report, ReasonModifierUnimplemented)
			members = append(members, m)
			continue
		}

		m.eligible = true
		if b.Kind == "command" {
			timeoutSecs := 120
			if b.TimeoutSecs > 0 {
				timeoutSecs = b.TimeoutSecs
			}
			m.resolvedTO = time.Duration(timeoutSecs) * time.Second
			if m.resolvedTO > maxResolvedTimeout {
				maxResolvedTimeout = m.resolvedTO
			}
		}

		members = append(members, m)
		eligibleMembers = append(eligibleMembers, m)
	}
	return members, eligibleMembers, maxResolvedTimeout
}

func executeGroup(
	plan hcr.Plan,
	authorized bool,
	first bindingCtx,
	maxResolvedTimeout time.Duration,
) *Verdict {
	b := first.binding
	targetProv := TargetProvenance{Name: first.target.Kind, Path: first.target.Path}

	switch b.Kind {
	case "command":
		cloned := *b
		cloned.TimeoutSecs = int(maxResolvedTimeout.Seconds())
		return EvaluateCommandBinding(authorized, targetProv, &cloned, plan.Root)
	case "grep":
		return Grep(plan, first.target, *b)
	case "config-assert":
		return ConfigAssert(plan, first.target, *b)
	}
	return nil
}

func finalizeVerdict(m *member, sharedVerdict *Verdict, maxResolvedTimeout time.Duration) *Verdict {
	v := &Verdict{
		report:     sharedVerdict.report,
		completion: sharedVerdict.completion,
		reason:     sharedVerdict.reason,
	}
	v.report.Identity = BindingIdentity{
		RecordPath:   m.ctx.binding.RecordPath,
		BindingIndex: m.ctx.binding.BindingIndex,
	}
	v.report.Target = TargetProvenance{
		Name: m.ctx.target.Kind,
		Path: m.ctx.target.Path,
	}
	v.report.Severity = m.ctx.binding.Severity

	if m.ctx.binding.Kind == "command" {
		var declared *time.Duration
		if m.ctx.binding.TimeoutSecs > 0 {
			d := time.Duration(m.ctx.binding.TimeoutSecs) * time.Second
			declared = &d
		}
		eff := maxResolvedTimeout
		v.report.Timeouts = TimeoutRecord{
			Declared:  declared,
			Resolved:  m.resolvedTO,
			Effective: &eff,
		}
	}
	return v
}
