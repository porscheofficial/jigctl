package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestReportGolden(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "A", BindingIndex: 0, RecordID: "HCR-0001", Title: "Title A", Summary: "Summary A", Kind: "command"},
					{RecordPath: "B", BindingIndex: 0, RecordID: "HCR-0002", Title: "Title B", Summary: "Summary B", Kind: "grep"},
					{RecordPath: "C", BindingIndex: 0, RecordID: "HCR-0003", Title: "Title C", Summary: "Summary C", Kind: "grep"},
					{RecordPath: "D", BindingIndex: 0, RecordID: "HCR-0004", Title: "Title D", Summary: "Summary D", Kind: "external"},
					{RecordPath: "E", BindingIndex: 0, RecordID: "HCR-0005", Title: "Title E", Summary: "Summary E", Kind: "command"},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "A", BindingIndex: 0},
			Kind:      "command",
			Execution: &Execution{Argv: []string{"test"}, Duration: 10 * time.Millisecond},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "B", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{{Locus: Locus{File: "b.txt", Pointer: "L10"}}},
		}),
		NewNotAttemptedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "C", BindingIndex: 0},
			Kind:     "grep",
		}, ReasonKindNotExecutable),
		NewBlockedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "D", BindingIndex: 0},
			Kind:     "external",
		}, ReasonExecutableMissing),
		NewOperationalVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "E", BindingIndex: 0},
			Kind:     "command",
		}, ReasonProcessStart),
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

	goldenFile := filepath.Join("testdata", "report.golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err = os.WriteFile(goldenFile, buf.Bytes(), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Render output does not match golden file.\nGot:\n%s\nExpected:\n%s", buf.String(), string(expected))
	}
}
