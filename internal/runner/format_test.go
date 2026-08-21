package runner

import (
	"strings"
	"testing"
)

func TestParseFormat(t *testing.T) {
	cases := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"plain", FormatPlain, false},
		{"human", FormatHuman, false},
		{"JSON", 0, true},
		{"jsonl", 0, true},
		{"", 0, true},
	}

	for _, tc := range cases {
		got, err := ParseFormat(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseFormat(%q): expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseFormat(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseFormat(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}

	if _, err := ParseFormat("jsonl"); err == nil {
		t.Fatal("ParseFormat(\"jsonl\"): expected error, got nil")
	} else if want := "human, plain, json"; !strings.Contains(err.Error(), want) {
		t.Errorf("ParseFormat(\"jsonl\") error = %q, want substring %q", err.Error(), want)
	}
}
