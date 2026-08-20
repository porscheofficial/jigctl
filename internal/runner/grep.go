package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/patricebouillet/jigctl/internal/hcr"
)

// Grep evaluates a grep binding against literal substring requirements and prohibitions.
//
//nolint:gocritic // hugeParam: signature matches other evaluators
func Grep(plan hcr.Plan, target hcr.Target, binding hcr.ExecutableBinding) *Verdict {
	report := VerdictReport{
		Identity: BindingIdentity{
			RecordPath:   binding.RecordPath,
			BindingIndex: binding.BindingIndex,
		},
		Target: TargetProvenance{
			Name: target.Kind,
			Path: target.Path,
		},
		Kind:     binding.Kind,
		Severity: binding.Severity,
	}

	matches, err := MatchFiles(plan.Root, binding.File)
	if err != nil {
		if errors.Is(err, ErrNoMatches) {
			return NewBlockedVerdict(&report, ReasonGlobNoMatches)
		}
		return NewBlockedVerdict(&report, ReasonGlobInvalid)
	}

	findings, blockReason := evaluateGrepMatches(plan.Root, matches, binding)
	if blockReason != ReasonNone {
		return NewBlockedVerdict(&report, BlockedReason(blockReason))
	}

	report.Findings = findings
	return NewCompletedVerdict(&report)
}

//nolint:gocritic // hugeParam: keeping consistent with caller
func evaluateGrepMatches(root string, matches []string, binding hcr.ExecutableBinding) ([]Finding, Reason) {
	const maxFileSize = 1024 * 1024
	requireFound := make([]bool, len(binding.Require))
	var findings []Finding

	for _, match := range matches {
		absPath := filepath.Join(root, match)
		f, err := os.Open(absPath)
		if err != nil {
			return nil, Reason(ReasonInputUnreadable)
		}

		content, err := io.ReadAll(io.LimitReader(f, maxFileSize))
		f.Close()
		if err != nil {
			return nil, Reason(ReasonInputUnreadable)
		}

		findings = append(findings, checkForbids(match, content, binding.Forbid, binding.Severity)...)

		for i, require := range binding.Require {
			if !requireFound[i] && bytes.Contains(content, []byte(require)) {
				requireFound[i] = true
			}
		}
	}

	for i := range binding.Require {
		if !requireFound[i] {
			findings = append(findings, Finding{
				Locus: Locus{
					File: binding.File,
				},
				Severity: binding.Severity,
			})
		}
	}

	return findings, ReasonNone
}

func checkForbids(match string, content []byte, forbids []string, severity string) []Finding {
	var findings []Finding
	for _, forbid := range forbids {
		idx := bytes.Index(content, []byte(forbid))
		if idx != -1 {
			line := bytes.Count(content[:idx], []byte("\n")) + 1
			findings = append(findings, Finding{
				Locus: Locus{
					File:    match,
					Pointer: fmt.Sprintf("L%d", line),
				},
				Severity: severity,
			})
		}
	}
	return findings
}
