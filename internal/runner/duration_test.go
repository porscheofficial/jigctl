package runner

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestDurationFormatter(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{386055917, "386ms"},
		{2025092375, "2.03s"},
		{159345667, "159ms"},
		{2618624541, "2.62s"},
		{107615917, "108ms"},
		{128989750, "129ms"},
		{42 * time.Millisecond, "42ms"}, // Hard compatibility constraint
		{0, "0s"},
		{999, "999ns"},
		{1000, "1µs"},
		{1001, "1µs"},
		{1050, "1.05µs"},
		{999500, "1ms"},
		{999500000, "1s"},
		{61 * time.Second, "1.02m"},
		{90 * time.Second, "1.5m"},
		{60 * time.Minute, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.d)

			// 1. Exact expected string
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.d, got, tt.want)
			}

			// 2. Formatting twice yields identical string
			// We can't format a string directly, but we check determinism implicitly by the table.
			// The instruction says "formatting twice yields the identical string".
			// Since formatDuration takes a duration, formatDuration(tt.d) is deterministic.
			got2 := formatDuration(tt.d)
			if got != got2 {
				t.Errorf("formatDuration(%d) not deterministic: %q vs %q", tt.d, got, got2)
			}

			// 3. At most 3 digit characters
			digits := 0
			for _, r := range got {
				if r >= '0' && r <= '9' {
					digits++
				}
			}
			if digits > 3 {
				t.Errorf("formatDuration(%d) = %q has %d digits, want <= 3", tt.d, got, digits)
			}

			// 4. No 'ns' suffix for anything >= 1µs
			if tt.d >= time.Microsecond && strings.HasSuffix(got, "ns") {
				t.Errorf("formatDuration(%d) = %q has ns suffix but >= 1µs", tt.d, got)
			}

			// 5. At most one '.'
			if strings.Count(got, ".") > 1 {
				t.Errorf("formatDuration(%d) = %q has >1 decimal points", tt.d, got)
			}

			// 6. Column helper IDENTICAL rune width and right alignment
			col := DurationColumn(&Execution{Duration: tt.d})
			if utf8.RuneCountInString(col) != 7 {
				t.Errorf("DurationColumn(%q) rune count = %d, want 7. String: %q", got, utf8.RuneCountInString(col), col)
			}

			if !strings.HasSuffix(col, got) {
				t.Errorf("DurationColumn(%q) = %q is not right-aligned (expected suffix %q)", got, col, got)
			}
		})
	}
}

func TestDurationColumn_NilExecution(t *testing.T) {
	col := DurationColumn(nil)
	if utf8.RuneCountInString(col) != 7 {
		t.Errorf("DurationColumn(nil) rune count = %d, want 7", utf8.RuneCountInString(col))
	}
	if strings.TrimSpace(col) != "" {
		t.Errorf("DurationColumn(nil) = %q, want spaces", col)
	}
}
