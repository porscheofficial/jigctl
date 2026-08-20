package runner

import (
	"reflect"
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

// oldComputeProjection replicates internal/runner/report.go:172-183.
func oldComputeProjection(v *Verdict, mutated []Finding) Projection {
	if v.Completion() == CompletionCompleted {
		for _, f := range mutated {
			if len(f.WaivedBy) == 0 {
				return ProjectionViolation
			}
		}
		return ProjectionPass
	}
	proj, _ := v.Projection()
	return proj
}

// oldCmdRunProjection replicates cmd/jigctl/run.go:127-138.
func oldCmdRunProjection(v *Verdict, mutated []Finding) Projection {
	var proj Projection
	if v.Completion() == CompletionCompleted {
		proj = ProjectionPass
		for _, f := range mutated {
			if len(f.WaivedBy) == 0 {
				proj = ProjectionViolation
				break
			}
		}
	} else {
		proj, _ = v.Projection()
	}
	return proj
}

func TestBaselineProjectionsMatch(t *testing.T) {
	// A representative verdict set covering all five projections plus waived and unwaived findings.
	tcs := []struct {
		name    string
		verdict *Verdict
		mutated []Finding
	}{
		{
			name: "completed pass",
			verdict: NewCompletedVerdict(&VerdictReport{
				Severity: "blocking",
				Findings: []Finding{},
			}),
			mutated: []Finding{},
		},
		{
			name: "completed violation",
			verdict: NewCompletedVerdict(&VerdictReport{
				Severity: "blocking",
				Findings: []Finding{{Locus: Locus{File: "a"}}},
			}),
			mutated: []Finding{{Locus: Locus{File: "a"}}},
		},
		{
			name: "completed waived pass",
			verdict: NewCompletedVerdict(&VerdictReport{
				Severity: "blocking",
				Findings: []Finding{{Locus: Locus{File: "a"}, WaivedBy: []ExceptionIdentity{{}}}},
			}),
			mutated: []Finding{{Locus: Locus{File: "a"}, WaivedBy: []ExceptionIdentity{{}}}},
		},
		{
			name: "completed partially waived violation",
			verdict: NewCompletedVerdict(&VerdictReport{
				Severity: "blocking",
				Findings: []Finding{
					{Locus: Locus{File: "a"}, WaivedBy: []ExceptionIdentity{{}}},
					{Locus: Locus{File: "b"}},
				},
			}),
			mutated: []Finding{
				{Locus: Locus{File: "a"}, WaivedBy: []ExceptionIdentity{{}}},
				{Locus: Locus{File: "b"}},
			},
		},
		{
			name: "blocked unchecked",
			verdict: NewBlockedVerdict(&VerdictReport{
				Severity: "blocking",
			}, ReasonExecutableMissing),
			mutated: nil,
		},
		{
			name: "expected unchecked",
			verdict: NewNotAttemptedVerdict(&VerdictReport{
				Severity: "blocking",
			}, ReasonKindNotExecutable),
			mutated: nil,
		},
		{
			name: "operational",
			verdict: NewOperationalVerdict(&VerdictReport{
				Severity: "blocking",
			}, ReasonProcessStart),
			mutated: nil,
		},
	}

	exitSummariesReport := make([]ExitSummary, 0, len(tcs))
	exitSummariesCmd := make([]ExitSummary, 0, len(tcs))

	for _, tc := range tcs {
		p1 := oldComputeProjection(tc.verdict, tc.mutated)
		p2 := oldCmdRunProjection(tc.verdict, tc.mutated)
		if p1 != p2 {
			t.Errorf("%s: projection mismatch: report.go=%v vs cmd/run.go=%v", tc.name, p1, p2)
		}

		isBlocking := tc.verdict.Report().Severity != "advisory"
		exitSummariesReport = append(exitSummariesReport, ExitSummary{Projection: p1, IsBlocking: isBlocking})
		exitSummariesCmd = append(exitSummariesCmd, ExitSummary{Projection: p2, IsBlocking: isBlocking})
	}

	if !reflect.DeepEqual(exitSummariesReport, exitSummariesCmd) {
		t.Errorf("ExitSummaries differ")
	}

	exitCodeReportFalse := AggregateExitCode(exitSummariesReport, false)
	exitCodeCmdFalse := AggregateExitCode(exitSummariesCmd, false)
	if exitCodeReportFalse != exitCodeCmdFalse {
		t.Errorf("ExitCode strict=false mismatch: %d vs %d", exitCodeReportFalse, exitCodeCmdFalse)
	}

	exitCodeReportTrue := AggregateExitCode(exitSummariesReport, true)
	exitCodeCmdTrue := AggregateExitCode(exitSummariesCmd, true)
	if exitCodeReportTrue != exitCodeCmdTrue {
		t.Errorf("ExitCode strict=true mismatch: %d vs %d", exitCodeReportTrue, exitCodeCmdTrue)
	}

	rows := BuildRows(nil, tcsVerdicts(tcs))
	exitSummariesRow := ExitSummaries(rows)

	if !reflect.DeepEqual(exitSummariesReport, exitSummariesRow) {
		t.Errorf("BuildRows ExitSummaries differ from baseline:\nReport: %v\nRow: %v", exitSummariesReport, exitSummariesRow)
	}
}

func tcsVerdicts(tcs []struct {
	name    string
	verdict *Verdict
	mutated []Finding
},
) []*Verdict {
	v := make([]*Verdict, 0, len(tcs))
	for _, tc := range tcs {
		v = append(v, tc.verdict)
	}
	return v
}

func TestTitle(t *testing.T) {
	plan := &hcr.Plan{
		Root: "/base",
		Targets: []hcr.Target{
			{
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "/base/normal.md", BindingIndex: 0, Title: "normal title"},
					{RecordPath: "/base/colon.md", BindingIndex: 0, Title: "title: with a colon"},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/normal.md", BindingIndex: 0},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/colon.md", BindingIndex: 0},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/missing.md", BindingIndex: 0},
		}),
	}

	rows := BuildRows(plan, verdicts)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0].Title != "normal title" {
		t.Errorf("expected normal title, got %q", rows[0].Title)
	}

	if rows[1].Title != "title: with a colon" {
		t.Errorf("expected title: with a colon, got %q", rows[1].Title)
	}

	if rows[2].Title != "" {
		t.Errorf("expected empty title for not-in-lookup, got %q", rows[2].Title)
	}
}

func TestUnwaivedFileCount(t *testing.T) {
	rows := []Row{
		{
			Findings: []Finding{
				{Locus: Locus{File: "a"}},                                    // 1
				{Locus: Locus{File: "a"}},                                    // duplicate file, count 1
				{Locus: Locus{File: "b"}},                                    // 2
				{Locus: Locus{File: "c"}, WaivedBy: []ExceptionIdentity{{}}}, // waived, count 2
				{Locus: Locus{File: ""}},                                     // empty file, count 2
			},
		},
	}

	count := UnwaivedFileCount(rows)
	if count != 2 {
		t.Errorf("expected 2 files, got %d", count)
	}
}

func TestFindingShapes(t *testing.T) {
	// A grep forbid-match has Locus{File: match, Pointer: "L<n>"}.
	// A grep missing-require has Locus{File: binding.File} where File is the GLOB PATTERN, and no pointer.
	// A config-assert finding has Locus{File: binding.File, Pointer: <RFC 6901 pointer>}.
	// A command finding has Finding{Severity: ...} with a completely EMPTY Locus.

	// The Row should just preserve these as they are, without modifying them.
	// We'll just verify the fields pass through properly.
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Kind: "grep",
			Findings: []Finding{
				{Locus: Locus{File: "main.go", Pointer: "L42"}},
				{Locus: Locus{File: "*.go"}},
			},
		}),
		NewCompletedVerdict(&VerdictReport{
			Kind: "config-assert",
			Findings: []Finding{
				{Locus: Locus{File: "jig.toml", Pointer: "/rationale/ADR"}},
			},
		}),
		NewCompletedVerdict(&VerdictReport{
			Kind: "command",
			Findings: []Finding{
				{Severity: "blocking"}, // empty locus
			},
		}),
	}

	rows := BuildRows(nil, verdicts)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows")
	}

	if rows[0].Findings[0].Locus.Pointer != "L42" {
		t.Errorf("expected L42")
	}
	if rows[0].Findings[1].Locus.File != "*.go" || rows[0].Findings[1].Locus.Pointer != "" {
		t.Errorf("expected *.go and empty pointer")
	}
	if rows[1].Findings[0].Locus.Pointer != "/rationale/ADR" {
		t.Errorf("expected /rationale/ADR")
	}
	if rows[2].Findings[0].Locus.File != "" || rows[2].Findings[0].Locus.Pointer != "" {
		t.Errorf("expected empty locus for command finding")
	}
}
