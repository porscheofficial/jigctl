package hcr

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

// Plan is a validated tree ready for execution. Root is the only base used to
// resolve authored paths, regardless of target scope.
type Plan struct {
	Root    string
	Targets []Target
}

// Target records binding provenance without defining a resolution base.
type Target struct {
	Kind     string
	Path     string
	Bindings []ExecutableBinding
}

// Exception is the execution data for a record exception.
type Exception struct {
	Scope string
}

// ExecutableBinding is the validated execution data for one record binding.
type ExecutableBinding struct {
	RecordID     string
	Title        string
	Summary      string
	RecordPath   string
	BindingIndex int
	Kind         string
	State        string
	Severity     string
	Cadence      []string
	Ref          string
	Run          string
	Docs         string
	File         string
	Path         string
	Op           string
	Value        interface{}
	Require      []string
	Forbid       []string
	TimeoutSecs  int
	Tool         string
	Pattern      string
	Select       string
	Exceptions   []Exception
}

// ExecutionPlan validates the complete tree before exposing any executable
// binding. currentDate is threaded into validation as the invocation clock.
func ExecutionPlan(root string, currentDate time.Time) (Plan, []Diagnostic, error) {
	canonicalRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return Plan{}, nil, fmt.Errorf("canonicalize tree root %s: %w", root, err)
	}
	diagnostics, err := validateTreeAt(canonicalRoot, currentDate)
	if err != nil {
		return Plan{}, nil, err
	}
	if len(diagnostics) != 0 {
		return Plan{}, diagnostics, nil
	}

	services, err := discoverServices(canonicalRoot)
	if err != nil {
		return Plan{}, nil, err
	}
	targets := make([]Target, 1, len(services)+1)
	targets[0] = Target{Kind: "repo"}
	targetByService := make(map[string]int, len(services))
	for _, service := range services {
		servicePath, relativeErr := filepath.Rel(canonicalRoot, service)
		if relativeErr != nil {
			return Plan{}, nil, fmt.Errorf("relativize service %s: %w", service, relativeErr)
		}
		targetByService[service] = len(targets)
		targets = append(targets, Target{Kind: "service", Path: servicePath})
	}

	records, err := indexTree(canonicalRoot)
	if err != nil {
		return Plan{}, nil, err
	}
	for _, record := range records {
		var meta parsedMeta
		frontmatter, _, _ := extractFrontmatter(record.source)
		if decodeErr := yaml.Unmarshal(frontmatter, &meta); decodeErr != nil {
			return Plan{}, nil, fmt.Errorf("decode validated frontmatter %s: %w", record.path, decodeErr)
		}
		resolved := resolveBindings(meta.State, meta.EnforcedBy)
		targetIndex := 0
		if record.service != "" {
			targetIndex = targetByService[record.service]
		}
		for bindingIndex := range meta.EnforcedBy {
			executable, bindingErr := executableBinding(bindingIdentity{
				recordID: meta.ID, title: meta.Title, summary: meta.Summary,
				recordPath: record.path, bindingIndex: bindingIndex,
			}, &meta.EnforcedBy[bindingIndex], resolved[bindingIndex], meta.Exceptions)
			if bindingErr != nil {
				return Plan{}, nil, bindingErr
			}
			targets[targetIndex].Bindings = append(targets[targetIndex].Bindings,
				executable)
		}
	}
	return Plan{Root: canonicalRoot, Targets: targets}, nil, nil
}

type bindingIdentity struct {
	recordID     string
	title        string
	summary      string
	recordPath   string
	bindingIndex int
}

func executableBinding(
	identity bindingIdentity,
	binding *parsedBinding,
	resolved ResolvedBinding,
	parsedExceptions []parsedException,
) (ExecutableBinding, error) {
	ref, refErr := executionRef(binding.Ref)
	if refErr != nil {
		return ExecutableBinding{}, fmt.Errorf(
			"decode validated binding %s[%d]: %w",
			identity.recordPath,
			identity.bindingIndex,
			refErr,
		)
	}
	var exceptions []Exception
	if len(parsedExceptions) > 0 {
		exceptions = make([]Exception, len(parsedExceptions))
		for i, pe := range parsedExceptions {
			exceptions[i] = Exception{Scope: pe.Scope}
		}
	}
	return ExecutableBinding{
		RecordID:     identity.recordID,
		Title:        identity.title,
		Summary:      identity.summary,
		RecordPath:   identity.recordPath,
		BindingIndex: identity.bindingIndex,
		Kind:         resolved.Kind,
		State:        resolved.State,
		Severity:     resolved.Severity,
		Cadence:      resolved.Cadence,
		Ref:          ref, Run: binding.Run, Docs: binding.Docs, File: binding.File,
		Path: binding.Path, Op: binding.Op, Value: binding.Value,
		Require: append([]string(nil), binding.Require...), Forbid: append([]string(nil), binding.Forbid...),
		TimeoutSecs: binding.TimeoutSecs, Tool: binding.Tool, Pattern: binding.Pattern, Select: binding.Select,
		Exceptions: exceptions,
	}, nil
}

func executionRef(value interface{}) (string, error) {
	switch ref := value.(type) {
	case nil:
		return "", nil
	case string:
		return ref, nil
	default:
		return "", fmt.Errorf("ref has type %T instead of string", value)
	}
}
