package runner

import (
	"bytes"
	"testing"
	"time"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

var expectedRenderedOutput = `  ✓  passed
  ✗  found a violation
  ○  did not run by design
  ▲  could not run and blocks the result
  ◆  jigctl itself failed

  ✓    HCR-0001  Pass                   <dur>  cmd
  ✗    HCR-0002  Violation                    1 finding
  ○    HCR-0003  Expected Unchecked           nothing here for jigctl to run
  ▲    HCR-0004  Blocked Unchecked            executable is not on PATH
  ◆    HCR-0005  Operational                  process could not be started
  ✗    HCR-0006  Waived                       1 finding

  ✗    HCR-0002  Violation
       [pattern] b.txt

  ▲    HCR-0004  Blocked Unchecked
       executable is not on PATH

  ◆    HCR-0005  Operational
       process could not be started

  ✗    HCR-0006  Waived
       [pattern] foo.txt
`

func TestRefactorCharacterization(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					// 1. Pass
					{RecordPath: "A", BindingIndex: 0, RecordID: "HCR-0001", Title: "Pass", Kind: "command"},
					// 2. Violation
					{RecordPath: "B", BindingIndex: 0, RecordID: "HCR-0002", Title: "Violation", Kind: "grep"},
					// 3. ExpectedUnchecked
					{RecordPath: "C", BindingIndex: 0, RecordID: "HCR-0003", Title: "Expected Unchecked", Kind: "grep"},
					// 4. BlockedUnchecked
					{RecordPath: "D", BindingIndex: 0, RecordID: "HCR-0004", Title: "Blocked Unchecked", Kind: "external"},
					// 5. Operational
					{RecordPath: "E", BindingIndex: 0, RecordID: "HCR-0005", Title: "Operational", Kind: "command"},
					// 6. Waived Violation (Should Pass projection but have findings)
					{RecordPath: "F", BindingIndex: 0, RecordID: "HCR-0006", Title: "Waived", Kind: "grep", Exceptions: []hcr.Exception{{Scope: "foo.txt"}}},
				},
			},
		},
	}
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "A", BindingIndex: 0},
			Kind:      "command",
			Execution: &Execution{Argv: []string{"cmd"}, Duration: 10 * time.Millisecond},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "B", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{{Locus: Locus{File: "b.txt"}}},
			Severity: "error",
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
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "F", BindingIndex: 0},
			Kind:     "grep",
			Findings: []Finding{{Locus: Locus{File: "foo.txt"}}}, // WaivedBy is mutated below
			Severity: "error",
		}),
	}

	var buf bytes.Buffer
	rows := BuildRows(plan, verdicts)
	err := Render(RenderOptions{
		Out:               &buf,
		Rows:              rows,
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	summaries := ExitSummaries(rows)
	strictExit := AggregateExitCode(summaries, true)
	nonStrictExit := AggregateExitCode(summaries, false)

	if strictExit != 2 {
		t.Errorf("Expected strict exit code 2, got %d", strictExit)
	}
	if nonStrictExit != 2 {
		t.Errorf("Expected non-strict exit code 2, got %d", nonStrictExit)
	}

	renderedOut := buf.String()
	if renderedOut != expectedRenderedOutput {
		t.Errorf("Rendered output mismatch.\nExpected:\n%q\nGot:\n%q", expectedRenderedOutput, renderedOut)
	}
}
