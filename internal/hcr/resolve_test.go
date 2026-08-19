package hcr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/patricebouillet/jigctl/schema"
)

type resolverSchema struct {
	Properties struct {
		EnforcedBy struct {
			Items struct {
				Properties struct {
					Kind struct {
						Enum []string `json:"enum"`
					} `json:"kind"`
				} `json:"properties"`
				AllOf []struct {
					If struct {
						Properties struct {
							Kind struct {
								Const string `json:"const"`
							} `json:"kind"`
						} `json:"properties"`
					} `json:"if"`
					Then struct {
						Properties struct {
							Severity struct {
								Default string `json:"default"`
							} `json:"severity"`
							Cadence struct {
								Default []string `json:"default"`
							} `json:"cadence"`
						} `json:"properties"`
					} `json:"then"`
				} `json:"allOf"`
			} `json:"items"`
		} `json:"enforced_by"`
	} `json:"properties"`
}

func TestBindingDefaults_match_embedded_schema(t *testing.T) {
	// Given
	var document resolverSchema
	if err := json.Unmarshal(schema.HCR, &document); err != nil {
		t.Fatalf("decode embedded schema: %v", err)
	}
	items := document.Properties.EnforcedBy.Items
	schemaDefaults := make(map[string]bindingDefault, len(items.AllOf))
	for i := range items.AllOf {
		branch := &items.AllOf[i]
		schemaDefaults[branch.If.Properties.Kind.Const] = bindingDefault{
			Severity: branch.Then.Properties.Severity.Default,
			Cadence:  branch.Then.Properties.Cadence.Default,
		}
	}

	// When / Then
	if !reflect.DeepEqual(kindDefaults, schemaDefaults) {
		t.Errorf("Go kind defaults differ from schema: Go=%v schema=%v", kindDefaults, schemaDefaults)
	}
	for _, kind := range items.Properties.Kind.Enum {
		if _, exists := kindDefaults[kind]; !exists {
			t.Errorf("schema kind %q is missing from Go kind defaults", kind)
		}
	}
}

func TestResolveBindings_injects_each_kind_default(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want ResolvedBinding
	}{
		{"command", "command", ResolvedBinding{Kind: "command", Severity: "blocking", Cadence: []string{"on-change", "ci"}}},
		{"config-assert", "config-assert", ResolvedBinding{Kind: "config-assert", Severity: "blocking", Cadence: []string{"on-change", "ci"}}},
		{"grep", "grep", ResolvedBinding{Kind: "grep", Severity: "blocking", Cadence: []string{"on-change", "ci"}}},
		{"external", "external", ResolvedBinding{Kind: "external", Severity: "blocking", Cadence: []string{"on-change", "ci"}}},
		{"agent-review", "agent-review", ResolvedBinding{Kind: "agent-review", Severity: "advisory", Cadence: []string{"scheduled"}}},
		{"inferential has no cadence default", "inferential", ResolvedBinding{Kind: "inferential", Severity: "advisory"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			got := resolveBindings("active", []parsedBinding{{Kind: test.kind}})

			// Then
			if len(got) != 1 || !reflect.DeepEqual(got[0], test.want) {
				t.Errorf("resolveBindings() = %v, want [%v]", got, test.want)
			}
		})
	}
}

func TestResolveBindings_preserves_explicit_values_when_not_warn(t *testing.T) {
	// Given
	bindings := []parsedBinding{{Kind: "command", Severity: "advisory", Cadence: []string{"production"}}}

	// When
	got := resolveBindings("active", bindings)

	// Then
	want := []ResolvedBinding{{Kind: "command", Severity: "advisory", Cadence: []string{"production"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveBindings() = %v, want %v", got, want)
	}
}

func TestResolveBindings_warn_overrides_explicit_blocking_and_preserves_cadence(t *testing.T) {
	// Given
	bindings := []parsedBinding{{Kind: "command", Severity: "blocking", Cadence: []string{"production"}}}

	// When
	got := resolveBindings("warn", bindings)

	// Then
	want := []ResolvedBinding{{Kind: "command", Severity: "advisory", Cadence: []string{"production"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveBindings() = %v, want %v", got, want)
	}
}

func TestResolveBindings_unknown_kind_preserves_declared_values(t *testing.T) {
	// Given
	bindings := []parsedBinding{{Kind: "future", Severity: "custom", Cadence: []string{"manual"}}}

	// When
	got := resolveBindings("active", bindings)

	// Then
	want := []ResolvedBinding{{Kind: "future", Severity: "custom", Cadence: []string{"manual"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolveBindings() = %v, want %v", got, want)
	}
}

func TestResolveRecord_downgrades_real_warn_record(t *testing.T) {
	// Given
	path := filepath.Join(corpusRoot, "records", "valid-warn-severity-coverage-bootstrap.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read corpus record: %v", err)
	}

	// When
	got, err := ResolveRecord(source)
	// Then
	if err != nil {
		t.Fatalf("ResolveRecord failed: %v", err)
	}
	if len(got) != 1 || got[0].Severity != "advisory" {
		t.Errorf("ResolveRecord() = %v, want one advisory binding", got)
	}
}

func TestResolveBindings_does_not_alias_the_default_table(t *testing.T) {
	// Given
	first := resolveBindings("active", []parsedBinding{{Kind: "command"}})

	// When
	first[0].Cadence[0] = "mutated"

	// Then
	second := resolveBindings("active", []parsedBinding{{Kind: "command"}})
	want := []string{"on-change", "ci"}
	if !reflect.DeepEqual(second[0].Cadence, want) {
		t.Errorf("later resolution saw %v, want %v: the default table was corrupted", second[0].Cadence, want)
	}
}
