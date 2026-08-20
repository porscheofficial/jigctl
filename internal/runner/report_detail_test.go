package runner

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func TestDetailBlockWidth(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{
						RecordPath:   "A.md",
						BindingIndex: 0,
						RecordID:     "HCR-0407",
						Title:        "Fixture expectations are never edited to make jigctl pass - this title is intentionally long to ensure it wraps correctly when rendered in the detail block.",
						Summary:      "corpus/ is normative. When jigctl disagrees with a fixture's declared valid, at or covers, fix jigctl — never the expectation. A fixture diff shows that an expectation changed, never whether the change was justified.",
						Kind:         "command",
					},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "A.md", BindingIndex: 0},
			Kind:     "command",
			Execution: &Execution{
				Argv:     []string{"false"},
				ExitCode: 1,
			},
			Findings: []Finding{{Locus: Locus{File: "some-file.go"}}},
		}),
	}

	var buf bytes.Buffer
	err := Render(RenderOptions{
		Out:               &buf,
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: true,
		Width:             100,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "HCR-0407") || !strings.Contains(out, "expectation.") {
		t.Errorf("detail block missing from output (needs a finding to trigger detail block):\n%s", out)
	}

	lines := strings.Split(out, "\n")

	for i, line := range lines {
		// Strip newline
		line = strings.TrimRight(line, "\n")
		count := utf8.RuneCountInString(line)
		if count > 100 {
			t.Errorf("line %d exceeds 100 runes (got %d): %s", i, count, line)
		}

		// Also check hanging indent for summary lines (they start with exactly 7 spaces)
		if strings.Contains(line, "corpus/") || strings.Contains(line, "expectation.") {
			if !strings.HasPrefix(line, "       ") {
				t.Errorf("summary line missing 7-space hanging indent: %q", line)
			}
			if strings.HasPrefix(line, "        ") {
				t.Errorf("summary line has too many spaces for hanging indent: %q", line)
			}
		}
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
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "tool: test-tool") || !strings.Contains(out, "docs: https://docs.example.com") {
		t.Errorf("missing tool/docs in external binding output:\n%s", out)
	}
	if !strings.Contains(out, "cat missing.txt (exit: 1)") {
		t.Errorf("missing argv/exit in command binding output:\n%s", out)
	}
}
