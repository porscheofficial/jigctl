package runner

import (
	"testing"
)

func TestTarget(t *testing.T) {
	verdicts := []*Verdict{
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/repo.md", BindingIndex: 0},
			Target:   TargetProvenance{Name: "repo", Path: ""},
		}),
		NewCompletedVerdict(&VerdictReport{
			Identity: BindingIdentity{RecordPath: "/base/service.md", BindingIndex: 0},
			Target:   TargetProvenance{Name: "service", Path: "services/foo"},
		}),
	}

	rows := BuildRows(nil, verdicts)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].TargetKind != "repo" || rows[0].TargetPath != "" {
		t.Errorf("repo row: expected Name=repo Path='', got Name=%q Path=%q", rows[0].TargetKind, rows[0].TargetPath)
	}

	if rows[1].TargetKind != "service" || rows[1].TargetPath != "services/foo" {
		t.Errorf("service row: expected Name=service Path='services/foo', got Name=%q Path=%q", rows[1].TargetKind, rows[1].TargetPath)
	}

	records := GroupRecords(rows)
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	if records[0].TargetKind != "repo" || records[0].TargetPath != "" {
		t.Errorf("repo record: expected Name=repo Path='', got Name=%q Path=%q", records[0].TargetKind, records[0].TargetPath)
	}
	if records[1].TargetKind != "service" || records[1].TargetPath != "services/foo" {
		t.Errorf("service record: expected Name=service Path='services/foo', got Name=%q Path=%q", records[1].TargetKind, records[1].TargetPath)
	}
}
