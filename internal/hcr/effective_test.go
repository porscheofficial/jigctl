package hcr

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestEffectiveSet(t *testing.T) {
	root := filepath.Join("..", "..", "corpus", "fixtures", "multi-service")
	services, err := DiscoverTreeServices(root)
	if err != nil {
		t.Fatalf("DiscoverTreeServices failed: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	canonicalRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		t.Fatalf("canonicalize root %s failed: %v", root, err)
	}

	apiExpected := []string{"HCR-0301", "HCR-0302", "HCR-0303", "HCR-0304"}
	billingExpected := []string{"HCR-0301", "HCR-0302", "HCR-0305"}

	for _, s := range services {
		rel, relErr := filepath.Rel(canonicalRoot, s.path)
		if relErr != nil {
			t.Fatalf("relativize path %s failed: %v", s.path, relErr)
		}

		records := EffectiveSet(s)
		var got []string
		for _, r := range records {
			got = append(got, r.ID)
		}

		switch rel {
		case filepath.Join("services", "api"):
			if !reflect.DeepEqual(got, apiExpected) {
				t.Errorf("api service expected %v, got %v", apiExpected, got)
			}
		case filepath.Join("services", "billing"):
			if !reflect.DeepEqual(got, billingExpected) {
				t.Errorf("billing service expected %v, got %v", billingExpected, got)
			}
		default:
			t.Errorf("unexpected service %s", rel)
		}
	}
}

func TestEffectiveSet_SortsCorrectly(t *testing.T) {
	s := Service{
		path: "test/path",
		repoRecords: []Record{
			{ID: "HCR-0900", Path: "r2"},
			{ID: "HCR-0100", Path: "r1"},
		},
		serviceRecords: []Record{
			{ID: "HCR-0500", Path: "s2"},
			{ID: "HCR-0300", Path: "s1"},
		},
	}
	got := EffectiveSet(s)
	expected := []string{"HCR-0100", "HCR-0300", "HCR-0500", "HCR-0900"}

	if len(got) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(got))
	}
	for i := range got {
		if got[i].ID != expected[i] {
			t.Errorf("at index %d expected %s, got %s", i, expected[i], got[i].ID)
		}
	}
}
