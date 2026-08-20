package runner

import (
	"io"
	"testing"
)

func TestReportRenderSortsCopy(t *testing.T) {
	opts := RenderOptions{
		Out: io.Discard,
		Rows: []Row{
			{Identity: BindingIdentity{RecordPath: "z.md", BindingIndex: 0}},
			{Identity: BindingIdentity{RecordPath: "a.md", BindingIndex: 0}},
			{Identity: BindingIdentity{RecordPath: "m.md", BindingIndex: 0}},
		},
	}

	// Make a copy of the identities to check against later
	original := make([]BindingIdentity, len(opts.Rows))
	for i, r := range opts.Rows {
		original[i] = r.Identity
	}

	err := Render(opts)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Assert the original slice is unmodified
	for i, r := range opts.Rows {
		if r.Identity != original[i] {
			t.Errorf("Render mutated opts.Rows! Index %d is now %v, expected %v", i, r.Identity.RecordPath, original[i].RecordPath)
		}
	}
}
