package runner

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestPlainMatchesPreChangeBinary(t *testing.T) {
	// Construct the exact rows that the pre-change binary produced in run-allowexec.txt
	rows := []Row{
		{
			Locator:    ".hcr/HCR-0401-schema-metaschema-valid.md",
			RecordID:   "HCR-0401",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
			Kind:       "command",
			Execution: &Execution{
				Argv:     []string{"mise", "run", "metaschema"},
				ExitCode: 0,
				Duration: 386055917 * time.Nanosecond,
			},
		},
		{
			Locator:    ".hcr/HCR-0402-corpus-gates-every-change.md",
			RecordID:   "HCR-0402",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
			Kind:       "command",
			Execution: &Execution{
				Argv:     []string{"mise", "run", "corpus"},
				ExitCode: 0,
				Duration: 2025092375 * time.Nanosecond,
			},
		},
		{
			Locator:    ".hcr/HCR-0403-schema-shape-matches-register.md",
			RecordID:   "HCR-0403",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
			Kind:       "command",
			Execution: &Execution{
				Argv:     []string{"tools/schema-shape.py"},
				ExitCode: 0,
				Duration: 159345667 * time.Nanosecond,
			},
		},
		{
			Locator:    ".hcr/HCR-0404-go-source-passes-gates.md",
			RecordID:   "HCR-0404",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
			Kind:       "command",
			Execution: &Execution{
				Argv:     []string{"mise", "run", "lint"},
				ExitCode: 0,
				Duration: 2618624541 * time.Nanosecond,
			},
		},
		{
			Locator:    ".hcr/HCR-0405-mise-owns-every-entrypoint.md",
			RecordID:   "HCR-0405",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0406-single-hcr-implementation.md",
			RecordID:   "HCR-0406",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0407-fixtures-never-edited-to-pass.md",
			RecordID:   "HCR-0407",
			Projection: ProjectionExpectedUnchecked,
			Reason:     Reason(ReasonKindNotExecutable),
		},
		{
			Locator:    ".hcr/HCR-0408-real-record-ids-stay-in-band.md",
			RecordID:   "HCR-0408",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
			Kind:       "command",
			Execution: &Execution{
				Argv:     []string{"tools/check-record-ids.py"},
				ExitCode: 0,
				Duration: 107615917 * time.Nanosecond,
			},
		},
		{
			Locator:    ".hcr/HCR-0409-adr-rationale-prefix-is-mapped.md",
			RecordID:   "HCR-0409",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0410-invocation-clock-is-not-swappable.md",
			RecordID:   "HCR-0410",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
	}

	expectedBytes, err := os.ReadFile("testdata/plain.golden")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	var buf bytes.Buffer
	err = RenderPlain(&buf, rows)
	if err != nil {
		t.Fatalf("RenderPlain returned error: %v", err)
	}

	expected := string(expectedBytes)
	got := buf.String()
	if expected != got {
		t.Errorf("RenderPlain output mismatch\nExpected:\n%s\n\nGot:\n%s", expected, got)
	}
}

func TestPlainMatchesPreChangeBinaryDenied(t *testing.T) {
	rows := []Row{
		{
			Locator:    ".hcr/HCR-0401-schema-metaschema-valid.md",
			RecordID:   "HCR-0401",
			Projection: ProjectionBlockedUnchecked,
			Reason:     Reason(ReasonAuthorizationDenied),
			Kind:       "command",
		},
		{
			Locator:    ".hcr/HCR-0402-corpus-gates-every-change.md",
			RecordID:   "HCR-0402",
			Projection: ProjectionBlockedUnchecked,
			Reason:     Reason(ReasonAuthorizationDenied),
			Kind:       "command",
		},
		{
			Locator:    ".hcr/HCR-0403-schema-shape-matches-register.md",
			RecordID:   "HCR-0403",
			Projection: ProjectionBlockedUnchecked,
			Reason:     Reason(ReasonAuthorizationDenied),
			Kind:       "command",
		},
		{
			Locator:    ".hcr/HCR-0404-go-source-passes-gates.md",
			RecordID:   "HCR-0404",
			Projection: ProjectionBlockedUnchecked,
			Reason:     Reason(ReasonAuthorizationDenied),
			Kind:       "command",
		},
		{
			Locator:    ".hcr/HCR-0405-mise-owns-every-entrypoint.md",
			RecordID:   "HCR-0405",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0406-single-hcr-implementation.md",
			RecordID:   "HCR-0406",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0407-fixtures-never-edited-to-pass.md",
			RecordID:   "HCR-0407",
			Projection: ProjectionExpectedUnchecked,
			Reason:     Reason(ReasonKindNotExecutable),
		},
		{
			Locator:    ".hcr/HCR-0408-real-record-ids-stay-in-band.md",
			RecordID:   "HCR-0408",
			Projection: ProjectionBlockedUnchecked,
			Reason:     Reason(ReasonAuthorizationDenied),
			Kind:       "command",
		},
		{
			Locator:    ".hcr/HCR-0409-adr-rationale-prefix-is-mapped.md",
			RecordID:   "HCR-0409",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
		{
			Locator:    ".hcr/HCR-0410-invocation-clock-is-not-swappable.md",
			RecordID:   "HCR-0410",
			Projection: ProjectionPass,
			Reason:     ReasonNone,
		},
	}

	expectedBytes, err := os.ReadFile("testdata/plain-denied.golden")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	var buf bytes.Buffer
	err = RenderPlain(&buf, rows)
	if err != nil {
		t.Fatalf("RenderPlain returned error: %v", err)
	}

	expected := string(expectedBytes)
	got := buf.String()
	if expected != got {
		t.Errorf("RenderPlain output mismatch\nExpected:\n%s\n\nGot:\n%s", expected, got)
	}
}

func normalizeDuration(s string) string {
	re := regexp.MustCompile(`duration: [a-zA-Z0-9.µ]+`)
	return re.ReplaceAllString(s, `duration: <normalized>`)
}

func TestPlainDurationOnlyDifference(t *testing.T) {
	// The original pre-change output
	originalBytes, err := os.ReadFile("/tmp/jigctl-prechange/run-allowexec.txt")
	if err != nil {
		t.Fatalf("failed to read original output: %v", err)
	}
	originalStr := string(originalBytes)

	// Our current output based on the new golden
	currentBytes, err := os.ReadFile("testdata/plain.golden")
	if err != nil {
		t.Fatalf("failed to read current golden output: %v", err)
	}
	currentStr := string(currentBytes)

	// They should be different before normalization
	if originalStr == currentStr {
		t.Errorf("Expected original and current to differ, but they are identical")
	}

	// Normalize both
	origNorm := normalizeDuration(originalStr)
	currNorm := normalizeDuration(currentStr)

	// They should be identical after normalization
	if origNorm != currNorm {
		t.Errorf("Outputs differ outside of duration fields.\nOrig Normalized:\n%s\n\nCurr Normalized:\n%s", origNorm, currNorm)
	}
}
