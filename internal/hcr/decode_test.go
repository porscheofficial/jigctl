package hcr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestParsedMeta_decodes_execution_fields(t *testing.T) {
	// Given
	source := []byte(`---
id: HCR-9001
title: Decode every execution field
summary: Preserve the complete executable binding for later evaluation.
enforced_by:
  - kind: external
    file: config.json
    path: /enabled
    op: equals
    value: true
    require: [present]
    forbid: [absent]
    timeout_secs: 17
    tool: policy-checker
    pattern: literal
    select: changed
---
`)
	frontmatter, ok := extractFrontmatter(source)
	if !ok {
		t.Fatal("frontmatter was not extracted")
	}

	// When
	var meta parsedMeta
	mustNoError(t, yaml.Unmarshal(frontmatter, &meta))
	binding := meta.EnforcedBy[0]

	// Then
	tests := []struct {
		name string
		want any
		got  any
	}{
		{name: "title", want: "Decode every execution field", got: meta.Title},
		{name: "summary", want: "Preserve the complete executable binding for later evaluation.", got: meta.Summary},
		{name: "file", want: "config.json", got: binding.File},
		{name: "path", want: "/enabled", got: binding.Path},
		{name: "op", want: "equals", got: binding.Op},
		{name: "value", want: true, got: binding.Value},
		{name: "require", want: []string{"present"}, got: binding.Require},
		{name: "forbid", want: []string{"absent"}, got: binding.Forbid},
		{name: "timeout_secs", want: 17, got: binding.TimeoutSecs},
		{name: "tool", want: "policy-checker", got: binding.Tool},
		{name: "pattern", want: "literal", got: binding.Pattern},
		{name: "select", want: "changed", got: binding.Select},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mustDeepEqual(t, test.want, test.got)
		})
	}
}

func TestValidateTree_non_string_ref_preserves_identity(t *testing.T) {
	// Given
	root := t.TempDir()
	mustNoError(t, os.Mkdir(filepath.Join(root, ".hcr"), 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(root, "jig.toml"), []byte("service_globs = []\n"), 0o600))
	mustNoError(t, os.WriteFile(filepath.Join(root, ".hcr", "HCR-9001-invalid-ref.md"), []byte(`---
id: HCR-9001
title: Invalid list reference
scope: repo
regulates: reliability
summary: This record remains identifiable even when its binding reference is invalid.
state: enforced
enforced_by:
  - kind: command
    severity: blocking
    cadence: [ci]
    ref: [shared, check]
    run: go test ./...
---
`), 0o600))
	mustNoError(t, os.WriteFile(filepath.Join(root, ".hcr", "HCR-9002-superseding.md"), []byte(`---
id: HCR-9002
title: Superseding record
scope: repo
regulates: reliability
summary: This valid record supersedes the schema-invalid target by its preserved identity.
state: draft
supersedes: HCR-9001
---
`), 0o600))

	// When
	diagnostics, err := ValidateTree(root)
	mustNoError(t, err)

	// Then
	foundSchemaDiagnostic := false
	for _, diagnostic := range diagnostics {
		if filepath.Base(diagnostic.File) == "HCR-9001-invalid-ref.md" && diagnostic.Code == "schema" {
			foundSchemaDiagnostic = true
		}
		if filepath.Base(diagnostic.File) == "HCR-9002-superseding.md" && diagnostic.Code == "R-102" {
			t.Fatalf("non-string ref dropped HCR-9001 from identityIndex: %#v", diagnostic)
		}
	}
	if !foundSchemaDiagnostic {
		t.Fatal("invalid ref did not produce a schema diagnostic")
	}
}
