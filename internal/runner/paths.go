package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// confine resolves authored paths from the single tree root fixed by ADR-0010.
func confine(base, authored string) (string, error) {
	canonicalBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("canonicalize tree root %s: %w", base, err)
	}
	canonicalBase, err = filepath.Abs(filepath.Clean(canonicalBase))
	if err != nil {
		return "", fmt.Errorf("make tree root absolute %s: %w", canonicalBase, err)
	}

	target := filepath.Join(canonicalBase, authored)
	if filepath.IsAbs(authored) {
		target = filepath.Clean(authored)
	}
	ancestor, remainder, err := deepestExistingAncestor(target)
	if err != nil {
		return "", err
	}
	canonicalAncestor, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", fmt.Errorf("canonicalize authored path ancestor %s: %w", ancestor, err)
	}
	resolved := filepath.Join(canonicalAncestor, remainder)
	relative, err := filepath.Rel(canonicalBase, resolved)
	if err != nil {
		return "", fmt.Errorf("relativize authored path %s: %w", authored, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes tree root: %s", authored)
	}
	return resolved, nil
}

// deepestExistingAncestor preserves a nonexistent suffix while exposing every
// existing parent to EvalSymlinks, including a symlinked parent directory.
func deepestExistingAncestor(target string) (ancestor, remainder string, err error) {
	ancestor = filepath.Clean(target)
	for {
		_, statErr := os.Lstat(ancestor)
		if statErr == nil {
			computedRemainder, relErr := filepath.Rel(ancestor, target)
			if relErr != nil {
				return "", "", fmt.Errorf("relativize nonexistent path suffix %s: %w", target, relErr)
			}
			return ancestor, computedRemainder, nil
		}
		if !os.IsNotExist(statErr) {
			return "", "", fmt.Errorf("inspect authored path ancestor %s: %w", ancestor, statErr)
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", "", fmt.Errorf("find existing ancestor of authored path %s: %w", target, statErr)
		}
		ancestor = parent
	}
}
