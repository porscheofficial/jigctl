package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func setupValidFixture(t *testing.T) string {
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
	return dir
}

// unsetEnv removes key from the process environment for the duration of the
// test and restores it automatically at test end -- to whatever it was
// before (present with a value, or genuinely absent), never to "".
//
// testing.T ships Setenv but no Unsetenv, and t.Setenv(key, "") is NOT a
// substitute for unsetting here: this codebase gives "present but empty"
// and "absent" different meanings. shouldEnableColor (color.go) disables
// colour whenever JIGCTL_NO_COLOR is present at all -- even set to "", see
// color_test.go's "disabled by JIGCTL_NO_COLOR (set but empty)" case --
// while a present-but-empty NO_COLOR does NOT disable colour, see
// color_test.go's "enabled if NO_COLOR is present but empty" case. Using
// t.Setenv(key, "") would silently turn "absent" into "present but empty"
// and flip JIGCTL_NO_COLOR's outcome, breaking TestRunWiring_ColorEnabled
// and the colour-enabled half of TestRunWiring_OutputDifference. Only a
// real absence behaves correctly for both variables, so we save/restore
// the true prior state by hand via t.Cleanup instead.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	orig, had := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

// clearColorEnv puts NO_COLOR, JIGCTL_NO_COLOR and TERM into the state every
// TestRunWiring_* test relies on (nothing forcing colour off, TERM not
// "dumb"), restoring the previous process environment automatically at
// test end.
func clearColorEnv(t *testing.T) {
	t.Helper()
	unsetEnv(t, "NO_COLOR")
	unsetEnv(t, "JIGCTL_NO_COLOR")
	t.Setenv("TERM", "xterm")
}

func TestRunWiring_NoColor(t *testing.T) {
	dir := setupValidFixture(t)
	var out bytes.Buffer

	// Save and restore globals
	origNoColor := runNoColor
	origPlain := runPlain
	origAllowExec := runAllowExec
	origExitCode := recordedExitCode
	defer func() {
		runNoColor = origNoColor
		runPlain = origPlain
		runAllowExec = origAllowExec
		recordedExitCode = origExitCode
	}()

	// runPlain must be explicit: if inherited true, RenderPlain would run
	// instead and this test would never exercise --no-color at all.
	runNoColor = true
	runPlain = false
	runAllowExec = true // we need to allow exec to get passing output

	clearColorEnv(t)

	err := runAction([]string{dir}, tty{IsTerminal: true}, &out) // pass isTerminal=true but no-color=true
	if err != nil {
		t.Fatalf("runAction failed: %v", err)
	}

	output := out.Bytes()
	if bytes.Contains(output, []byte("\x1b")) {
		t.Errorf("Expected 0 escape bytes with --no-color, got output:\n%s", output)
	}

	// Positive assertion: the run must actually have produced the report
	// (not merely avoided escape bytes by producing nothing at all).
	if !bytes.Contains(output, []byte("HCR-0001")) {
		t.Errorf("Expected output to contain the report for HCR-0001, got:\n%s", output)
	}
}

func TestRunWiring_Plain(t *testing.T) {
	dir := setupValidFixture(t)
	var out bytes.Buffer

	origPlain := runPlain
	origAllowExec := runAllowExec
	origNoColor := runNoColor
	origExitCode := recordedExitCode
	defer func() {
		runPlain = origPlain
		runAllowExec = origAllowExec
		runNoColor = origNoColor
		recordedExitCode = origExitCode
	}()

	runPlain = true
	runAllowExec = true
	runNoColor = false

	clearColorEnv(t)

	err := runAction([]string{dir}, tty{IsTerminal: true}, &out)
	if err != nil {
		t.Fatalf("runAction failed: %v", err)
	}

	output := out.Bytes()
	if bytes.Contains(output, []byte("\x1b")) {
		t.Errorf("Expected 0 escape bytes with --plain, got output:\n%s", output)
	}

	// Plain output check
	outStr := string(output)
	if !bytes.Contains(output, []byte("HCR-0001")) {
		t.Errorf("Expected plain output to contain HCR-0001, got:\n%s", outStr)
	}
	if !bytes.Contains(output, []byte("expected-unchecked")) {
		t.Errorf("Expected plain output to contain 'expected-unchecked', got:\n%s", outStr)
	}
}

func TestRunWiring_ColorEnabled(t *testing.T) {
	dir := setupValidFixture(t)
	var out bytes.Buffer

	// Ensure flags that disable color are off
	origNoColor := runNoColor
	origPlain := runPlain
	origAllowExec := runAllowExec
	origExitCode := recordedExitCode
	defer func() {
		runNoColor = origNoColor
		runPlain = origPlain
		runAllowExec = origAllowExec
		recordedExitCode = origExitCode
	}()
	runNoColor = false
	runPlain = false
	runAllowExec = true

	// Also clear env vars that might disable color for this process
	clearColorEnv(t)

	err := runAction([]string{dir}, tty{IsTerminal: true}, &out)
	if err != nil {
		t.Fatalf("runAction failed: %v", err)
	}

	output := out.Bytes()
	if !bytes.Contains(output, []byte("\x1b")) {
		t.Errorf("Expected escape bytes with color enabled, got plain output:\n%s", output)
	}
}

func TestRunWiring_OutputDifference(t *testing.T) {
	dir := setupValidFixture(t)
	var outColor bytes.Buffer
	var outPlain bytes.Buffer

	origNoColor := runNoColor
	origPlain := runPlain
	origAllowExec := runAllowExec
	origExitCode := recordedExitCode
	defer func() {
		runNoColor = origNoColor
		runPlain = origPlain
		runAllowExec = origAllowExec
		recordedExitCode = origExitCode
	}()

	runAllowExec = true
	clearColorEnv(t)

	// Color enabled run
	runNoColor = false
	runPlain = false
	if err := runAction([]string{dir}, tty{IsTerminal: true}, &outColor); err != nil {
		t.Fatalf("runAction failed: %v", err)
	}

	// Plain run
	runNoColor = false
	runPlain = true
	if err := runAction([]string{dir}, tty{IsTerminal: true}, &outPlain); err != nil {
		t.Fatalf("runAction failed: %v", err)
	}

	// Both runs must actually have produced the report -- otherwise the
	// inequality check below could pass vacuously (e.g. one side empty)
	// instead of proving a genuine difference in rendering shape.
	if !bytes.Contains(outColor.Bytes(), []byte("HCR-0001")) {
		t.Errorf("Expected color-enabled output to contain HCR-0001, got:\n%s", outColor.String())
	}
	if !bytes.Contains(outPlain.Bytes(), []byte("HCR-0001")) {
		t.Errorf("Expected plain output to contain HCR-0001, got:\n%s", outPlain.String())
	}

	// If plain and default output are identical, renderer selection failed
	if bytes.Equal(outColor.Bytes(), outPlain.Bytes()) {
		t.Errorf("Expected --plain and default mode to produce different output shapes, but they were identical")
	}
}
