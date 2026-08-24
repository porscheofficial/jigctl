package hcr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/porscheofficial/jigctl/schema"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

const schemaURL = "mem://hcr.json"

var (
	compileOnce sync.Once
	compiled    *jsonschema.Schema
	compileErr  error
)

func compiledSchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		var document any
		if err := json.Unmarshal(schema.HCR, &document); err != nil {
			compileErr = fmt.Errorf("decode embedded HCR schema: %w", err)
			return
		}
		compiler := jsonschema.NewCompiler()
		// No URL loader is registered: references cannot trigger network access.
		if err := compiler.AddResource(schemaURL, document); err != nil {
			compileErr = fmt.Errorf("add embedded HCR schema: %w", err)
			return
		}
		compiled, compileErr = compiler.Compile(schemaURL)
		if compileErr != nil {
			compileErr = fmt.Errorf("compile embedded HCR schema: %w", compileErr)
		}
	})
	return compiled, compileErr
}

func schemaDiagnostics(path string, source []byte) ([]Diagnostic, error) {
	frontmatter, _, present := extractFrontmatter(source)
	if !present {
		return []Diagnostic{{File: path, Code: "schema", Message: "no YAML frontmatter"}}, nil
	}
	var value any
	if decodeErr := yaml.Unmarshal(frontmatter, &value); decodeErr != nil {
		return malformedRecordDiagnostic(path, decodeErr), nil
	}
	validator, err := compiledSchema()
	if err != nil {
		return nil, err
	}
	err = validator.Validate(value)
	if err == nil {
		return nil, nil
	}
	var validationError *jsonschema.ValidationError
	if !errors.As(err, &validationError) {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	pruned := make([]*jsonschema.ValidationError, 0, 1)
	pruneValidationErrors(validationError, &pruned)
	diagnostics := make([]Diagnostic, 0, len(pruned))
	for _, item := range pruned {
		diagnostics = append(diagnostics, Diagnostic{
			File:    path,
			Pointer: instancePointer(item.InstanceLocation),
			Code:    "schema",
			Message: item.Error(),
		})
	}
	return diagnostics, nil
}

func malformedRecordDiagnostic(path string, err error) []Diagnostic {
	return []Diagnostic{{File: path, Code: "schema", Message: err.Error()}}
}

func pruneValidationErrors(current *jsonschema.ValidationError, output *[]*jsonschema.ValidationError) {
	switch current.ErrorKind.(type) {
	case *kind.AnyOf:
		*output = append(*output, current)
		return
	case *kind.OneOf:
		// The current schema has no oneOf. Keep this defensive arm so a future
		// schema choice retains the same deliberate pruning semantics.
		*output = append(*output, current)
		return
	default:
		if len(current.Causes) == 0 {
			*output = append(*output, current)
			return
		}
		for _, cause := range current.Causes {
			pruneValidationErrors(cause, output)
		}
	}
}

func instancePointer(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	return "/" + strings.Join(segments, "/")
}
