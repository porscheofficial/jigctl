package hcr

import (
	"regexp"
	"sort"
)

var idPattern = regexp.MustCompile(`^HCR-\d{4}$`)

type parsedBinding struct {
	Kind string `yaml:"kind"`
	// Ref is interface{} rather than string because parsedMeta is decoded for every
	// discovered record, including schema-INVALID ones, to populate identityIndex.
	// If Ref were a string, a record with a non-string ref would fail decoding
	// completely, its id would be lost from identityIndex, and a valid supersedes
	// pointing to it would get a false R-102. It is safely type-asserted where used.
	Ref  interface{} `yaml:"ref"`
	Run  string      `yaml:"run"`
	Docs string      `yaml:"docs"`
}

// parsedMeta holds fields needed for META rules.
type parsedMeta struct {
	ID         string          `yaml:"id"`
	Supersedes string          `yaml:"supersedes"`
	Rationale  string          `yaml:"rationale"`
	EnforcedBy []parsedBinding `yaml:"enforced_by"`
	Scope      string          `yaml:"scope"`
}

type emitterRecord struct {
	path    string
	relPath string
	service string
	meta    parsedMeta
}

// ValidateRecord returns malformed-record diagnostics as data. The error is
// reserved for operational failures, such as schema compilation.
func ValidateRecord(path string, source []byte) ([]Diagnostic, error) {
	diagnostics, err := schemaDiagnostics(path, source)
	if err != nil {
		return nil, err
	}
	sortDiagnostics(diagnostics)
	return diagnostics, nil
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(left, right int) bool {
		a := diagnostics[left]
		b := diagnostics[right]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Pointer != b.Pointer {
			return a.Pointer < b.Pointer
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
}
