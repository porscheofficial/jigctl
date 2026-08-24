package runner

import (
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestParseCadenceSet(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		supplied     bool
		wantErr      bool
		wantExplicit bool
		bindings     []*hcr.ExecutableBinding
		wantSelects  []bool
	}{
		{
			name:         `("", false) -> DefaultCadenceSet`,
			raw:          "",
			supplied:     false,
			wantErr:      false,
			wantExplicit: false,
			bindings: []*hcr.ExecutableBinding{
				{Cadence: []string{"ci"}},
				{Cadence: []string{"on-change"}},
				{Cadence: []string{"scheduled"}},
				{Cadence: []string{}},
			},
			wantSelects: []bool{true, true, false, false},
		},
		{
			name:     `("", true) -> error`,
			raw:      "",
			supplied: true,
			wantErr:  true,
		},
		{
			name:         `("all", true)`,
			raw:          "all",
			supplied:     true,
			wantErr:      false,
			wantExplicit: true,
			bindings: []*hcr.ExecutableBinding{
				{Cadence: []string{"ci"}},
				{Cadence: []string{"on-change"}},
				{Cadence: []string{"scheduled"}},
				{Cadence: []string{"production"}},
				{Cadence: []string{}},
			},
			wantSelects: []bool{true, true, true, true, false},
		},
		{
			name:         `("ci", true)`,
			raw:          "ci",
			supplied:     true,
			wantErr:      false,
			wantExplicit: true,
			bindings: []*hcr.ExecutableBinding{
				{Cadence: []string{"ci"}},
				{Cadence: []string{"on-change"}},
				{Cadence: []string{}},
			},
			wantSelects: []bool{true, false, false},
		},
		{
			name:         `("on-change,ci", true)`,
			raw:          "on-change,ci",
			supplied:     true,
			wantErr:      false,
			wantExplicit: true,
			bindings: []*hcr.ExecutableBinding{
				{Cadence: []string{"ci"}},
				{Cadence: []string{"on-change"}},
				{Cadence: []string{"scheduled"}},
			},
			wantSelects: []bool{true, true, false},
		},
		{
			name:         `("scheduled", true)`,
			raw:          "scheduled",
			supplied:     true,
			wantErr:      false,
			wantExplicit: true,
			bindings: []*hcr.ExecutableBinding{
				{Cadence: []string{"ci"}},
				{Cadence: []string{"scheduled"}},
			},
			wantSelects: []bool{false, true},
		},
		{
			name:     `("all,ci", true) -> error`,
			raw:      "all,ci",
			supplied: true,
			wantErr:  true,
		},
		{
			name:     `("nightly", true) -> error`,
			raw:      "nightly",
			supplied: true,
			wantErr:  true,
		},
		{
			name:     `("ci,", true) -> error`,
			raw:      "ci,",
			supplied: true,
			wantErr:  true,
		},
		{
			name:     `(",", true) -> error`,
			raw:      ",",
			supplied: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCadenceSet(tt.raw, tt.supplied)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCadenceSet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.Explicit() != tt.wantExplicit {
				t.Errorf("Explicit() = %v, want %v", got.Explicit(), tt.wantExplicit)
			}
			for i, b := range tt.bindings {
				if got.Selects(b) != tt.wantSelects[i] {
					t.Errorf("Selects(%v) = %v, want %v", b.Cadence, got.Selects(b), tt.wantSelects[i])
				}
			}
		})
	}
}
