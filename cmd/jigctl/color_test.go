package main

import (
	"testing"
)

func TestShouldEnableColor(t *testing.T) {
	tests := []struct {
		name     string
		inputs   colorDecisionInputs
		expected bool
	}{
		{
			name: "default enabled",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv:  func(k string) (string, bool) { return "", false },
			},
			expected: true,
		},
		{
			name: "disabled if not terminal",
			inputs: colorDecisionInputs{
				IsTerminal: false,
				LookupEnv:  func(k string) (string, bool) { return "", false },
			},
			expected: false,
		},
		{
			name: "disabled by --no-color flag",
			inputs: colorDecisionInputs{
				IsTerminal:  true,
				FlagNoColor: true,
				LookupEnv:   func(k string) (string, bool) { return "", false },
			},
			expected: false,
		},
		{
			name: "disabled by --plain flag",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				FlagPlain:  true,
				LookupEnv:  func(k string) (string, bool) { return "", false },
			},
			expected: false,
		},
		{
			name: "disabled by NO_COLOR=1",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "NO_COLOR" {
						return "1", true
					}
					return "", false
				},
			},
			expected: false,
		},
		{
			name: "disabled by NO_COLOR=0 (value-insensitive)",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "NO_COLOR" {
						return "0", true
					}
					return "", false
				},
			},
			expected: false,
		},
		{
			name: "enabled if NO_COLOR is present but empty",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "NO_COLOR" {
						return "", true
					}
					return "", false
				},
			},
			expected: true,
		},
		{
			name: "disabled by TERM=dumb",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "TERM" {
						return "dumb", true
					}
					return "", false
				},
			},
			expected: false,
		},
		{
			name: "enabled if TERM is not dumb",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "TERM" {
						return "xterm-256color", true
					}
					return "", false
				},
			},
			expected: true,
		},
		{
			name: "disabled by JIGCTL_NO_COLOR=1",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "JIGCTL_NO_COLOR" {
						return "1", true
					}
					return "", false
				},
			},
			expected: false,
		},
		{
			name: "disabled by JIGCTL_NO_COLOR (set but empty)",
			inputs: colorDecisionInputs{
				IsTerminal: true,
				LookupEnv: func(k string) (string, bool) {
					if k == "JIGCTL_NO_COLOR" {
						return "", true
					}
					return "", false
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldEnableColor(tt.inputs)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
