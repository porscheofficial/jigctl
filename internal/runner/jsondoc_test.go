package runner

import (
	"os"
	"strings"
	"testing"
)

func TestJSONDoc_ExhaustiveProjectionCode(t *testing.T) {
	if len(jsonProjectionCode) != 6 {
		t.Errorf("expected jsonProjectionCode to have exactly 6 entries, got %d", len(jsonProjectionCode))
	}

	for k, v := range jsonProjectionCode {
		if v == "" {
			t.Errorf("jsonProjectionCode entry for %v is empty", k)
		}
	}
}

func TestJSONDoc_NoOmitEmpty(t *testing.T) {
	content, err := os.ReadFile("jsondoc.go")
	if err != nil {
		t.Fatalf("failed to read jsondoc.go: %v", err)
	}

	if strings.Contains(string(content), "omitempty") {
		t.Errorf("jsondoc.go must not contain 'omitempty' in any JSON struct tag")
	}
}
