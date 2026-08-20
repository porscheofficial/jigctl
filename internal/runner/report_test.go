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
			Rows:              BuildRows(plan, verdicts),
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
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: false,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(bufUnnorm.String(), "42ms") {
		t.Errorf("expected un-normalized render to contain real duration substring, got:\n%s", bufUnnorm.String())
	}
}
