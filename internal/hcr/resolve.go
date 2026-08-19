package hcr

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
)

var errResolveFrontmatterAbsent = errors.New("frontmatter absent")

type bindingDefault struct {
	Severity string
	Cadence  []string
}

var kindDefaults = map[string]bindingDefault{
	"command":       {Severity: "blocking", Cadence: []string{"on-change", "ci"}},
	"config-assert": {Severity: "blocking", Cadence: []string{"on-change", "ci"}},
	"grep":          {Severity: "blocking", Cadence: []string{"on-change", "ci"}},
	"external":      {Severity: "blocking", Cadence: []string{"on-change", "ci"}},
	"agent-review":  {Severity: "advisory", Cadence: []string{"scheduled"}},
	"inferential":   {Severity: "advisory"},
}

// ResolvedBinding is the effective severity and cadence after kind-tier defaults and the warn downgrade.
type ResolvedBinding struct {
	Kind     string
	Severity string
	Cadence  []string
}

// ResolveRecord resolves effective binding values from a record's frontmatter.
func ResolveRecord(source []byte) ([]ResolvedBinding, error) {
	frontmatter, present := extractFrontmatter(source)
	if !present {
		return nil, fmt.Errorf("resolve record: %w", errResolveFrontmatterAbsent)
	}
	var meta parsedMeta
	if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
		return nil, fmt.Errorf("decode record frontmatter: %w", err)
	}
	return resolveBindings(meta.State, meta.EnforcedBy), nil
}

func resolveBindings(state string, bindings []parsedBinding) []ResolvedBinding {
	resolved := make([]ResolvedBinding, 0, len(bindings))
	for i := range bindings {
		binding := &bindings[i]
		severity := binding.Severity
		// Cloned deliberately: handing back kindDefaults' own backing array lets a
		// caller mutating a resolved value corrupt the table for every later call.
		cadence := slices.Clone(binding.Cadence)
		if defaults, recognized := kindDefaults[binding.Kind]; recognized {
			if severity == "" {
				severity = defaults.Severity
			}
			if len(cadence) == 0 {
				cadence = slices.Clone(defaults.Cadence)
			}
		}
		// Only warn downgrades severity. Draft and deprecated have no R-105 behavior.
		if state == "warn" {
			severity = "advisory"
		}
		// Unknown kinds arise only after schema failure; preserve their declared values.
		resolved = append(resolved, ResolvedBinding{Kind: binding.Kind, Severity: severity, Cadence: cadence})
	}
	return resolved
}
