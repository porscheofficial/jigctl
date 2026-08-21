package hcr

import (
	"testing"
)

func TestExecutionPlan_body_preservation(t *testing.T) {
	// Given
	root := t.TempDir()

	writePlanFixture(t, root, "service_globs = []\n", map[string]string{
		".hcr/HCR-9001-normal.md": commandRecord(commandFixture{
			id: "HCR-9001", scope: "repo", ref: "normal", run: "go test",
		}) + "Guidance.\n\n",
		".hcr/HCR-9002-spaces.md": commandRecord(commandFixture{
			id: "HCR-9002", scope: "repo", ref: "spaces", run: "go test",
		}) + "Guidance.  \n",
		".hcr/HCR-9003-indent.md": commandRecord(commandFixture{
			id: "HCR-9003", scope: "repo", ref: "indent", run: "go test",
		}) + "    indented block\n",
		".hcr/HCR-9004-empty.md": commandRecord(commandFixture{
			id: "HCR-9004", scope: "repo", ref: "empty", run: "go test",
		}),
		".hcr/HCR-9005-crlf.md": "---\r\n" +
			"id: HCR-9005\r\n" +
			"title: Execute a validated command\r\n" +
			"scope: repo\r\n" +
			"regulates: reliability\r\n" +
			"summary: The command is validated before it can execute.\r\n" +
			"state: enforced\r\n" +
			"enforced_by:\r\n" +
			"  - kind: command\r\n" +
			"    severity: blocking\r\n" +
			"    cadence: [ci]\r\n" +
			"    ref: crlf\r\n" +
			"    run: go test\r\n" +
			"---\r\n" +
			"Line1\r\nLine2\r\n",
	})

	// When
	plan, diagnostics, err := ExecutionPlan(root, planTestDate)

	// Then
	mustNoError(t, err)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}

	bodies := make(map[string]string)
	for _, binding := range plan.Targets[0].Bindings {
		bodies[binding.RecordID] = binding.Body
	}

	// Assertions
	if bodies["HCR-9001"] != "Guidance.\n\n" {
		t.Errorf("expected 'Guidance.\\n\\n', got %q", bodies["HCR-9001"])
	}
	if bodies["HCR-9002"] != "Guidance.  \n" {
		t.Errorf("expected 'Guidance.  \\n', got %q", bodies["HCR-9002"])
	}
	if bodies["HCR-9003"] != "    indented block\n" {
		t.Errorf("expected '    indented block\\n', got %q", bodies["HCR-9003"])
	}
	if bodies["HCR-9004"] != "" {
		t.Errorf("expected empty string, got %q", bodies["HCR-9004"])
	}
	if bodies["HCR-9005"] != "Line1\nLine2\n" {
		t.Errorf("expected 'Line1\\nLine2\\n', got %q", bodies["HCR-9005"])
	}
}
