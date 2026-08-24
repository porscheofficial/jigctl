package runner

import (
	"testing"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

func TestRowBody(t *testing.T) {
	plan := &hcr.Plan{
		Root: "/root",
		Targets: []hcr.Target{
			{
				Path: "/root/svc",
				Kind: "service",
				Bindings: []hcr.ExecutableBinding{
					{
						RecordPath:   "/root/.hcr/test.md",
						BindingIndex: 0,
						Body:         "Hello World\n\nThis is a body.",
					},
				},
			},
		},
	}

	builder := NewRowBuilder(plan)

	rep := VerdictReport{
		Identity: BindingIdentity{RecordPath: "/root/.hcr/test.md", BindingIndex: 0},
	}
	verdict := NewCompletedVerdict(&rep)

	row := builder.Row(verdict)
	if row.Body != "Hello World\n\nThis is a body." {
		t.Errorf("expected Body %q, got %q", "Hello World\n\nThis is a body.", row.Body)
	}

	rows := []Row{row}
	records := GroupRecords(rows)

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Body != "Hello World\n\nThis is a body." {
		t.Errorf("expected record Body %q, got %q", "Hello World\n\nThis is a body.", records[0].Body)
	}
}
