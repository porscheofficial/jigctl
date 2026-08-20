package hcr

import (
	"testing"

	"github.com/goccy/go-yaml"
)

// TestExtractFrontmatter_pins_current_behavior characterizes extractFrontmatter's
// existing contract before it gains a body return value: a UTF-8 BOM is
// stripped, CRLF line endings are normalized to LF, and the bytes strictly
// between the opening and closing delimiters are returned as frontmatter.
// This test must keep passing once extractFrontmatter also returns the
// record body, proving the frontmatter half of the contract did not shift.
func TestExtractFrontmatter_pins_current_behavior(t *testing.T) {
	tests := []struct {
		name            string
		source          []byte
		wantFrontmatter string
		wantPresent     bool
	}{
		{
			name:            "plain LF source",
			source:          []byte("---\nid: HCR-0001\ntitle: Example\n---\nBody text.\n"),
			wantFrontmatter: "id: HCR-0001\ntitle: Example",
			wantPresent:     true,
		},
		{
			name:            "CRLF source normalizes to LF",
			source:          []byte("---\r\nid: HCR-0001\r\ntitle: Example\r\n---\r\nBody text.\r\n"),
			wantFrontmatter: "id: HCR-0001\ntitle: Example",
			wantPresent:     true,
		},
		{
			name:            "UTF-8 BOM is stripped before matching",
			source:          append([]byte{0xef, 0xbb, 0xbf}, []byte("---\nid: HCR-0001\n---\nBody.\n")...),
			wantFrontmatter: "id: HCR-0001",
			wantPresent:     true,
		},
		{
			name:        "missing opening delimiter yields absent",
			source:      []byte("id: HCR-0001\n---\n"),
			wantPresent: false,
		},
		{
			name:        "missing closing delimiter yields absent",
			source:      []byte("---\nid: HCR-0001\n"),
			wantPresent: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			frontmatter, _, present := extractFrontmatter(test.source)
			if present != test.wantPresent {
				t.Fatalf("present = %v, want %v", present, test.wantPresent)
			}
			if present && string(frontmatter) != test.wantFrontmatter {
				t.Fatalf("frontmatter = %q, want %q", frontmatter, test.wantFrontmatter)
			}
		})
	}
}

// TestExtractFrontmatter_returns_body_after_closing_delimiter covers Todo 1's
// core acceptance criteria: the bytes after the closing "\n---\n" delimiter
// come back as body, on the same normalized (BOM-stripped, CRLF-to-LF) source
// the frontmatter half already uses, and the closing-delimiter search runs
// exactly once so a body that itself contains that sequence is not truncated.
func TestExtractFrontmatter_returns_body_after_closing_delimiter(t *testing.T) {
	tests := []struct {
		name     string
		source   []byte
		wantBody string
	}{
		{
			name:     "body present",
			source:   []byte("---\nid: HCR-0001\n---\nGuidance prose.\n"),
			wantBody: "Guidance prose.\n",
		},
		{
			name:     "empty body yields empty string, not an error",
			source:   []byte("---\nid: HCR-0001\n---\n"),
			wantBody: "",
		},
		{
			name:     "CRLF body normalizes to LF, consistent with the frontmatter path",
			source:   []byte("---\r\nid: HCR-0001\r\n---\r\nLine one.\r\nLine two.\r\n"),
			wantBody: "Line one.\nLine two.\n",
		},
		{
			name:     "UTF-8 BOM source still yields a normalized body",
			source:   append([]byte{0xef, 0xbb, 0xbf}, []byte("---\nid: HCR-0001\n---\nBody.\n")...),
			wantBody: "Body.\n",
		},
		{
			name:     "body containing the closing delimiter sequence is not truncated",
			source:   []byte("---\nid: HCR-0001\n---\nBefore.\n---\nAfter.\n"),
			wantBody: "Before.\n---\nAfter.\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, body, present := extractFrontmatter(test.source)
			if !present {
				t.Fatal("expected frontmatter to be present")
			}
			if string(body) != test.wantBody {
				t.Fatalf("body = %q, want %q", body, test.wantBody)
			}
		})
	}
}

// TestExtractFrontmatter_title_and_body_round_trip is Todo 1's QA "happy" and
// "failure" scenarios in one test: a record's title (decoded from the
// returned frontmatter) and its body (the second return value) both round
// trip from a single extractFrontmatter call, and a title that differs from
// the summary proves the two fields were not conflated.
func TestExtractFrontmatter_title_and_body_round_trip(t *testing.T) {
	// Given
	source := []byte(`---
id: HCR-9001
title: Round-trip both halves of a record
summary: A different sentence than the title, so conflation would be visible.
state: enforced
---
Guidance prose that a human reads before writing code.
`)

	// When
	frontmatter, body, present := extractFrontmatter(source)
	if !present {
		t.Fatal("expected frontmatter to be present")
	}
	var meta parsedMeta
	mustNoError(t, yaml.Unmarshal(frontmatter, &meta))

	// Then
	if meta.Title != "Round-trip both halves of a record" {
		t.Fatalf("title = %q, want %q", meta.Title, "Round-trip both halves of a record")
	}
	wantBody := "Guidance prose that a human reads before writing code.\n"
	if string(body) != wantBody {
		t.Fatalf("body = %q, want %q", body, wantBody)
	}
	if meta.Title == meta.Summary {
		t.Fatal("title and summary must not be conflated, but they are equal")
	}
}
