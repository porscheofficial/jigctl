package runner

import (
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		verdicts []ExitSummary
		strict   bool
		expected int
	}{
		{
			name:     "no verdicts",
			verdicts: nil,
			expected: 77,
		},
		{
			name: "only-inferential (all expected-unchecked) → 77",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked},
				{Projection: ProjectionExpectedUnchecked},
			},
			expected: 77,
		},
		{
			name: "this repo's real shape (8 passes + HCR-0407 expected-unchecked) → 0",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionPass},
				{Projection: ProjectionExpectedUnchecked},
			},
			expected: 0,
		},
		{
			name: "same set with one switched to blocked-unchecked → 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionExpectedUnchecked},
				{Projection: ProjectionBlockedUnchecked},
			},
			expected: 1,
		},
		{
			name: "some-expected-unchecked-no-violations → 0",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked},
				{Projection: ProjectionPass},
			},
			expected: 0,
		},
		{
			name: "mix-of-blocked-and-expected-unchecked → 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked},
				{Projection: ProjectionBlockedUnchecked},
			},
			expected: 1,
		},
		{
			name: "strict-flag + some-unchecked (that would otherwise be 0) → non-zero",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionExpectedUnchecked},
			},
			strict:   true,
			expected: 1,
		},
		{
			name: "one blocking violation → 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionViolation, IsBlocking: true},
			},
			expected: 1,
		},
		{
			name: "one advisory-only violation → 0",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionViolation, IsBlocking: false},
			},
			expected: 0,
		},
		{
			name: "one operational verdict → 2",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionOperational},
			},
			expected: 2,
		},
		{
			name: "aggregate scenario: pass + advisory-violation + expected-unchecked + blocked-unchecked in one run → exit 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionViolation, IsBlocking: false},
				{Projection: ProjectionExpectedUnchecked},
				{Projection: ProjectionBlockedUnchecked},
			},
			expected: 1,
		},
		{
			name: "operational overrides blocked-unchecked",
			verdicts: []ExitSummary{
				{Projection: ProjectionBlockedUnchecked},
				{Projection: ProjectionOperational},
			},
			expected: 2,
		},
		{
			name: "operational overrides blocking violation",
			verdicts: []ExitSummary{
				{Projection: ProjectionViolation, IsBlocking: true},
				{Projection: ProjectionOperational},
			},
			expected: 2,
		},
		{
			name: "one pass + one cadence-deselected under strict=true → exit code 0",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionExpectedUnchecked, Deselected: true},
			},
			strict:   true,
			expected: 0,
		},
		{
			name: "one pass + one cadence-excluded under strict=true → exit code 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionPass},
				{Projection: ProjectionExpectedUnchecked, Deselected: false},
			},
			strict:   true,
			expected: 1,
		},
		{
			name: "all-deselected → exit code 77",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked, Deselected: true},
				{Projection: ProjectionExpectedUnchecked, Deselected: true},
			},
			strict:   true,
			expected: 77,
		},
		{
			name: "deselected + one blocked → exit code 1",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked, Deselected: true},
				{Projection: ProjectionBlockedUnchecked},
			},
			expected: 1,
		},
		{
			name: "deselected + one advisory violation → exit code 0",
			verdicts: []ExitSummary{
				{Projection: ProjectionExpectedUnchecked, Deselected: true},
				{Projection: ProjectionViolation, IsBlocking: false},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateExitCode(tt.verdicts, tt.strict)
			if got != tt.expected {
				t.Errorf("AggregateExitCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
