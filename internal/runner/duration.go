package runner

import (
	"fmt"
	"strings"
	"time"
)

// formatDuration formats a duration to at most 3 significant digits.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	var unit string
	var val float64

	// Scale and pick unit
	switch {
	case d >= time.Hour-time.Minute/2:
		unit = "h"
		val = float64(d) / float64(time.Hour)
	case d >= time.Minute-time.Second/2:
		unit = "m"
		val = float64(d) / float64(time.Minute)
	case d >= time.Second-time.Millisecond/2:
		unit = "s"
		val = float64(d) / float64(time.Second)
	case d >= time.Millisecond-time.Microsecond/2:
		unit = "ms"
		val = float64(d) / float64(time.Millisecond)
	case d >= time.Microsecond-time.Nanosecond/2:
		unit = "µs"
		val = float64(d) / float64(time.Microsecond)
	default:
		unit = "ns"
		val = float64(d)
	}

	// Format to at most 3 digits
	var s string
	switch {
	case val >= 99.95:
		s = fmt.Sprintf("%.0f", val)
	case val >= 9.995:
		s = fmt.Sprintf("%.1f", val)
	default:
		s = fmt.Sprintf("%.2f", val)
	}

	// Strip trailing zeroes after decimal point, and the decimal point itself.
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	return s + unit
}

// DurationColumn renders an execution duration right-aligned to 7 runes.
// If exec is nil, it renders 7 spaces.
func DurationColumn(exec *Execution) string {
	if exec == nil {
		return "       "
	}
	return fmt.Sprintf("%7s", formatDuration(exec.Duration))
}
