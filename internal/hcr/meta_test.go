package hcr

import (
	"reflect"
	"testing"
)

func TestR103Determinism(t *testing.T) {
	// Create deliberately unsorted bindings for the same ref
	bindings := []bindingRef{
		{
			file:         "c.md",
			relPath:      "services/c.md",
			bindingIndex: 0,
			run:          "make c",
			ref:          "shared",
		},
		{
			file:         "a.md",
			relPath:      "services/a.md",
			bindingIndex: 1,
			run:          "make a2",
			ref:          "shared",
		},
		{
			file:         "a.md",
			relPath:      "services/a.md",
			bindingIndex: 0,
			run:          "make a1",
			ref:          "shared",
		},
	}

	var diagnostics []Diagnostic
	applyR103Group("shared", bindings, &diagnostics)

	if len(diagnostics) != 3 {
		t.Fatalf("expected exactly 3 diagnostics, got %d", len(diagnostics))
	}

	// The sorted order should be:
	// 1. a.md#0 (make a1)
	// 2. a.md#1 (make a2)
	// 3. c.md#0 (make c)

	// For a.md#0 (make a1), the first differing peer in sorted order is a.md#1 (make a2)
	// For a.md#1 (make a2), the first differing peer in sorted order is a.md#0 (make a1)
	// For c.md#0 (make c), the first differing peer in sorted order is a.md#0 (make a1)

	// applyR103Group iterates over the sorted bindings, so the diagnostics will be in that same sorted order.
	expected := []Diagnostic{
		{
			File:    "a.md",
			Pointer: "/enforced_by/0/run",
			Code:    "R-103",
			Message: "ref shared disagrees on run (services/a.md#/enforced_by/1/run)",
		},
		{
			File:    "a.md",
			Pointer: "/enforced_by/1/run",
			Code:    "R-103",
			Message: "ref shared disagrees on run (services/a.md#/enforced_by/0/run)",
		},
		{
			File:    "c.md",
			Pointer: "/enforced_by/0/run",
			Code:    "R-103",
			Message: "ref shared disagrees on run (services/a.md#/enforced_by/0/run)",
		},
	}

	if !reflect.DeepEqual(diagnostics, expected) {
		t.Fatalf("diagnostics mismatch.\nwant: %+v\ngot:  %+v", expected, diagnostics)
	}
}
