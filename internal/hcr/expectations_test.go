package hcr

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const expectationsFailure = "Fixture expectations changed. Per D3, a fixture's `valid`/`at`/`covers` may never be edited to make jigctl pass; a mismatch is a tool bug until proven otherwise. If this is genuinely a corpus correction, regenerate with `mise run expectations:freeze` and say why in the commit message."

func TestExpectationsFrozen(t *testing.T) {
	assertExpectationsFrozen(t)
}

func assertExpectationsFrozen(t *testing.T) {
	t.Helper()
	// This backstop surfaces expectation edits for review; it cannot prevent
	// someone from editing both a fixture and the frozen digest.
	digest, err := expectationDigest()
	mustNoError(t, err)
	path := filepath.Join("testdata", "expectations.sha256")
	if os.Getenv("JIGCTL_FREEZE_EXPECTATIONS") == "1" {
		mustNoError(t, os.WriteFile(path, []byte(digest+"\n"), 0o644))
		return
	}
	frozen, err := os.ReadFile(path)
	mustNoError(t, err)
	mustEqual(t, strings.TrimSpace(string(frozen)), digest, expectationsFailure)
}

func expectationDigest() (string, error) {
	parts := make([]string, 0, 64)
	records, err := filepath.Glob(filepath.Join(corpusRoot, "records", "*.md"))
	if err != nil {
		return "", fmt.Errorf("discover corpus records: %w", err)
	}
	for _, path := range records {
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", fmt.Errorf("read corpus record %s: %w", path, readErr)
		}
		for _, match := range expectationPattern.FindAllSubmatch(source, -1) {
			parts = append(parts, string(match[1]))
		}
	}
	err = filepath.WalkDir(filepath.Join(corpusRoot, "fixtures"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "expect.yaml" {
			return nil
		}
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read tree expectations %s: %w", path, readErr)
		}
		parts = append(parts, string(source))
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("discover tree expectations: %w", err)
	}
	sort.Strings(parts)
	digest := sha256.Sum256([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(digest[:]), nil
}
