package runner

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestReportInferentialSummary(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "Inf.md", BindingIndex: 0, RecordID: "HCR-2001", Title: "Inferential check", Summary: "A human decides this one", Kind: "inferential"},
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
		Rows:              BuildRows(plan, verdicts),
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
	closingLine := lines[len(lines)-1]

	if closingLine == "OK" || closingLine == "pass" || closingLine == "success" {
		t.Errorf("a run must not close on a bare success word, got: %s", closingLine)
	}
	if !strings.Contains(out, unexecutableKinds["inferential"]) {
		t.Errorf("an expected-unchecked binding must say why it did not run, got:\n%s", out)
	}
	if !strings.Contains(out, "HCR-2001") || !strings.Contains(out, "Inferential check") {
		t.Errorf("an expected-unchecked binding must be named in the scan list, got:\n%s", out)
	}
	if strings.Contains(out, "A human decides this one") {
		t.Errorf("an expected-unchecked binding must not be repeated in the detail block, got:\n%s", out)
	}
}

func TestReportLocatorAndSummary(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "tree")
	multi := filepath.Join(root, ".hcr", "MULTI.md")
	single := filepath.Join(root, ".hcr", "SINGLE.md")

	plan := &hcr.Plan{
		Root: root,
		Targets: []hcr.Target{{
			Kind: "repo",
			Bindings: []hcr.ExecutableBinding{
				{RecordPath: multi, BindingIndex: 0, RecordID: "HCR-2001", Summary: "Multi zero summary", Kind: "grep"},
				{RecordPath: multi, BindingIndex: 1, RecordID: "HCR-2001", Summary: "Multi one summary", Kind: "grep"},
				{RecordPath: single, BindingIndex: 0, RecordID: "HCR-2002", Summary: "Single summary", Kind: "grep"},
			},
		}},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: multi, BindingIndex: 1},
			Kind:     "grep",
			Findings: []Finding{{Locus: Locus{File: "x.go"}}},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: single, BindingIndex: 0},
			Kind:     "grep",
		}),
	}

	var buf bytes.Buffer
	if err := Render(RenderOptions{Out: &buf, Rows: BuildRows(plan, verdicts)}); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, root) {
		t.Errorf("tree root must not prefix any line:\n%s", out)
	}

	if !strings.Contains(out, "Multi one summary") {
		t.Errorf("a violated binding states its summary:\n%s", out)
	}
	if strings.Contains(out, "Single summary") {
		t.Errorf("a passing binding states no summary:\n%s", out)
	}
}
