package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func TestCommandAuthorization(t *testing.T) {
	t.Run("flag overrides all", func(t *testing.T) {
		auth := ResolveAuthorization(true, "", nil)
		if !auth {
			t.Fatal("expected auth true, got false")
		}
		auth = ResolveAuthorization(true, "0", func(s string) (string, error) { return "no", nil })
		if !auth {
			t.Fatal("expected auth true, got false")
		}
	})

	t.Run("env exact 1 authorizes", func(t *testing.T) {
		auth := ResolveAuthorization(false, "1", nil)
		if !auth {
			t.Fatal("expected auth true, got false")
		}
	})

	t.Run("env not 1 denies", func(t *testing.T) {
		auth := ResolveAuthorization(false, "yes", nil)
		if auth {
			t.Fatal("expected auth false, got true")
		}
	})

	t.Run("interactive prompt yes authorizes", func(t *testing.T) {
		auth := ResolveAuthorization(false, "", func(s string) (string, error) { return "yes", nil })
		if !auth {
			t.Fatal("expected auth true, got false")
		}
	})

	t.Run("interactive prompt no denies", func(t *testing.T) {
		auth := ResolveAuthorization(false, "", func(s string) (string, error) { return "no", nil })
		if auth {
			t.Fatal("expected auth false, got true")
		}
	})

	t.Run("no flag, no env, no prompt denies", func(t *testing.T) {
		auth := ResolveAuthorization(false, "", nil)
		if auth {
			t.Fatal("expected auth false, got true")
		}
	})
}

func TestCommandBinding(t *testing.T) {
	tmpRoot := t.TempDir()

	runCheck := func(t *testing.T, runStr string, auth bool) *Verdict {
		t.Helper()
		binding := &hcr.ExecutableBinding{
			RecordPath:   "test.md",
			BindingIndex: 0,
			Kind:         "repo",
			Severity:     "HIGH",
			Run:          runStr,
			TimeoutSecs:  2,
		}
		target := TargetProvenance{Name: "repo", Path: ""}
		return EvaluateCommandBinding(auth, target, binding, tmpRoot)
	}

	testCommandBindingSuccessScenarios(t, runCheck)
	testCommandBindingFailureScenarios(t, runCheck, tmpRoot)
	testCommandBindingExecutionProofs(t, runCheck, tmpRoot)
	testCommandBindingExecutionProofsTwo(t, runCheck, tmpRoot)
	testCommandBindingCwdProof(t)
}

func testCommandBindingSuccessScenarios(t *testing.T, runCheck func(*testing.T, string, bool) *Verdict) {
	t.Run("exit 0 passes", func(t *testing.T) {
		verdict := runCheck(t, "true", true)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completion %v, got %v", CompletionCompleted, verdict.Completion())
		}
		proj, _ := verdict.Projection()
		if proj != ProjectionPass {
			t.Fatalf("expected projection %v, got %v", ProjectionPass, proj)
		}
	})

	t.Run("exit 1 is a violation", func(t *testing.T) {
		verdict := runCheck(t, "false", true)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completion %v, got %v", CompletionCompleted, verdict.Completion())
		}
		proj, _ := verdict.Projection()
		if proj != ProjectionViolation {
			t.Fatalf("expected projection %v, got %v", ProjectionViolation, proj)
		}
		if len(verdict.Report().Findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(verdict.Report().Findings))
		}
	})
}

func testCommandBindingFailureScenarios(t *testing.T, runCheck func(*testing.T, string, bool) *Verdict, tmpRoot string) {
	t.Run("unauthorized blocks execution", func(t *testing.T) {
		verdict := runCheck(t, "true", false)
		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected completion %v, got %v", CompletionBlocked, verdict.Completion())
		}
		if verdict.Reason() != Reason(ReasonAuthorizationDenied) {
			t.Fatalf("expected reason %v, got %v", Reason(ReasonAuthorizationDenied), verdict.Reason())
		}
	})

	t.Run("nonexistent binary", func(t *testing.T) {
		verdict := runCheck(t, "does-not-exist-xyz123", true)
		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected completion %v, got %v", CompletionBlocked, verdict.Completion())
		}
		if verdict.Reason() != Reason(ReasonExecutableMissing) {
			t.Fatalf("expected reason %v, got %v", Reason(ReasonExecutableMissing), verdict.Reason())
		}
	})

	t.Run("modifiers unimplemented", func(t *testing.T) {
		binding := &hcr.ExecutableBinding{
			Run:     "true",
			Pattern: "foo",
		}
		verdict := EvaluateCommandBinding(true, TargetProvenance{}, binding, tmpRoot)
		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected completion %v, got %v", CompletionBlocked, verdict.Completion())
		}
		if verdict.Reason() != Reason(ReasonModifierUnimplemented) {
			t.Fatalf("expected reason %v, got %v", Reason(ReasonModifierUnimplemented), verdict.Reason())
		}
	})

	t.Run("argv invalid (empty)", func(t *testing.T) {
		verdict := runCheck(t, "   ", true)
		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected completion %v, got %v", CompletionBlocked, verdict.Completion())
		}
		if verdict.Reason() != Reason(ReasonArgvInvalid) {
			t.Fatalf("expected reason %v, got %v", Reason(ReasonArgvInvalid), verdict.Reason())
		}
	})
}

func testCommandBindingExecutionProofs(t *testing.T, runCheck func(*testing.T, string, bool) *Verdict, tmpRoot string) {
	t.Run("stdin proof", func(t *testing.T) {
		scriptPath := filepath.Join(tmpRoot, "read.sh")
		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nread x\n"), 0o755)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		verdict := runCheck(t, scriptPath, true)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completion %v, got %v", CompletionCompleted, verdict.Completion())
		}
		proj, _ := verdict.Projection()
		if proj != ProjectionViolation {
			t.Fatalf("expected projection %v, got %v", ProjectionViolation, proj)
		}
	})

	t.Run("timeout kills process group", func(t *testing.T) {
		scriptPath := filepath.Join(tmpRoot, "sleep_trap.sh")
		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\ntrap 'echo SIGTERM' TERM\nsleep 10\n"), 0o755)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		start := time.Now()
		verdict := runCheck(t, scriptPath, true)
		elapsed := time.Since(start)

		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected completion %v, got %v", CompletionBlocked, verdict.Completion())
		}
		if verdict.Reason() != Reason(ReasonTimeout) {
			t.Fatalf("expected reason %v, got %v", Reason(ReasonTimeout), verdict.Reason())
		}
		if elapsed >= 4*time.Second {
			t.Fatalf("expected elapsed < 4s, got %v", elapsed)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "pgrep", "-f", "sleep_trap.sh")
		out, err := cmd.Output()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if len(out) != 0 {
			t.Fatalf("expected empty output, got %q", string(out))
		}
	})
}

func testCommandBindingExecutionProofsTwo(t *testing.T, runCheck func(*testing.T, string, bool) *Verdict, tmpRoot string) {
	t.Run("shell-free proof", func(t *testing.T) {
		scriptPath := filepath.Join(tmpRoot, "echo_args.sh")
		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho \"$*\"\n"), 0o755)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		verdict := runCheck(t, scriptPath+" a && echo b", true)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completion %v, got %v", CompletionCompleted, verdict.Completion())
		}
		if verdict.Report().Execution == nil {
			t.Fatalf("expected execution not nil")
		}

		argv := verdict.Report().Execution.Argv
		if len(argv) != 5 {
			t.Fatalf("expected 5 args, got %d", len(argv))
		}
		if argv[1] != "a" || argv[2] != "&&" || argv[3] != "echo" || argv[4] != "b" {
			t.Fatalf("unexpected argv %v", argv)
		}
	})
}

func testCommandBindingCwdProof(t *testing.T) {
	t.Run("cwd proof", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		repoRoot := filepath.Dir(filepath.Dir(cwd)) // up from internal/runner

		tempOtherDir := t.TempDir()
		err = os.Chdir(tempOtherDir)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer os.Chdir(cwd) //nolint:errcheck // test cleanup

		binding := &hcr.ExecutableBinding{
			RecordPath:   "test.md",
			BindingIndex: 0,
			Kind:         "repo",
			Severity:     "HIGH",
			Run:          "python3 .hcr/checks/check-record-ids.py",
			TimeoutSecs:  2,
		}
		target := TargetProvenance{Name: "repo", Path: ""}

		verdict := EvaluateCommandBinding(true, target, binding, repoRoot)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completion %v, got %v", CompletionCompleted, verdict.Completion())
		}
		proj, _ := verdict.Projection()
		if proj != ProjectionPass {
			t.Fatalf("expected projection %v, got %v", ProjectionPass, proj)
		}
	})
}
