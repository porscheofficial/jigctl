package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "jigctl")

	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "./")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build jigctl: %v\nOutput:\n%s", err, out)
	}
	return bin
}

func testNoJigToml(t *testing.T, bin string) {
	dir := t.TempDir()
	cmd := exec.CommandContext(context.Background(), bin, "run", dir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if cmd.ProcessState.ExitCode() != 2 {
		t.Errorf("Expected exit code 2, got %d", cmd.ProcessState.ExitCode())
	}
	if !strings.Contains(stderr.String(), "no jig.toml found") {
		t.Errorf("Expected 'no jig.toml found' in stderr, got: %s", stderr.String())
	}
}

func testSchemaInvalid(t *testing.T, bin string) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jig.toml"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".hcr"), 0o755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hcr", "invalid.md"), []byte("---\nid: invalid\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), bin, "run", dir)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got %d", cmd.ProcessState.ExitCode())
	}
	if !strings.Contains(stdout.String(), "schema: at '': missing properties") {
		t.Errorf("Expected schema error in stdout, got: %s", stdout.String())
	}
}

func testAuthEnforcement(t *testing.T, bin string) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jig.toml"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".hcr"), 0o755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	record := `---
id: HCR-0001
title: Test
scope: repo
regulates: reliability
summary: Test
state: enforced
enforced_by:
  - kind: command
    run: echo ok
---
`
	if err := os.WriteFile(filepath.Join(dir, ".hcr", "HCR-0001-test.md"), []byte(record), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Without --allow-exec
	cmd := exec.CommandContext(context.Background(), bin, "run", dir)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(stdout.String(), "needs --allow-exec") {
		t.Errorf("Expected the run to name the flag it needs, got: %s", stdout.String())
	}

	// With --allow-exec
	cmd2 := exec.CommandContext(context.Background(), bin, "run", dir, "--allow-exec")
	err2 := cmd2.Run()
	if err2 != nil {
		t.Fatalf("Expected success, got: %v", err2)
	}
}

func testSymlinkEscape(t *testing.T, bin string) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jig.toml"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".hcr"), 0o755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Create symlink pointing outside
	err := os.Symlink("/bin/echo", filepath.Join(dir, "escape"))
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	record := `---
id: HCR-0002
title: Test Symlink
scope: repo
regulates: reliability
summary: Test Symlink
state: enforced
enforced_by:
  - kind: command
    run: ./escape
---
`
	if writeErr := os.WriteFile(filepath.Join(dir, ".hcr", "HCR-0002-test-symlink.md"), []byte(record), 0o644); writeErr != nil {
		t.Fatalf("WriteFile failed: %v", writeErr)
	}

	cmd := exec.CommandContext(context.Background(), bin, "run", dir, "--allow-exec")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	out := stdout.String()
	if !strings.Contains(out, "path escapes the repository root") {
		t.Errorf("Expected the run to report the escaped path, got: %s", out)
	}
}

func TestRunCommand(t *testing.T) {
	bin := buildCLI(t)
	t.Run("No jig.toml", func(t *testing.T) { testNoJigToml(t, bin) })
	t.Run("Schema invalid record", func(t *testing.T) { testSchemaInvalid(t, bin) })
	t.Run("Authorization enforcement", func(t *testing.T) { testAuthEnforcement(t, bin) })
	t.Run("Symlink escape security checkpoint", func(t *testing.T) { testSymlinkEscape(t, bin) })
}
