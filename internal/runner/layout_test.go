package runner

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func widthPlan() (*hcr.Plan, []*Verdict) {
	long := strings.Repeat("very long title that no terminal can hold ", 4)
	plan := &hcr.Plan{Targets: []hcr.Target{{
		Kind: "repo",
		Bindings: []hcr.ExecutableBinding{
			{RecordPath: "A.md", RecordID: "HCR-0001", Title: long, Summary: long, Kind: "grep"},
			{RecordPath: "B.md", RecordID: "HCR-0002", Title: "short", Kind: "command"},
		},
	}}}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "A.md"},
			Kind:     "grep",
			Findings: []Finding{{Locus: Locus{File: "some/deeply/nested/path/to/a/file.go", Pointer: "L42"}}},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "B.md"},
			Kind:      "command",
			Execution: &Execution{Argv: []string{"echo"}},
		}),
	}
	return plan, verdicts
}

func TestScanLineNeverExceedsWidth(t *testing.T) {
	plan, verdicts := widthPlan()
	rows := BuildRows(plan, verdicts)

	for width := scanFixedCells + minTitleCells + 1; width <= 200; width++ {
		var buf bytes.Buffer
		if err := Render(RenderOptions{Out: &buf, Rows: rows, Width: width}); err != nil {
			t.Fatalf("Render at width %d failed: %v", width, err)
		}
		for _, line := range strings.Split(buf.String(), "\n") {
			if n := utf8.RuneCountInString(line); n > width && breakable(line) {
				t.Fatalf("width %d produced a %d-rune line: %q", width, n, line)
			}
		}
	}
}

// breakable reports whether a line could have been made to fit. A line
// carrying a single unbreakable token — a file path an editor must be able to
// open, a word with no spaces in it — is left whole and allowed to wrap,
// because truncating it would cost the reader the thing the line is for.
func breakable(line string) bool {
	return len(strings.Fields(strings.TrimSpace(line))) > 1
}

func TestUnmeasurableWidthTruncatesNothing(t *testing.T) {
	plan, verdicts := widthPlan()
	rows := BuildRows(plan, verdicts)

	var buf bytes.Buffer
	if err := Render(RenderOptions{Out: &buf, Rows: rows}); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, "…") {
		t.Errorf("a destination with no measurable width must truncate nothing:\n%s", out)
	}
	if !strings.Contains(out, rows[0].Title) {
		t.Errorf("full title missing from output:\n%s", out)
	}
}
