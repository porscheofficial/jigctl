package hcr

import "sort"

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
