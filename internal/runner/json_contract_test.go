package runner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func createJSONFixture() (*hcr.Plan, []*Verdict) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "A.md", BindingIndex: 0, RecordID: "HCR-1001", Title: "Title A", Summary: "Check A", Body: "This is a body", State: "enforced", Kind: "command"},
					{RecordPath: "C.md", BindingIndex: 0, RecordID: "HCR-1003", Title: "Title C", Summary: "Check C", State: "enforced", Kind: "external"},
				},
			},
			{
				Kind: "service",
				Path: "svc1",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "B.md", BindingIndex: 0, RecordID: "HCR-1002", Title: "Title B", Summary: "Check B", State: "enforced", Kind: "grep"},
				},
			},
		},
	}

	verdicts := []*Verdict{
		// passing binding
		NewCompletedVerdict(&VerdictReport{
			Identity:  BindingIdentity{RecordPath: "A.md", BindingIndex: 0},
			Kind:      "command",
			Severity:  "blocking",
			Target:    TargetProvenance{Name: "repo", Path: ""},
			Execution: &Execution{Argv: []string{"ls", "-la"}, ExitCode: 0, Duration: 42 * time.Millisecond},
		}),
		// expected-unchecked binding
		NewNotAttemptedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "C.md", BindingIndex: 0},
			Kind:     "external",
			Severity: "advisory",
			Target:   TargetProvenance{Name: "repo", Path: ""},
		}, ReasonKindNotExecutable),
		// violating binding with exactly one finding, target kind service
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "B.md", BindingIndex: 0},
			Kind:     "grep",
			Severity: "blocking",
			Target:   TargetProvenance{Name: "service", Path: "svc1"},
			Findings: []Finding{{Locus: Locus{File: "b.txt", Pointer: "L10"}, Severity: "blocking"}},
		}),
	}
	return plan, verdicts
}

func TestJSONGolden(t *testing.T) {
	plan, verdicts := createJSONFixture()

	var buf bytes.Buffer
	err := RenderJSON(&buf, JSONOptions{
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	goldenFile := filepath.Join("testdata", "json.golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err = os.WriteFile(goldenFile, buf.Bytes(), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("RenderJSON output does not match golden file.\nGot:\n%s\nExpected:\n%s", buf.String(), string(expected))
	}
}

func TestJSONSchemaValidation(t *testing.T) {
	checkJSONSchema, err := exec.LookPath("check-jsonschema")
	if err != nil {
		t.Skip("check-jsonschema not installed")
	}

	plan, verdicts := createJSONFixture()

	var buf bytes.Buffer
	err = RenderJSON(&buf, JSONOptions{
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "output.json")
	if writeErr := os.WriteFile(tmpFile, buf.Bytes(), 0o644); writeErr != nil {
		t.Fatalf("Failed to write temp JSON file: %v", writeErr)
	}

	// Make sure we pass the absolute path to schemafile
	cmd := exec.CommandContext(context.Background(), checkJSONSchema, "--schemafile", "../../schema/run-output-v1.schema.json", tmpFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Schema validation failed: %v\nOutput:\n%s", err, string(out))
	}
}

func TestJSONDeterminism(t *testing.T) {
	plan, verdicts := createJSONFixture()

	hashes := make(map[string]struct{})
	for i := 0; i < 5; i++ {
		var buf bytes.Buffer
		err := RenderJSON(&buf, JSONOptions{
			Rows:              BuildRows(plan, verdicts),
			NormalizeDuration: true,
		})
		if err != nil {
			t.Fatalf("RenderJSON failed: %v", err)
		}
		hash := sha256.Sum256(buf.Bytes())
		hexHash := hex.EncodeToString(hash[:])
		hashes[hexHash] = struct{}{}
	}

	if len(hashes) != 1 {
		t.Errorf("expected exactly 1 distinct hash across 5 renders, got %d", len(hashes))
	}
}

// TestJSONDraftBindingExecutionNull verifies that a state: draft command
// binding, which is never executed, serializes with an explicit JSON
// "execution": null rather than omitting the field entirely.
func TestJSONDraftBindingExecutionNull(t *testing.T) {
	plan := &hcr.Plan{
		Targets: []hcr.Target{
			{
				Kind: "repo",
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "D.md", BindingIndex: 0, RecordID: "HCR-1004", Title: "Title D", Summary: "Check D", State: "draft", Kind: "command"},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewNotAttemptedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "D.md", BindingIndex: 0},
			Kind:     "command",
			Severity: "advisory",
			Target:   TargetProvenance{Name: "repo", Path: ""},
		}, ReasonKindNotExecutable),
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, JSONOptions{
		Rows:              BuildRows(plan, verdicts),
		NormalizeDuration: true,
	})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	assertDraftBindingTyped(t, buf.Bytes())
	assertDraftBindingRawExecutionNull(t, buf.Bytes())
}

// assertDraftBindingTyped checks the typed JSONDocument decode of a draft
// binding: exactly one record/binding, correct state, and a nil Execution.
func assertDraftBindingTyped(t *testing.T, data []byte) {
	t.Helper()

	var doc JSONDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if len(doc.Records) != 1 || len(doc.Records[0].Bindings) != 1 {
		t.Fatalf("expected exactly 1 record with 1 binding, got %+v", doc.Records)
	}
	if doc.Records[0].State != "draft" {
		t.Fatalf("expected record state %q, got %q", "draft", doc.Records[0].State)
	}
	if doc.Records[0].Bindings[0].Execution != nil {
		t.Errorf("expected Execution to unmarshal to nil for a draft binding, got %+v", doc.Records[0].Bindings[0].Execution)
	}
}

// assertDraftBindingRawExecutionNull confirms the "execution" key is present
// AND explicitly null in the raw JSON, not merely absent from the output.
func assertDraftBindingRawExecutionNull(t *testing.T, data []byte) {
	t.Helper()

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw JSON: %v", err)
	}
	records, ok := raw["records"].([]interface{})
	if !ok || len(records) != 1 {
		t.Fatalf("expected 1 raw record, got %+v", raw["records"])
	}
	record, ok := records[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected record to be a JSON object, got %T", records[0])
	}
	bindings, ok := record["bindings"].([]interface{})
	if !ok || len(bindings) != 1 {
		t.Fatalf("expected 1 raw binding, got %+v", record["bindings"])
	}
	binding, ok := bindings[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected binding to be a JSON object, got %T", bindings[0])
	}
	value, present := binding["execution"]
	if !present {
		t.Fatalf("expected \"execution\" key to be present in binding JSON, but it was absent: %+v", binding)
	}
	if value != nil {
		t.Errorf("expected \"execution\" to be JSON null, got %+v", value)
	}
}
