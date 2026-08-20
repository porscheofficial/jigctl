package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfine_resolves_only_paths_inside_canonical_tree_root(t *testing.T) {
	// Given
	root := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "leaf.txt")
	writePathFixture(t, outsideFile, "outside leaf")
	if err := os.Symlink(outsideFile, filepath.Join(root, "linked-leaf")); err != nil {
		t.Fatalf("create leaf symlink: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked-parent")); err != nil {
		t.Fatalf("create parent symlink: %v", err)
	}
	nestedParent := filepath.Join(root, "legitimate")
	if err := os.Mkdir(nestedParent, 0o750); err != nil {
		t.Fatalf("create legitimate parent: %v", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("canonicalize fixture root: %v", err)
	}

	tests := []struct {
		name     string
		authored string
		want     string
		wantErr  bool
	}{
		{name: "parent traversal escapes", authored: "../escape", wantErr: true},
		{name: "absolute path escapes", authored: outsideFile, wantErr: true},
		{name: "symlinked leaf escapes", authored: "linked-leaf", wantErr: true},
		{name: "symlinked parent directory escapes", authored: "linked-parent/not-yet-created.toml", wantErr: true},
		{name: "legitimate nested path resolves", authored: "legitimate/nested/not-yet-created.toml", want: filepath.Join(canonicalRoot, "legitimate", "nested", "not-yet-created.toml")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			got, confineErr := confine(root, test.authored)

			// Then
			t.Logf("authored=%q resolved=%q error=%v", test.authored, got, confineErr)
			if test.wantErr {
				if confineErr == nil {
					t.Fatalf("confine(%q) error = nil, want escape error", test.authored)
				}
				return
			}
			if confineErr != nil {
				t.Fatalf("confine(%q) error = %v", test.authored, confineErr)
			}
			if got != test.want {
				t.Fatalf("confine(%q) = %q, want %q", test.authored, got, test.want)
			}
		})
	}
}

func TestConfine_config_assert_escape_is_operational_without_reading_outside(t *testing.T) {
	// Given
	workspace := t.TempDir()
	root := filepath.Join(workspace, "tree")
	if err := os.Mkdir(root, 0o750); err != nil {
		t.Fatalf("create tree root: %v", err)
	}
	outsideContent := "RECOGNISABLE_OUTSIDE_SECRET_22"
	outsidePath := filepath.Join(workspace, "outside.toml")
	writePathFixture(t, outsidePath, outsideContent)
	record := "enforced_by:\n  - kind: config-assert\n    file: ../outside.toml\n"
	writePathFixture(t, filepath.Join(root, ".hcr", "HCR-9999-escape.md"), record)
	var output strings.Builder

	// When — this is the config-assert resolution boundary; reading is permitted only after confinement.
	resolved, err := confine(root, "../outside.toml")
	if err == nil {
		content, readErr := os.ReadFile(resolved)
		if readErr != nil {
			t.Fatalf("read confined config fixture: %v", readErr)
		}
		output.Write(content)
	}

	// Then
	if err == nil {
		t.Fatal("confine() error = nil, want operational path-escape failure")
	}
	reason := ReasonPathEscapesRoot
	if strings.Contains(output.String(), outsideContent) {
		t.Fatalf("output exposed outside content %q", outsideContent)
	}
	t.Logf("completion=operational reason=%v error=%v output=%q outside-content-present=%t", reason, err, output.String(), strings.Contains(output.String(), outsideContent))
}

func writePathFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("create fixture parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
