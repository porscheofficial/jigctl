package runner

import (
	"os"
	"testing"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestRefGroup_CadenceExcluded(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(0, 10, []string{"production"}, "", "", "", scriptPath),
		makeTestBinding(1, 20, nil, "", "", "", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 1)

	if verdicts[0].completion != CompletionNotAttempted {
		t.Errorf("expected first to be not attempted, got %v", verdicts[0].completion)
	}
	if verdicts[0].reason != Reason(ReasonCadenceExcluded) {
		t.Errorf("expected ReasonCadenceExcluded")
	}
	if verdicts[1].completion != CompletionCompleted {
		t.Errorf("expected second to be completed, got %v", verdicts[1].completion)
	}
	to1 := verdicts[1].report.Timeouts
	if to1.Effective == nil || *to1.Effective != 20*time.Second {
		t.Errorf("expected effective 20s, got %v", to1.Effective)
	}
}

type recordingProgress struct {
	starts []BindingIdentity
	dones  []BindingIdentity
}

func (r *recordingProgress) Start(id BindingIdentity) {
	r.starts = append(r.starts, id)
}

func (r *recordingProgress) Done(id BindingIdentity, v *Verdict) {
	r.dones = append(r.dones, id)
}

func TestEvaluatePlanWithProgress_Cadence(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)

	// (a) an independent deselected binding — zero starts, one done
	t.Run("independent_deselected", func(t *testing.T) {
		b := makeTestBinding(0, 10, []string{"ci"}, "", "", "", scriptPath)
		b.Ref = ""

		plan := hcr.Plan{Root: tmpRoot, Targets: []hcr.Target{{Kind: "repo", Path: "", Bindings: []hcr.ExecutableBinding{b}}}}
		prog := &recordingProgress{}
		cadence, err := ParseCadenceSet("scheduled", true)
		if err != nil {
			t.Fatal(err)
		}
		_ = cadence
		os.Remove(counterPath)
		EvaluatePlanWithProgress(plan, true, cadence, prog)

		if len(prog.starts) != 0 {
			t.Errorf("expected 0 starts, got %d", len(prog.starts))
		}
		if len(prog.dones) != 1 {
			t.Errorf("expected 1 done, got %d", len(prog.dones))
		}
	})

	// (b) a shared-ref group with one selected + one deselected member
	t.Run("shared_mixed", func(t *testing.T) {
		b0 := makeTestBinding(0, 10, []string{"scheduled"}, "", "", "", scriptPath)
		b1 := makeTestBinding(1, 10, []string{"ci"}, "", "", "", scriptPath)

		plan := hcr.Plan{Root: tmpRoot, Targets: []hcr.Target{{Kind: "repo", Path: "", Bindings: []hcr.ExecutableBinding{b0, b1}}}}
		prog := &recordingProgress{}
		cadence, err := ParseCadenceSet("scheduled", true)
		if err != nil {
			t.Fatal(err)
		}
		_ = cadence
		os.Remove(counterPath)
		verdicts := EvaluatePlanWithProgress(plan, true, cadence, prog)

		if len(prog.starts) != 1 {
			t.Errorf("expected 1 start, got %d", len(prog.starts))
		}
		if len(prog.dones) != 2 {
			t.Errorf("expected 2 dones, got %d", len(prog.dones))
		}
		checkExecutionCount(t, counterPath, 1)

		if verdicts[1].reason != Reason(ReasonCadenceExcluded) && verdicts[1].reason != Reason(ReasonCadenceDeselected) {
			if verdicts[1].reason != Reason(ReasonCadenceDeselected) {
				t.Errorf("expected reason CadenceDeselected, got %v", verdicts[1].reason)
			}
		}
	})

	// (c) a shared-ref group wholly deselected
	t.Run("shared_deselected", func(t *testing.T) {
		b0 := makeTestBinding(0, 10, []string{"ci"}, "", "", "", scriptPath)
		b1 := makeTestBinding(1, 10, []string{"ci"}, "", "", "", scriptPath)

		plan := hcr.Plan{Root: tmpRoot, Targets: []hcr.Target{{Kind: "repo", Path: "", Bindings: []hcr.ExecutableBinding{b0, b1}}}}
		prog := &recordingProgress{}
		cadence, err := ParseCadenceSet("scheduled", true)
		if err != nil {
			t.Fatal(err)
		}
		_ = cadence
		os.Remove(counterPath)
		EvaluatePlanWithProgress(plan, true, cadence, prog)

		if len(prog.starts) != 0 {
			t.Errorf("expected 0 starts, got %d", len(prog.starts))
		}
		if len(prog.dones) != 2 {
			t.Errorf("expected 2 dones, got %d", len(prog.dones))
		}
		checkExecutionCount(t, counterPath, 0)
	})

	// (d) under DefaultCadenceSet(), inferential HCR-0407 record receives zero starts and one done
	t.Run("inferential_excluded", func(t *testing.T) {
		b := makeTestBinding(0, 10, []string{"ci"}, "", "", "", scriptPath)
		b.Kind = "inferential"
		b.RecordPath = "HCR-0407-format.md"
		b.Ref = ""

		plan := hcr.Plan{Root: tmpRoot, Targets: []hcr.Target{{Kind: "repo", Path: "", Bindings: []hcr.ExecutableBinding{b}}}}
		prog := &recordingProgress{}
		os.Remove(counterPath)
		EvaluatePlanWithProgress(plan, true, DefaultCadenceSet(), prog)

		if len(prog.starts) != 0 {
			t.Errorf("expected 0 starts, got %d", len(prog.starts))
		}
		if len(prog.dones) != 1 {
			t.Errorf("expected 1 done, got %d", len(prog.dones))
		}
	})
}
