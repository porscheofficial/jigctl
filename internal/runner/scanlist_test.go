package runner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestBlockedExternalReason(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{
						RecordPath:   "D",
						BindingIndex: 0,
						RecordID:     "HCR-0004",
						Title:        "Title D",
						Summary:      "Summary D",
						Kind:         "external",
					},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewBlockedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "D", BindingIndex: 0},
			Kind:     "external",
		}, ReasonExecutableMissing),
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
	if strings.Contains(out, "tool: , docs:") {
		t.Errorf("Expected output to NOT contain 'tool: , docs:', but it did.\nGot:\n%s", out)
	}

	expectedReason := reasonPhrase(Reason(ReasonExecutableMissing))
	if !strings.Contains(out, expectedReason) {
		t.Errorf("Expected output to contain reason '%s', but it did not.\nGot:\n%s", expectedReason, out)
	}
}

// TestPlannedEvidenceMatchesSettled pins the invariant the live view depends
// on: what a line says a binding is about to do must be the bytes it will say
// the binding did. Argv is taken from parseCommandBinding rather than
// reconstructed, so a change to how a command line is split fails here instead
// of showing up as a line that rewrites itself mid-run.
func TestPlannedEvidenceMatchesSettled(t *testing.T) {
	binding := hcr.ExecutableBinding{
		RecordPath: "A", RecordID: "HCR-0001", Title: "Title A",
		Kind: "command", State: "enforced", Cadence: []string{"ci"},
		Run: "  mise   run    lint  ",
	}

	argv, blocked := parseCommandBinding(&binding, &VerdictReport{})
	if blocked != nil {
		t.Fatalf("parseCommandBinding refused a well-formed binding: %v", blocked.Reason())
	}

	settled := Row{
		Kind:       "command",
		Projection: ProjectionPass,
		Execution:  &Execution{Argv: argv},
	}

	planned := plannedEvidence(&binding)
	if want := bindingEvidence(&settled); planned != want {
		t.Errorf("planned evidence %q does not match settled evidence %q", planned, want)
	}
	if planned != "mise run lint" {
		t.Errorf("expected normalised command line, got %q", planned)
	}
}

func TestPlannedEvidenceSilentWhenNothingWillRun(t *testing.T) {
	base := hcr.ExecutableBinding{
		RecordPath: "A", Kind: "command", State: "enforced",
		Cadence: []string{"ci"}, Run: "mise run lint",
	}

	for name, mutate := range map[string]func(*hcr.ExecutableBinding){
		"draft":            func(b *hcr.ExecutableBinding) { b.State = "draft" },
		"deprecated":       func(b *hcr.ExecutableBinding) { b.State = "deprecated" },
		"cadence excluded": func(b *hcr.ExecutableBinding) { b.Cadence = []string{"scheduled"} },
		"inferential":      func(b *hcr.ExecutableBinding) { b.Kind = "inferential" },
	} {
		t.Run(name, func(t *testing.T) {
			binding := base
			mutate(&binding)
			if got := plannedEvidence(&binding); got != "" {
				t.Errorf("expected no planned evidence for a binding that will not run, got %q", got)
			}
		})
	}
}
