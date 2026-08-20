package runner

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

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
	summaryLine := lines[len(lines)-1]

	if summaryLine == "OK" || summaryLine == "pass" || summaryLine == "success" {
		t.Errorf("summary line should not be a bare success word, got: %s", summaryLine)
	}
	if !strings.Contains(summaryLine, "UNCHECKED=1") {
		t.Errorf("summary line should mention unchecked count, got: %s", summaryLine)
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
