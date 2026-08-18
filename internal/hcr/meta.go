package hcr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type bindingRef struct {
	file         string
	relPath      string
	bindingIndex int
	run          string
	ref          string
}

// applyMetaRules delegates to individual rule implementations for META validations.
// Decision D10: identityIndex contains every discovered record whose id matches
// ^HCR-[0-9]{4}$, INCLUDING schema-failed ones. A record that failed the schema layer
// is NOT in emitters, so it neither emits R-101/R-102 diagnostics nor participates in
// an R-101 collision - but it IS in identityIndex, so it can still satisfy someone else's supersedes.
// This asymmetry is deliberate.
func applyMetaRules(
	canonicalRoot string,
	identityIndex map[string]bool,
	emitters []emitterRecord,
	diagnostics *[]Diagnostic,
) {
	applyR101(emitters, diagnostics)
	applyR102(identityIndex, emitters, diagnostics)
	applyR103AndR104(canonicalRoot, emitters, diagnostics)
	applyR109(emitters, diagnostics)
	applyR110(canonicalRoot, emitters, diagnostics)
	applyR112(emitters, diagnostics)
}

// applyR101 enforces id uniqueness across the whole tree.
func applyR101(emitters []emitterRecord, diagnostics *[]Diagnostic) {
	emitterIDs := make(map[string][]string)
	for _, e := range emitters {
		if e.meta.ID != "" {
			emitterIDs[e.meta.ID] = append(emitterIDs[e.meta.ID], e.relPath)
		}
	}

	for _, e := range emitters {
		if e.meta.ID != "" {
			others := make([]string, 0, len(emitterIDs[e.meta.ID])-1)
			for _, p := range emitterIDs[e.meta.ID] {
				if p != e.relPath {
					others = append(others, p)
				}
			}
			if len(others) > 0 {
				sort.Strings(others)
				msg := fmt.Sprintf("duplicate id %s (also in %s)", e.meta.ID, strings.Join(others, ", "))
				*diagnostics = append(*diagnostics, Diagnostic{
					File:    e.path,
					Pointer: "/id",
					Code:    "R-101",
					Message: msg,
				})
			}
		}
	}
}

// applyR102 ensures no dangling supersedes references exist.
// Scope decisions, all deliberate:
//   - It does NOT check the target's state. corpus/RULES.md says the target must exist.
//     Asserting the target is not deprecated would silently widen the rule.
//   - Cycles are NOT detected at M1. A supersedes B / B supersedes A is a legal input
//     that R-102 accepts, because every supersedes target exists.
//   - Consequence of D10: it checks identityIndex, not emitters, so a target that exists
//     but failed the schema layer still satisfies R-102.
func applyR102(identityIndex map[string]bool, emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for _, e := range emitters {
		if e.meta.Supersedes != "" {
			if e.meta.Supersedes == e.meta.ID {
				*diagnostics = append(*diagnostics, Diagnostic{
					File:    e.path,
					Pointer: "/supersedes",
					Code:    "R-102",
					Message: "record supersedes itself",
				})
			} else if !identityIndex[e.meta.Supersedes] {
				*diagnostics = append(*diagnostics, Diagnostic{
					File:    e.path,
					Pointer: "/supersedes",
					Code:    "R-102",
					Message: fmt.Sprintf("supersedes %s which does not exist", e.meta.Supersedes),
				})
			}
		}
	}
}

// applyR103AndR104 groups bindings for R-103 and delegates path resolution for R-104.
func applyR103AndR104(canonicalRoot string, emitters []emitterRecord, diagnostics *[]Diagnostic) {
	refs := make(map[string][]bindingRef)

	for _, e := range emitters {
		for i, binding := range e.meta.EnforcedBy {
			if refStr, ok := binding.Ref.(string); ok && refStr != "" {
				refs[refStr] = append(refs[refStr], bindingRef{
					file:         e.path,
					relPath:      e.relPath,
					bindingIndex: i,
					run:          binding.Run,
					ref:          refStr,
				})
			}
			if binding.Run != "" {
				applyR104(canonicalRoot, &e, i, binding, diagnostics)
			}
		}
	}

	for refName, bindings := range refs {
		applyR103Group(refName, bindings, diagnostics)
	}
}

// applyR104 ensures path-shaped run values resolve against the tree root.
// A run is path-shaped if its first whitespace-separated token contains a /.
// It resolves against canonicalRoot, not the record's directory. Existence only is
// checked (no executable bit), leading ./ is not special-cased, and symlinks are
// not followed specially (os.Stat semantics).
func applyR104(canonicalRoot string, e *emitterRecord, i int, binding parsedBinding, diagnostics *[]Diagnostic) {
	tokens := strings.Fields(binding.Run)
	if len(tokens) > 0 && strings.Contains(tokens[0], "/") {
		runPath := filepath.Join(canonicalRoot, tokens[0])
		if _, err := os.Stat(runPath); err != nil {
			if os.IsNotExist(err) {
				*diagnostics = append(*diagnostics, Diagnostic{
					File:    e.path,
					Pointer: fmt.Sprintf("/enforced_by/%d/run", i),
					Code:    "R-104",
					Message: fmt.Sprintf("run path %s does not exist", tokens[0]),
				})
			}
		}
	}
}

// applyR103Group enforces that all bindings sharing a ref agree on run.
// The group is sorted byte-wise ascending by (File, bindingIndex). Each diagnostic
// names the first binding in that sorted order whose run differs from its own run,
// ensuring deterministic output for three-or-more-way disagreements.
func applyR103Group(refName string, bindings []bindingRef, diagnostics *[]Diagnostic) {
	if len(bindings) <= 1 {
		return
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].file != bindings[j].file {
			return bindings[i].file < bindings[j].file
		}
		return bindings[i].bindingIndex < bindings[j].bindingIndex
	})
	for _, b := range bindings {
		var diffPeer *bindingRef
		for i := range bindings {
			if bindings[i].run != b.run {
				diffPeer = &bindings[i]
				break
			}
		}
		if diffPeer != nil {
			msg := fmt.Sprintf("ref %s disagrees on run (%s#/enforced_by/%d/run)",
				refName, diffPeer.relPath, diffPeer.bindingIndex)
			*diagnostics = append(*diagnostics, Diagnostic{
				File:    b.file,
				Pointer: fmt.Sprintf("/enforced_by/%d/run", b.bindingIndex),
				Code:    "R-103",
				Message: msg,
			})
		}
	}
}
