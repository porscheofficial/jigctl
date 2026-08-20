package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ApplyExceptions evaluates exceptions against a finding and populates WaivedBy
// for every matching exception. It returns the unmodified finding otherwise.
//
//nolint:gocritic // hugeParam: finding is passed by value to avoid side effects
func ApplyExceptions(
	finding Finding,
	kind string,
	exceptions []string,
	recordPath string,
	knownServicePaths []string,
) (Finding, error) {
	if kind == "command" || kind == "external" || kind == "agent-review" || kind == "inferential" {
		return finding, nil
	}
	if kind != "grep" && kind != "config-assert" {
		return finding, nil
	}

	for i, scope := range exceptions {
		matched := false

		// 1. Service shape: byte-for-byte equal to any discovered service's tree-relative directory
		isServiceShape := contains(knownServicePaths, scope)

		if isServiceShape {
			prefix := scope + string(filepath.Separator)
			if finding.Locus.File == scope || strings.HasPrefix(finding.Locus.File, prefix) {
				matched = true
			}
		} else {
			// 2. Path shape: contains any of `/`, `*`, `?`, `[`, `{`, or `\`
			isPathShaped := strings.ContainsAny(scope, "/*?[{\\")
			if isPathShaped {
				var err error
				matched, err = doublestar.Match(scope, finding.Locus.File)
				if err != nil {
					return finding, fmt.Errorf("doublestar match: %w", err)
				}
			} else {
				// 3. Invalid scope shape
				return finding, &ScopeInvalidError{Scope: scope}
			}
		}

		if matched {
			finding.WaivedBy = append(finding.WaivedBy, ExceptionIdentity{
				RecordPath:     recordPath,
				ExceptionIndex: i,
			})
		}
	}

	return finding, nil
}

// ScopeInvalidError is returned when an exception scope shape is neither
// service-shaped nor path-shaped.
type ScopeInvalidError struct {
	Scope string
}

func (e *ScopeInvalidError) Error() string {
	return "invalid exception scope shape: " + e.Scope
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
