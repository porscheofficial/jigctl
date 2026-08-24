package runner

import (
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestBuildRowsAppliesExceptions(t *testing.T) {
	plan := &hcr.Plan{
		Root: "/base",
		Targets: []hcr.Target{
			{
				Bindings: []hcr.ExecutableBinding{
					{
						RecordPath:   "/base/match.md",
						BindingIndex: 0,
						Exceptions: []hcr.Exception{
							{Scope: "*.go"}, // valid path-shape scope, matches finding
						},
					},
					{
						RecordPath:   "/base/nomatch.md",
						BindingIndex: 0,
						Exceptions: []hcr.Exception{
							{Scope: "*.ts"}, // valid path-shape scope, does NOT match
						},
					},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/match.md", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{
				{Locus: Locus{File: "main.go"}},
			},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/nomatch.md", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{
				{Locus: Locus{File: "main.go"}},
			},
		}),
	}

	rows := BuildRows(plan, verdicts)

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// 1. match.md: finding matches exception, so WaivedBy should be populated, and Projection should not be a violation (it's Pass).
	if len(rows[0].Findings) != 1 {
		t.Fatalf("expected 1 finding for match.md")
	}
	if len(rows[0].Findings[0].WaivedBy) == 0 {
		t.Errorf("expected finding to be waived, but WaivedBy is empty")
	}
	if rows[0].Projection == ProjectionViolation {
		t.Errorf("expected projection to NOT be ProjectionViolation when waived")
	}

	// 2. nomatch.md: finding does NOT match exception, so WaivedBy should be empty, and Projection should be a violation.
	if len(rows[1].Findings) != 1 {
		t.Fatalf("expected 1 finding for nomatch.md")
	}
	if len(rows[1].Findings[0].WaivedBy) != 0 {
		t.Errorf("expected finding to NOT be waived, but WaivedBy is %v", rows[1].Findings[0].WaivedBy)
	}
	if rows[1].Projection != ProjectionViolation {
		t.Errorf("expected projection to be ProjectionViolation when not waived, got %v", rows[1].Projection)
	}
}

func TestBuildRowsSwallowsScopeInvalidError(t *testing.T) {
	plan := &hcr.Plan{
		Root: "/base",
		Targets: []hcr.Target{
			{
				Bindings: []hcr.ExecutableBinding{
					{
						RecordPath:   "/base/invalid.md",
						BindingIndex: 0,
						Exceptions: []hcr.Exception{
							{Scope: "invalid_scope_no_wildcard"}, // invalid scope (not in service paths, no glob chars)
						},
					},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/invalid.md", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{
				{Locus: Locus{File: "main.go"}},
			},
		}),
	}

	// Should not panic or fail, just silently leave finding unwaived
	rows := BuildRows(plan, verdicts)

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0].Findings) != 1 {
		t.Fatalf("expected 1 finding")
	}
	if len(rows[0].Findings[0].WaivedBy) != 0 {
		t.Errorf("expected finding to NOT be waived due to invalid scope")
	}
	if rows[0].Projection != ProjectionViolation {
		t.Errorf("expected projection to be ProjectionViolation, got %v", rows[0].Projection)
	}
}
