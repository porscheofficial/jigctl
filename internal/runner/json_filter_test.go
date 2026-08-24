package runner

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONOnlyFailures(t *testing.T) {
	rows := []Row{
		{Identity: BindingIdentity{RecordPath: "pass.md", BindingIndex: 0}, RecordID: "HCR-0001-pass", Projection: ProjectionPass, Title: "Pass"},
		{Identity: BindingIdentity{RecordPath: "exp.md", BindingIndex: 0}, RecordID: "HCR-0002-exp", Projection: ProjectionExpectedUnchecked, Title: "Expected"},
		{Identity: BindingIdentity{RecordPath: "block.md", BindingIndex: 0}, RecordID: "HCR-0003-block", Projection: ProjectionBlockedUnchecked, Title: "Blocked"},
		{Identity: BindingIdentity{RecordPath: "oper.md", BindingIndex: 0}, RecordID: "HCR-0004-oper", Projection: ProjectionOperational, Title: "Oper"},
		{Identity: BindingIdentity{RecordPath: "inv.md", BindingIndex: 0}, RecordID: "HCR-0005-inv", Projection: ProjectionInvalid, Title: "Inv"},
		{Identity: BindingIdentity{RecordPath: "viol.md", BindingIndex: 0}, RecordID: "HCR-0006-viol", Projection: ProjectionViolation, Title: "Viol"},
		// A mixed record (violation overall)
		{Identity: BindingIdentity{RecordPath: "mixed.md", BindingIndex: 0}, RecordID: "HCR-0007-mix", Projection: ProjectionPass, Title: "Mix"},
		{Identity: BindingIdentity{RecordPath: "mixed.md", BindingIndex: 1}, RecordID: "HCR-0007-mix", Projection: ProjectionViolation, Title: "Mix"},
	}

	exitCode := AggregateExitCode(ExitSummaries(rows), false)

	optsFull := JSONOptions{Root: ".", Rows: rows, ExitCode: exitCode, OnlyFailures: false}
	optsFiltered := JSONOptions{Root: ".", Rows: rows, ExitCode: exitCode, OnlyFailures: true}

	var bufFull, bufFiltered bytes.Buffer
	if err := RenderJSON(&bufFull, optsFull); err != nil {
		t.Fatalf("RenderJSON full: %v", err)
	}
	if err := RenderJSON(&bufFiltered, optsFiltered); err != nil {
		t.Fatalf("RenderJSON filtered: %v", err)
	}

	var docFull, docFiltered JSONDocument
	if err := json.Unmarshal(bufFull.Bytes(), &docFull); err != nil {
		t.Fatalf("unmarshal full: %v", err)
	}
	if err := json.Unmarshal(bufFiltered.Bytes(), &docFiltered); err != nil {
		t.Fatalf("unmarshal filtered: %v", err)
	}

	t.Run("keeps only actionable records", func(t *testing.T) {
		assertKeptRecords(t, &docFiltered)
	})

	t.Run("summary and exit code are identical", func(t *testing.T) {
		assertIdenticalSummaryAndExitCode(t, &docFull, &docFiltered)
	})

	t.Run("kept record retains all bindings", func(t *testing.T) {
		assertMixedRecordBindings(t, &docFiltered)
	})

	t.Run("all pass input marshals as empty array", func(t *testing.T) {
		allPassRows := []Row{
			{Identity: BindingIdentity{RecordPath: "pass2.md", BindingIndex: 0}, RecordID: "HCR-0010-pass", Projection: ProjectionPass, Title: "Pass 2"},
		}
		optsAllPass := JSONOptions{Root: ".", Rows: allPassRows, ExitCode: AggregateExitCode(ExitSummaries(allPassRows), false), OnlyFailures: true}
		var bufAllPass bytes.Buffer
		if err := RenderJSON(&bufAllPass, optsAllPass); err != nil {
			t.Fatalf("RenderJSON all pass: %v", err)
		}
		if !bytes.Contains(bufAllPass.Bytes(), []byte(`"records": []`)) {
			t.Errorf("expected '\"records\": []' in JSON, got: %s", bufAllPass.String())
		}
	})
}

func assertKeptRecords(t *testing.T, docFiltered *JSONDocument) {
	t.Helper()
	keptRecords := make(map[string]bool)
	for i := range docFiltered.Records {
		keptRecords[docFiltered.Records[i].RecordID] = true
	}

	expectedKept := map[string]bool{
		"HCR-0001-pass":  false,
		"HCR-0002-exp":   false,
		"HCR-0003-block": true,
		"HCR-0004-oper":  true,
		"HCR-0005-inv":   true,
		"HCR-0006-viol":  true,
		"HCR-0007-mix":   true,
	}

	for id, expected := range expectedKept {
		if keptRecords[id] != expected {
			t.Errorf("record %s: expected kept=%v, got %v", id, expected, keptRecords[id])
		}
	}
}

func assertIdenticalSummaryAndExitCode(t *testing.T, docFull, docFiltered *JSONDocument) {
	t.Helper()
	if docFull.Summary.Records != docFiltered.Summary.Records {
		t.Errorf("summary records mismatch")
	}
	if docFull.Summary.Bindings != docFiltered.Summary.Bindings {
		t.Errorf("summary bindings mismatch")
	}
	for k, v := range docFull.Summary.BindingsByProjection {
		if docFiltered.Summary.BindingsByProjection[k] != v {
			t.Errorf("summary bindings_by_projection mismatch")
		}
	}
	if docFull.Summary.UnwaivedFindings != docFiltered.Summary.UnwaivedFindings {
		t.Errorf("summary unwaived_findings mismatch")
	}
	if docFull.ExitCode != docFiltered.ExitCode {
		t.Errorf("exit code mismatch")
	}
}

func assertMixedRecordBindings(t *testing.T, docFiltered *JSONDocument) {
	t.Helper()
	for i := range docFiltered.Records {
		if docFiltered.Records[i].RecordID != "HCR-0007-mix" {
			continue
		}
		if len(docFiltered.Records[i].Bindings) != 2 {
			t.Errorf("expected mixed record to have 2 bindings, got %d", len(docFiltered.Records[i].Bindings))
		}
		hasPass := false
		hasViol := false
		for j := range docFiltered.Records[i].Bindings {
			if docFiltered.Records[i].Bindings[j].Projection == "pass" {
				hasPass = true
			}
			if docFiltered.Records[i].Bindings[j].Projection == "violation" {
				hasViol = true
			}
		}
		if !hasPass || !hasViol {
			t.Errorf("expected mixed record to have 'pass' and 'violation' bindings")
		}
	}
}
