package runner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
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

	expectedReason := reasonCode(Reason(ReasonExecutableMissing))
	if !strings.Contains(out, expectedReason) {
		t.Errorf("Expected output to contain reason '%s', but it did not.\nGot:\n%s", expectedReason, out)
	}
}
