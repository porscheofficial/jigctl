package runner

import (
	"reflect"
	"testing"
	"time"
)

func TestVerdictProjection_is_invalid_for_zero_value(t *testing.T) {
	// Given
	var verdict Verdict

	// When
	projection, valid := verdict.Projection()

	// Then
	if valid || projection == ProjectionPass {
		t.Fatalf("zero Verdict projection = (%q, %t), want non-pass invalid projection", projection, valid)
	}
}

func TestVerdictProjection_is_invalid_for_nil_receiver(t *testing.T) {
	// Given
	var verdict *Verdict

	// When
	projection, valid := verdict.Projection()

	// Then
	if valid || projection == ProjectionPass {
		t.Fatalf("nil *Verdict projection = (%q, %t), want non-pass invalid projection", projection, valid)
	}
}

func TestVerdictProjection_derives_completed_result_from_unwaived_findings(t *testing.T) {
	tests := []struct {
		name     string
		findings []Finding
		want     Projection
	}{
		{name: "no findings passes", want: ProjectionPass},
		{
			name: "only waived findings pass",
			findings: []Finding{{
				Locus:    Locus{File: "main.go"},
				Severity: "blocking",
				WaivedBy: []ExceptionIdentity{{RecordPath: ".hcr/HCR-0406.md", ExceptionIndex: 0}},
			}},
			want: ProjectionPass,
		},
		{
			name:     "an unwaived finding violates",
			findings: []Finding{{Locus: Locus{File: "main.go"}, Severity: "blocking"}},
			want:     ProjectionViolation,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			verdict := NewCompletedVerdict(&VerdictReport{Findings: test.findings})

			// When
			projection, valid := verdict.Projection()

			// Then
			if !valid || projection != test.want {
				t.Fatalf("Projection() = (%q, %t), want (%q, true)", projection, valid, test.want)
			}
		})
	}
}

func TestVerdictProjection_noncompleted_state_takes_precedence_over_findings(t *testing.T) {
	// Given
	report := VerdictReport{Findings: []Finding{{Locus: Locus{File: "partial.go"}, Severity: "blocking", Partial: true}}}
	tests := []struct {
		name       string
		verdict    *Verdict
		projection Projection
	}{
		{"blocked", NewBlockedVerdict(&report, ReasonTimeout), ProjectionBlockedUnchecked},
		{"not attempted", NewNotAttemptedVerdict(&report, ReasonCadenceExcluded), ProjectionExpectedUnchecked},
		{"operational", NewOperationalVerdict(&report, ReasonInvocationCancelled), ProjectionOperational},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			projection, valid := test.verdict.Projection()

			// Then
			if !valid || projection != test.projection {
				t.Fatalf("Projection() = (%q, %t), want (%q, true)", projection, valid, test.projection)
			}
		})
	}
}

func TestVerdictReport_carries_actionable_execution_data(t *testing.T) {
	// Given
	declared := 30 * time.Second
	effective := 120 * time.Second
	report := VerdictReport{
		Identity:  BindingIdentity{RecordPath: ".hcr/HCR-0401.md", BindingIndex: 2},
		Target:    TargetProvenance{Name: "api", Path: "services/api"},
		Kind:      "command",
		Severity:  "blocking",
		Timeouts:  TimeoutRecord{Declared: &declared, Resolved: declared, Effective: &effective},
		Execution: &Execution{Argv: []string{"mise", "run", "lint"}, ExitCode: 1, Duration: time.Second},
	}

	// When
	verdict := NewCompletedVerdict(&report)

	// Then
	if !reflect.DeepEqual(verdict.Report(), report) {
		t.Fatalf("Report() = %#v, want original report", verdict.Report())
	}
}

func TestRunnerTaxonomyIsExhaustive(t *testing.T) {
	if projectionCount != 6 {
		t.Fatalf("projectionCount = %d, want exactly 6", projectionCount)
	}

	for i := 0; i < reasonCount; i++ {
		r := Reason(i)

		if _, ok := reasonData[r]; !ok {
			t.Errorf("reason %d missing from reasonData", i)
		}
		if _, ok := reasonPhrases[r]; !ok {
			t.Errorf("reason %d missing from reasonPhrases", i)
		}

		var expectedCompletion Completion
		var expectedProjection Projection

		report := &VerdictReport{}
		var verdict *Verdict

		switch r {
		case ReasonNone:
			expectedCompletion = CompletionCompleted
			expectedProjection = ProjectionPass
			verdict = NewCompletedVerdict(report)
		case Reason(ReasonExecutableMissing),
			Reason(ReasonExecutableDenied),
			Reason(ReasonTimeout),
			Reason(ReasonAuthorizationDenied),
			Reason(ReasonGlobNoMatches),
			Reason(ReasonInputMissing),
			Reason(ReasonInputUnreadable),
			Reason(ReasonInputMalformed),
			Reason(ReasonPointerMalformed),
			Reason(ReasonPatternInvalid),
			Reason(ReasonScopeInvalid),
			Reason(ReasonGlobInvalid),
			Reason(ReasonArgvInvalid),
			Reason(ReasonFormatUnsupported),
			Reason(ReasonModifierUnimplemented):
			expectedCompletion = CompletionBlocked
			expectedProjection = ProjectionBlockedUnchecked
			verdict = NewBlockedVerdict(report, BlockedReason(r))
		case Reason(ReasonKindNotExecutable),
			Reason(ReasonCadenceExcluded),
			Reason(ReasonCadenceDeselected),
			Reason(ReasonRecordDraft),
			Reason(ReasonRecordDeprecated):
			expectedCompletion = CompletionNotAttempted
			expectedProjection = ProjectionExpectedUnchecked
			verdict = NewNotAttemptedVerdict(report, ExpectedReason(r))
		case Reason(ReasonProcessStart),
			Reason(ReasonPathEscapesRoot),
			Reason(ReasonLimitExceeded),
			Reason(ReasonInvocationCancelled):
			expectedCompletion = CompletionOperational
			expectedProjection = ProjectionOperational
			verdict = NewOperationalVerdict(report, OperationalReason(r))
		default:
			t.Fatalf("reason %d not mapped in taxonomy test", i)
		}

		if verdict.Completion() != expectedCompletion {
			t.Errorf("reason %d completion = %v, want %v", i, verdict.Completion(), expectedCompletion)
		}
		proj, valid := verdict.Projection()
		if !valid {
			t.Errorf("reason %d projection invalid", i)
		}
		if proj != expectedProjection {
			t.Errorf("reason %d projection = %v, want %v", i, proj, expectedProjection)
		}
	}
}
