package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

type configAssertTestCase struct {
	name    string
	binding hcr.ExecutableBinding
	check   func(t *testing.T, v *Verdict)
}

func TestConfigAssert(t *testing.T) {
	tmp := t.TempDir()

	jigToml := `
[rationale]
ADR = "docs/adr/{rest}-*.md"

[some_table]
date = 1979-05-27T07:32:00Z
number = 42
bool_val = true
`
	if err := os.WriteFile(filepath.Join(tmp, "jig.toml"), []byte(jigToml), 0o644); err != nil {
		t.Fatal(err)
	}

	testYaml := `
nested:
  key: value
array:
  - 1
  - 2
`
	if err := os.WriteFile(filepath.Join(tmp, "test.yaml"), []byte(testYaml), 0o644); err != nil {
		t.Fatal(err)
	}

	testJSON := `{"escaped/key": {"~escaped": "found"}}`
	if err := os.WriteFile(filepath.Join(tmp, "test.json"), []byte(testJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := configAssertTestCasesPart1()
	tests = append(tests, configAssertTestCasesPart2()...)

	plan := hcr.Plan{Root: tmp}
	target := hcr.Target{}

	// A relative root is a shape ExecutionPlan cannot produce, and confine
	// canonicalizes the base before making it absolute, so passing one made
	// this case report an escape under a symlinked checkout.
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	realPlan := hcr.Plan{Root: repoRoot}

	tests = append(tests, configAssertTestCase{
		name: "real jig.toml rationale ADR correct",
		binding: hcr.ExecutableBinding{
			File:  "jig.toml",
			Path:  "/rationale/ADR",
			Op:    "equals",
			Value: "docs/adr/{rest}-*.md",
		},
		check: func(t *testing.T, v *Verdict) {
			if v.Completion() != CompletionCompleted {
				t.Errorf("Expected completed, got %v", v.Completion())
			}
			if len(v.Report().Findings) != 0 {
				t.Errorf("Expected pass, got findings: %v", v.Report().Findings)
			}
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plan
			if tt.name == "real jig.toml rationale ADR correct" {
				p = realPlan
			}
			v := ConfigAssert(p, target, tt.binding)
			tt.check(t, v)
		})
	}
}

func configAssertTestCasesPart1() []configAssertTestCase {
	return []configAssertTestCase{
		{
			name: "jig.toml rationale ADR correct",
			binding: hcr.ExecutableBinding{
				File:  "jig.toml",
				Path:  "/rationale/ADR",
				Op:    "equals",
				Value: "docs/adr/{rest}-*.md",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted {
					t.Errorf("Expected completed, got %v", v.Completion())
				}
				if len(v.Report().Findings) != 0 {
					t.Errorf("Expected pass, got findings: %v", v.Report().Findings)
				}
			},
		},
		{
			name: "jig.toml time type mismatch",
			binding: hcr.ExecutableBinding{
				File:  "jig.toml",
				Path:  "/some_table/date",
				Op:    "equals",
				Value: "1979-05-27T07:32:00Z",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted {
					t.Errorf("Expected completed, got %v", v.Completion())
				}
				if len(v.Report().Findings) != 1 {
					t.Errorf("Expected finding due to time mismatch")
				}
			},
		},
		{
			name: "gte against string mismatch",
			binding: hcr.ExecutableBinding{
				File:  "jig.toml",
				Path:  "/rationale/ADR",
				Op:    "gte",
				Value: 123.0,
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted {
					t.Errorf("Expected completed, got %v", v.Completion())
				}
				if len(v.Report().Findings) != 1 {
					t.Errorf("Expected finding due to type mismatch")
				}
			},
		},
		{
			name: "absent missing file",
			binding: hcr.ExecutableBinding{
				File: "does-not-exist.json",
				Path: "/missing",
				Op:   "absent",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionBlocked || v.Reason() != Reason(ReasonInputMissing) {
					t.Errorf("Expected blocked missing file, got %v %v", v.Completion(), v.Reason())
				}
			},
		},
	}
}

func configAssertTestCasesPart2() []configAssertTestCase {
	return []configAssertTestCase{
		{
			name: "absent missing pointer",
			binding: hcr.ExecutableBinding{
				File: "test.json",
				Path: "/missing",
				Op:   "absent",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted || len(v.Report().Findings) != 0 {
					t.Errorf("Expected completed with no findings")
				}
			},
		},
		{
			name: "absent existing pointer",
			binding: hcr.ExecutableBinding{
				File: "test.yaml",
				Path: "/nested/key",
				Op:   "absent",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted || len(v.Report().Findings) == 0 {
					t.Errorf("Expected finding for existing pointer on absent")
				}
			},
		},
		{
			name: "pointer escaped json",
			binding: hcr.ExecutableBinding{
				File:  "test.json",
				Path:  "/escaped~1key/~0escaped",
				Op:    "equals",
				Value: "found",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionCompleted || len(v.Report().Findings) != 0 {
					t.Errorf("Expected pass for escaped pointer, findings %v", v.Report().Findings)
				}
			},
		},
		{
			name: "pointer malformed array index",
			binding: hcr.ExecutableBinding{
				File:  "test.yaml",
				Path:  "/array/str",
				Op:    "equals",
				Value: "1",
			},
			check: func(t *testing.T, v *Verdict) {
				if v.Completion() != CompletionBlocked || v.Reason() != Reason(ReasonPointerMalformed) {
					t.Errorf("Expected blocked ReasonPointerMalformed, got %v", v.Reason())
				}
			},
		},
	}
}
