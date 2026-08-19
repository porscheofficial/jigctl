package hcr

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyR111(t *testing.T) {
	root := t.TempDir()
	writeArtifact := func(relative string) {
		t.Helper()
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create artifact directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("artifact\n"), 0o600); err != nil {
			t.Fatalf("write artifact: %v", err)
		}
	}
	writeArtifact("docs/adr/0001-existing.md")
	writeArtifact("evidence/NOTE-0002.txt")

	cases := []struct {
		name      string
		mapping   map[string]string
		rationale string
		wantCount int
	}{
		{
			name:      "unmapped prefix is skipped",
			mapping:   map[string]string{"ADR": "docs/adr/{rest}-*.md"},
			rationale: "RFC-0009",
			wantCount: 0,
		},
		{
			name:      "rest substitution resolves an existing target",
			mapping:   map[string]string{"ADR": "docs/adr/{rest}-*.md"},
			rationale: "ADR-0001",
			wantCount: 0,
		},
		{
			name:      "missing mapped target is reported",
			mapping:   map[string]string{"ADR": "docs/adr/{rest}-*.md"},
			rationale: "ADR-0009",
			wantCount: 1,
		},
		{
			name:      "id substitution resolves an existing target",
			mapping:   map[string]string{"NOTE": "evidence/{id}.txt"},
			rationale: "NOTE-0002",
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			emitters := []emitterRecord{{path: "record.md", meta: parsedMeta{Rationale: tc.rationale}}}
			var diagnostics []Diagnostic

			// When
			applyR111(root, tc.mapping, emitters, &diagnostics)

			// Then
			if len(diagnostics) != tc.wantCount {
				t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), tc.wantCount, diagnostics)
			}
			if tc.wantCount == 1 {
				diagnostic := diagnostics[0]
				if diagnostic.Code != "R-111" {
					t.Errorf("got Code %q, want R-111", diagnostic.Code)
				}
				if diagnostic.Pointer != "/rationale" {
					t.Errorf("got Pointer %q, want /rationale", diagnostic.Pointer)
				}
			}
		})
	}
}

func TestApplyR112(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		filename string
		want     bool
	}{
		{
			name:     "accepted - normal",
			id:       "HCR-0301",
			filename: "HCR-0301-x.md",
			want:     true,
		},
		{
			name:     "accepted - multi slug",
			id:       "HCR-0301",
			filename: "HCR-0301-a-b.md",
			want:     true,
		},
		{
			name:     "rejected - no slug",
			id:       "HCR-0301",
			filename: "HCR-0301.md",
			want:     false,
		},
		{
			name:     "rejected - prefix leak",
			id:       "HCR-0301",
			filename: "HCR-03011-x.md",
			want:     false,
		},
		{
			name:     "rejected - uppercase slug",
			id:       "HCR-0301",
			filename: "HCR-0301-API.md",
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			emitters := []emitterRecord{
				{
					path: tc.filename,
					meta: parsedMeta{ID: tc.id},
				},
			}
			var diags []Diagnostic
			applyR112(emitters, &diags)

			if tc.want {
				if len(diags) != 0 {
					t.Errorf("expected no diagnostics, got %v", diags)
				}
			} else {
				if len(diags) != 1 {
					t.Fatalf("expected 1 diagnostic, got %d", len(diags))
				}
				d := diags[0]
				if d.File != tc.filename {
					t.Errorf("got File %q, want %q", d.File, tc.filename)
				}
				if d.Pointer != "/id" {
					t.Errorf("got Pointer %q, want /id", d.Pointer)
				}
				if d.Code != "R-112" {
					t.Errorf("got Code %q, want R-112", d.Code)
				}
				wantMsg := fmt.Sprintf("filename %s does not match id %s", tc.filename, tc.id)
				if d.Message != wantMsg {
					t.Errorf("got Message %q, want %q", d.Message, wantMsg)
				}
			}
		})
	}
}

func TestJigctlSlug(t *testing.T) {
	cases := []struct {
		name    string
		heading string
		want    string
	}{
		{
			name:    "simple heading",
			heading: "# Policy notes",
			want:    "policy-notes",
		},
		{
			name:    "multiple spaces",
			heading: "### Data handling  (draft)",
			want:    "data-handling-draft",
		},
		{
			name:    "mixed case and punctuation",
			heading: "##    Mixed   CASE & Punctuation!   ",
			want:    "mixed-case--punctuation",
		},
		{
			name:    "documented API twice 1",
			heading: "## API",
			want:    "api",
		},
		{
			name:    "documented API twice 2",
			heading: "## API", // testing that identical headings produce the same slug
			want:    "api",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := jigctlSlug(tc.heading)
			if got != tc.want {
				t.Errorf("jigctlSlug(%q) = %q, want %q", tc.heading, got, tc.want)
			}
		})
	}
}
