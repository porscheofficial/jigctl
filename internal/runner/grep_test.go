package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

func TestGrep(t *testing.T) {
	root := t.TempDir()

	// Write some test files
	file1 := filepath.Join(root, "main.go")
	if err := os.WriteFile(file1, []byte("package main\nimport \"os\"\nfunc main() {\n\tos.Exit(1)\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(root, "ci.yml")
	if err := os.WriteFile(file2, []byte("steps:\n  - run: mise run check\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plan := hcr.Plan{Root: root}
	target := hcr.Target{Kind: "repo"}

	t.Run("require passes", func(t *testing.T) {
		binding := hcr.ExecutableBinding{
			File:    "ci.yml",
			Require: []string{"mise run check"},
		}
		verdict := Grep(plan, target, binding)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completed, got %v with reason %v", verdict.Completion(), verdict.Reason())
		}
		if len(verdict.Report().Findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(verdict.Report().Findings))
		}
	})

	t.Run("require fails", func(t *testing.T) {
		binding := hcr.ExecutableBinding{
			File:    "ci.yml",
			Require: []string{"mise run test"},
		}
		verdict := Grep(plan, target, binding)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completed, got %v with reason %v", verdict.Completion(), verdict.Reason())
		}
		if len(verdict.Report().Findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(verdict.Report().Findings))
		}
		if verdict.Report().Findings[0].Locus.File != "ci.yml" {
			t.Errorf("expected finding file to be ci.yml, got %s", verdict.Report().Findings[0].Locus.File)
		}
	})

	t.Run("forbid literal string", func(t *testing.T) {
		binding := hcr.ExecutableBinding{
			File:   "*.go",
			Forbid: []string{"os.Exit("},
		}
		verdict := Grep(plan, target, binding)
		if verdict.Completion() != CompletionCompleted {
			t.Fatalf("expected completed, got %v with reason %v", verdict.Completion(), verdict.Reason())
		}
		if len(verdict.Report().Findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(verdict.Report().Findings))
		}
		if verdict.Report().Findings[0].Locus.Pointer != "L4" {
			t.Errorf("expected pointer to be L4, got %s", verdict.Report().Findings[0].Locus.Pointer)
		}
	})

	t.Run("no matches", func(t *testing.T) {
		binding := hcr.ExecutableBinding{
			File:    "*.ts",
			Require: []string{"import"},
		}
		verdict := Grep(plan, target, binding)
		if verdict.Completion() != CompletionBlocked {
			t.Fatalf("expected blocked, got %v with reason %v", verdict.Completion(), verdict.Reason())
		}
		if verdict.Reason() != Reason(ReasonGlobNoMatches) {
			t.Errorf("expected reason glob no matches, got %v", verdict.Reason())
		}
	})
}
