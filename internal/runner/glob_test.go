package runner

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestMatchFiles_matches_go_files_at_every_depth(t *testing.T) {
	// Given
	root := t.TempDir()
	want := []string{"one.go", "a/two.go", "a/b/three.go"}
	writeFixtureFiles(t, root, want)
	writeFixtureFiles(t, root, []string{"a/not-go.txt"})

	// When
	got, err := MatchFiles(root, "**/*.go")

	// Then
	if len(got) != len(want) {
		t.Fatalf("MatchFiles() count = %d, want %d: %v", len(got), len(want), got)
	}
	assertMatches(t, got, err, want)
}

func TestMatchFiles_matches_exactly_every_git_visible_go_file_in_repository(t *testing.T) {
	// Given
	root := filepath.Clean(filepath.Join(mustWorkingDirectory(t), "..", ".."))
	command := exec.CommandContext(
		t.Context(),
		"git", "ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", "*.go",
	)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list git-visible Go files: %v", err)
	}
	want := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	for index := range want {
		want[index] = filepath.ToSlash(filepath.Clean(want[index]))
	}
	slices.Sort(want)

	// When
	got, err := MatchFiles(root, "**/*.go")
	// Then
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	for index := range got {
		got[index] = filepath.ToSlash(filepath.Clean(got[index]))
	}
	slices.Sort(got)
	missing := make([]string, 0)
	for _, path := range want {
		if _, found := slices.BinarySearch(got, path); !found {
			missing = append(missing, path)
		}
	}
	spurious := make([]string, 0)
	for _, path := range got {
		if _, found := slices.BinarySearch(want, path); !found {
			spurious = append(spurious, path)
		}
	}
	if len(missing) > 0 || len(spurious) > 0 || len(got) != len(want) {
		t.Fatalf(
			"MatchFiles() count = %d, git ls-files count = %d; missing = %v; spurious = %v",
			len(got), len(want), missing, spurious,
		)
	}
	t.Logf("MatchFiles() count = %d; git ls-files count = %d", len(got), len(want))
}

func TestMatchFiles_roots_pattern_at_tree_root(t *testing.T) {
	// Given
	root := t.TempDir()
	writeFixtureFiles(t, root, []string{"src/a/b/c.py", "other/a.py"})

	// When
	got, err := MatchFiles(root, "src/**/*.py")

	// Then
	assertMatches(t, got, err, []string{"src/a/b/c.py"})
}

func TestMatchFiles_returns_named_signal_when_nothing_matches(t *testing.T) {
	// Given
	root := t.TempDir()

	// When
	got, err := MatchFiles(root, "**/*.go")

	// Then
	if !errors.Is(err, ErrNoMatches) {
		t.Fatalf("MatchFiles() error = %v, want ErrNoMatches", err)
	}
	if got != nil {
		t.Fatalf("MatchFiles() = %v, want nil", got)
	}
}

func TestMatchFiles_does_not_descend_into_git(t *testing.T) {
	// Given
	root := t.TempDir()
	writeFixtureFiles(t, root, []string{"kept.go", ".git/objects/x.go"})

	// When
	got, err := MatchFiles(root, "**/*.go")

	// Then
	assertMatches(t, got, err, []string{"kept.go"})
}

func TestMatchFiles_does_not_follow_directory_symlinks(t *testing.T) {
	// Given
	root := t.TempDir()
	outside := t.TempDir()
	writeFixtureFiles(t, root, []string{"kept.go"})
	writeFixtureFiles(t, outside, []string{"hidden.go"})
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatalf("create fixture symlink: %v", err)
	}

	// When
	got, err := MatchFiles(root, "**/*.go")

	// Then
	assertMatches(t, got, err, []string{"kept.go"})
}

func writeFixtureFiles(t *testing.T, root string, paths []string) {
	t.Helper()
	for _, path := range paths {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
			t.Fatalf("create fixture directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte("fixture\n"), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}
}

func assertMatches(t *testing.T, got []string, err error, want []string) {
	t.Helper()
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("MatchFiles() = %v, want %v", got, want)
	}
}

func mustWorkingDirectory(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return directory
}
