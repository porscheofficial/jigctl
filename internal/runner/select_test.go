package runner

import (
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func TestSelect_StateDraft(t *testing.T) {
	report := &VerdictReport{}
	binding := &hcr.ExecutableBinding{State: "draft"}
	v := Select(report, binding)
	if v == nil || v.Completion() != CompletionNotAttempted || v.Reason() != Reason(ReasonRecordDraft) {
		t.Errorf("Expected NotAttempted / ReasonRecordDraft, got %v", v)
	}
}

func TestSelect_StateDeprecated(t *testing.T) {
	report := &VerdictReport{}
	binding := &hcr.ExecutableBinding{State: "deprecated"}
	v := Select(report, binding)
	if v == nil || v.Completion() != CompletionNotAttempted || v.Reason() != Reason(ReasonRecordDeprecated) {
		t.Errorf("Expected NotAttempted / ReasonRecordDeprecated, got %v", v)
	}
}

func TestSelect_StateEnforcedAndWarn(t *testing.T) {
	report := &VerdictReport{}
	for _, state := range []string{"enforced", "warn", "active"} {
		binding := &hcr.ExecutableBinding{
			State:   state,
			Kind:    "command",
			Cadence: []string{"on-change"},
		}
		v := Select(report, binding)
		if v != nil {
			t.Errorf("Expected state %q to be selected, got %v", state, v)
		}
	}
}

func TestSelect_KindNotExecutable(t *testing.T) {
	report := &VerdictReport{}
	for _, kind := range []string{"external", "agent-review", "inferential"} {
		binding := &hcr.ExecutableBinding{
			State:   "active",
			Kind:    kind,
			Cadence: []string{"on-change"}, // would be valid cadence
		}
		v := Select(report, binding)
		if v == nil || v.Completion() != CompletionNotAttempted || v.Reason() != Reason(ReasonKindNotExecutable) {
			t.Errorf("Expected %q to be excluded by kind, got %v", kind, v)
		}
	}
}

func TestSelect_Cadence(t *testing.T) {
	tests := []struct {
		name     string
		cadence  []string
		selected bool
	}{
		{"on-change included", []string{"scheduled", "on-change"}, true},
		{"ci included", []string{"ci"}, true},
		{"only production", []string{"production"}, false},
		{"only scheduled", []string{"scheduled"}, false},
		{"empty cadence", []string{}, false},
	}
	report := &VerdictReport{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := &hcr.ExecutableBinding{
				State:   "active",
				Kind:    "command",
				Cadence: tt.cadence,
			}
			v := Select(report, binding)
			if tt.selected {
				if v != nil {
					t.Errorf("Expected selected, got %v", v)
				}
			} else {
				if v == nil || v.Completion() != CompletionNotAttempted || v.Reason() != Reason(ReasonCadenceExcluded) {
					t.Errorf("Expected CadenceExcluded, got %v", v)
				}
			}
		})
	}
}

func TestSelect_RealRecords(t *testing.T) {
	// Synthetic equivalents of the 10 real dogfood records to show
	// exactly one is excluded by kind and none by cadence if they use defaults.
	tests := []struct {
		name       string
		record     hcr.ExecutableBinding
		wantReason Reason
	}{
		{"HCR-0401", hcr.ExecutableBinding{State: "active", Kind: "config-assert", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0402", hcr.ExecutableBinding{State: "active", Kind: "command", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0403", hcr.ExecutableBinding{State: "active", Kind: "command", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0404", hcr.ExecutableBinding{State: "active", Kind: "grep", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0405", hcr.ExecutableBinding{State: "active", Kind: "grep", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0406", hcr.ExecutableBinding{State: "active", Kind: "command", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0407", hcr.ExecutableBinding{State: "active", Kind: "inferential", Cadence: []string{}}, Reason(ReasonKindNotExecutable)},
		{"HCR-0408", hcr.ExecutableBinding{State: "active", Kind: "command", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0409", hcr.ExecutableBinding{State: "active", Kind: "config-assert", Cadence: []string{"on-change", "ci"}}, ReasonNone},
		{"HCR-0410", hcr.ExecutableBinding{State: "active", Kind: "config-assert", Cadence: []string{"on-change", "ci"}}, ReasonNone},
	}
	report := &VerdictReport{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Select(report, &tt.record)
			if tt.wantReason == ReasonNone {
				if v != nil {
					t.Errorf("Expected %s to be selected, got %v", tt.name, v)
				}
			} else {
				if v == nil || v.Reason() != tt.wantReason {
					t.Errorf("Expected %s to be excluded with reason %v, got %v", tt.name, tt.wantReason, v)
				}
			}
		})
	}
}

func TestSelect_DistinguishableReasons(t *testing.T) {
	if Reason(ReasonRecordDraft) == Reason(ReasonRecordDeprecated) {
		t.Fatal("Expected draft and deprecated reasons to be distinguishable")
	}
}
