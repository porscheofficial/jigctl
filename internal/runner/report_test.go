package runner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func TestReportDeterminism(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "A.md", BindingIndex: 1, RecordID: "HCR-1001", Summary: "Check A1", Kind: "command"},
					{RecordPath: "B.md", BindingIndex: 0, RecordID: "HCR-1002", Summary: "Check B0", Kind: "grep"},
					{RecordPath: "A.md", BindingIndex: 0, RecordID: "HCR-1003", Summary: "Check A0", Kind: "external", Tool: "mytool", Docs: "http://docs"},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "A.md", BindingIndex: 1},
			Kind:      "command",
			Execution: &Execution{Argv: []string{"ls", "-la"}, ExitCode: 0, Duration: 42 * time.Millisecond},
		}),
		NewBlockedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "B.md", BindingIndex: 0},
			Kind:     "grep",
		}, ReasonGlobNoMatches),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "A.md", BindingIndex: 0},
			Kind:     "external",
		}),
	}

	hashes := make(map[string]struct{})
	for i := 0; i < 5; i++ {
		var buf bytes.Buffer
		err := Render(RenderOptions{
			Out:               &buf,
			Plan:              plan,
			Verdicts:          verdicts,
			NormalizeDuration: true,
		})
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		hash := sha256.Sum256(buf.Bytes())
		hexHash := hex.EncodeToString(hash[:])
		hashes[hexHash] = struct{}{}
	}

	if len(hashes) != 1 {
		t.Errorf("expected exactly 1 distinct hash across 5 renders, got %d", len(hashes))
	}

	// Now check un-normalized
	var bufUnnorm bytes.Buffer
	err := Render(RenderOptions{
		Out:               &bufUnnorm,
		Plan:              plan,
		Verdicts:          verdicts,
		NormalizeDuration: false,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(bufUnnorm.String(), "42ms") {
		t.Errorf("expected un-normalized render to contain real duration substring, got:\n%s", bufUnnorm.String())
	}
}

func TestReportInferentialSummary(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "Inf.md", BindingIndex: 0, RecordID: "HCR-2001", Summary: "Inferential check", Kind: "inferential"},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewNotAttemptedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "Inf.md", BindingIndex: 0},
			Kind:     "inferential",
		}, ReasonKindNotExecutable),
	}

	var buf bytes.Buffer
	err := Render(RenderOptions{
		Out:               &buf,
		Plan:              plan,
		Verdicts:          verdicts,
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines of output, got %d", len(lines))
	}
	summaryLine := lines[len(lines)-1]

	if summaryLine == "OK" || summaryLine == "pass" || summaryLine == "success" {
		t.Errorf("summary line should not be a bare success word, got: %s", summaryLine)
	}
	if !strings.Contains(summaryLine, "1 expected-unchecked") {
		t.Errorf("summary line should mention unchecked count, got: %s", summaryLine)
	}
}

func TestReportExternalAndCommand(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "Ext.md", BindingIndex: 0, RecordID: "HCR-3001", Summary: "External check", Kind: "external", Tool: "test-tool", Docs: "https://docs.example.com"},
					{RecordPath: "Cmd.md", BindingIndex: 0, RecordID: "HCR-3002", Summary: "Command check", Kind: "command"},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "Ext.md", BindingIndex: 0},
			Kind:     "external",
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "Cmd.md", BindingIndex: 0},
			Kind:      "command",
			Execution: &Execution{Argv: []string{"cat", "missing.txt"}, ExitCode: 1, Duration: 10 * time.Millisecond},
			Findings:  []Finding{{Locus: Locus{File: "missing.txt"}}},
		}),
	}

	var buf bytes.Buffer
	err := Render(RenderOptions{
		Out:               &buf,
		Plan:              plan,
		Verdicts:          verdicts,
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "tool: test-tool") || !strings.Contains(out, "docs: https://docs.example.com") {
		t.Errorf("missing tool/docs in external binding output:\n%s", out)
	}
	if !strings.Contains(out, "argv: cat missing.txt") || !strings.Contains(out, "exit: 1") {
		t.Errorf("missing argv/exit in command binding output:\n%s", out)
	}
}
