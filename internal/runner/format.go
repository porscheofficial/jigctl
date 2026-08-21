package runner

import "fmt"

// Format selects how run results are rendered.
type Format int

const (
	FormatHuman Format = iota
	FormatPlain
	FormatJSON
)

// ParseFormat converts a user-supplied string into a Format. Unrecognised
// values are a named error, never a silent fallback.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "human":
		return FormatHuman, nil
	case "plain":
		return FormatPlain, nil
	case "json":
		return FormatJSON, nil
	default:
		return 0, fmt.Errorf("unknown format %q: valid formats are human, plain, json", s)
	}
}
