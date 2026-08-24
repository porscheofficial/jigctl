package runner

import (
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestLocator(t *testing.T) {
	plan := &hcr.Plan{
		Root: "/base",
		Targets: []hcr.Target{
			{
				Bindings: []hcr.ExecutableBinding{
					{RecordPath: "/base/single.md", BindingIndex: 0},
					{RecordPath: "/base/multi.md", BindingIndex: 0},
					{RecordPath: "/base/multi.md", BindingIndex: 1},
				},
			},
		},
	}

	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/single.md", BindingIndex: 0},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/multi.md", BindingIndex: 0},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/multi.md", BindingIndex: 1},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/missing.md", BindingIndex: 0},
		}),
	}

	rows := BuildRows(plan, verdicts)

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	// single.md: single binding -> locator is just the relative path
	if rows[0].Locator != "single.md" {
		t.Errorf("expected single.md, got %q", rows[0].Locator)
	}

	// multi.md: multiple bindings -> locator has index
	if rows[1].Locator != "multi.md:0" {
		t.Errorf("expected multi.md:0, got %q", rows[1].Locator)
	}
	if rows[2].Locator != "multi.md:1" {
		t.Errorf("expected multi.md:1, got %q", rows[2].Locator)
	}

	// missing.md: not in plan lookup -> locator unconditionally has index
	if rows[3].Locator != "missing.md:0" {
		t.Errorf("expected missing.md:0, got %q", rows[3].Locator)
	}
}
