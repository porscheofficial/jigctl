package runner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func makeTestBinding(idx, timeout int, cadence []string, state, pattern, selectStr, scriptPath string) hcr.ExecutableBinding {
	if len(cadence) == 0 {
		cadence = []string{"ci"}
	}
	if state == "" {
		state = "enforced"
	}
	return hcr.ExecutableBinding{
		RecordPath:   "record.md",
		BindingIndex: idx,
		Kind:         "command",
		State:        state,
		Severity:     "blocking",
		Cadence:      cadence,
		Ref:          "group1",
		Run:          scriptPath,
		TimeoutSecs:  timeout,
		Pattern:      pattern,
		Select:       selectStr,
	}
}

func setupTestEnvironment(t *testing.T) (tmpRoot, scriptPath, counterPath string) {
	t.Helper()
	tmpRoot = t.TempDir()

	scriptPath = filepath.Join(tmpRoot, "script.sh")
	counterPath = filepath.Join(tmpRoot, "counter.txt")
	scriptBody := "#!/bin/sh\n" +
		"val=$(cat counter.txt 2>/dev/null || echo 0)\n" +
		"val=$((val+1))\n" +
		"echo $val > counter.txt\n" +
		"echo \"executed\"\n"
	err := os.WriteFile(scriptPath, []byte(scriptBody), 0o755)
	if err != nil {
		t.Fatalf("write script: %v", err)
	}

	return tmpRoot, scriptPath, counterPath
}

func checkExecutionCount(t *testing.T, counterPath string, expected int) {
	t.Helper()
	b, err := os.ReadFile(counterPath)
	if err != nil && expected > 0 {
		t.Fatalf("failed to read counter: %v", err)
	}
	count := 0
	if err == nil && len(b) > 0 {
		count = int(b[0] - '0')
	}
	if count != expected {
		t.Fatalf("expected %d executions, got %d", expected, count)
	}
}

func runTestPlan(tmpRoot, counterPath string, bindings []hcr.ExecutableBinding) []*Verdict {
	os.Remove(counterPath)
	plan := hcr.Plan{
		Root: tmpRoot,
		Targets: []hcr.Target{
			{Kind: "repo", Path: "", Bindings: bindings},
		},
	}
	return EvaluatePlan(plan, true)
}

func TestRefGroup_MixedTimeouts(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(1, 5, nil, "", "", "", scriptPath),
		makeTestBinding(0, 30, nil, "", "", "", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 1)

	if len(verdicts) != 2 {
		t.Fatalf("expected 2 verdicts, got %d", len(verdicts))
	}
	if verdicts[0].report.Identity.BindingIndex != 1 {
		t.Errorf("expected original position 0 to be binding index 1")
	}
	if verdicts[1].report.Identity.BindingIndex != 0 {
		t.Errorf("expected original position 1 to be binding index 0")
	}

	to0 := verdicts[0].report.Timeouts
	if to0.Declared == nil || *to0.Declared != 5*time.Second {
		t.Errorf("expected declared 5s, got %v", to0.Declared)
	}
	if to0.Resolved != 5*time.Second {
		t.Errorf("expected resolved 5s, got %v", to0.Resolved)
	}
	if to0.Effective == nil || *to0.Effective != 30*time.Second {
		t.Errorf("expected effective 30s, got %v", to0.Effective)
	}

	to1 := verdicts[1].report.Timeouts
	if to1.Declared == nil || *to1.Declared != 30*time.Second {
		t.Errorf("expected declared 30s, got %v", to1.Declared)
	}
	if to1.Resolved != 30*time.Second {
		t.Errorf("expected resolved 30s, got %v", to1.Resolved)
	}
	if to1.Effective == nil || *to1.Effective != 30*time.Second {
		t.Errorf("expected effective 30s, got %v", to1.Effective)
	}
}

func TestRefGroup_AbsentTimeout(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(0, 1, nil, "", "", "", scriptPath),
		makeTestBinding(1, 0, nil, "", "", "", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 1)

	to0 := verdicts[0].report.Timeouts
	if to0.Effective == nil || *to0.Effective != 120*time.Second {
		t.Errorf("expected effective 120s, got %v", to0.Effective)
	}
	to1 := verdicts[1].report.Timeouts
	if to1.Declared != nil {
		t.Errorf("expected no declared timeout, got %v", to1.Declared)
	}
	if to1.Resolved != 120*time.Second {
		t.Errorf("expected resolved 120s, got %v", to1.Resolved)
	}
	if to1.Effective == nil || *to1.Effective != 120*time.Second {
		t.Errorf("expected effective 120s, got %v", to1.Effective)
	}
}

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

func TestRefGroup_ModifierBlocked(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(0, 100, nil, "", "pattern", "", scriptPath),
		makeTestBinding(1, 10, nil, "", "", "", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 1)

	if verdicts[0].completion != CompletionBlocked {
		t.Errorf("expected first to be blocked, got %v", verdicts[0].completion)
	}
	if verdicts[0].reason != Reason(ReasonModifierUnimplemented) {
		t.Errorf("expected ReasonModifierUnimplemented")
	}
	if verdicts[1].completion != CompletionCompleted {
		t.Errorf("expected second to be completed, got %v", verdicts[1].completion)
	}
	to1 := verdicts[1].report.Timeouts
	if to1.Effective == nil || *to1.Effective != 10*time.Second {
		t.Errorf("expected effective 10s, got %v", to1.Effective)
	}
}

func TestRefGroup_AllBlocked(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(0, 5, nil, "", "pattern", "", scriptPath),
		makeTestBinding(1, 10, nil, "", "", "select", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 0)

	for i, v := range verdicts {
		if v.completion != CompletionBlocked {
			t.Errorf("expected %d to be blocked, got %v", i, v.completion)
		}
		if v.reason != Reason(ReasonModifierUnimplemented) {
			t.Errorf("expected ReasonModifierUnimplemented")
		}
	}
}

func TestRefGroup_NoneSurvives(t *testing.T) {
	tmpRoot, scriptPath, counterPath := setupTestEnvironment(t)
	bindings := []hcr.ExecutableBinding{
		makeTestBinding(0, 5, nil, "draft", "", "", scriptPath),
		makeTestBinding(1, 10, nil, "deprecated", "", "", scriptPath),
	}
	verdicts := runTestPlan(tmpRoot, counterPath, bindings)
	checkExecutionCount(t, counterPath, 0)

	if verdicts[0].completion != CompletionNotAttempted {
		t.Errorf("expected first not attempted")
	}
	if verdicts[1].completion != CompletionNotAttempted {
		t.Errorf("expected second not attempted")
	}
}
