package hcr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
	nonSlugRegex    = regexp.MustCompile(`[^a-z0-9-]+`)
	atxHeadingRegex = regexp.MustCompile(`^#{1,6}\s+`)
)

// utcDateOf discards the time of day, so that a waiver dated today stays in
// force for all of today whatever the hour and whatever the caller's timezone.
func utcDateOf(instant time.Time) time.Time {
	utc := instant.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func applyR107(currentDate time.Time, emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for emitterIndex := range emitters {
		emitter := &emitters[emitterIndex]
		for exceptionIndex := range emitter.meta.Exceptions {
			until := emitter.meta.Exceptions[exceptionIndex].Until
			if until == "" {
				continue
			}
			pointer := fmt.Sprintf("/exceptions/%d/until", exceptionIndex)
			expiry, err := time.Parse(dateLayout, until)
			if err != nil {
				*diagnostics = append(*diagnostics, Diagnostic{
					File: emitter.path, Pointer: pointer, Code: "R-107",
					Message: fmt.Sprintf("exception until %s is not a calendar date", until),
				})
				continue
			}
			if expiry.Before(currentDate) {
				*diagnostics = append(*diagnostics, Diagnostic{
					File: emitter.path, Pointer: pointer, Code: "R-107",
					Message: fmt.Sprintf("exception expired on %s", until),
				})
			}
		}
	}
}

// jigctlSlug computes the slug for an ATX markdown heading according to the
// exact jigctl rules.
// 1. Strip the leading '#'s and surrounding whitespace.
// 2. Lowercase.
// 3. Replace each run of whitespace with a single '-'.
// 4. Delete every character outside [a-z0-9-].
//
// Limitations (deliberate):
//   - This is NOT GitHub-style. A file with two `## API` headings has anchors `#api` and `#api-1`
//     under GitHub's rules. The jigctl slug maps both headings to `#api`, so it accepts `#api` but
//     rejects `#api-1`, which GitHub considers valid.
//   - It does not strip markdown emphasis or inline code backticks from heading text.
func jigctlSlug(heading string) string {
	s := strings.TrimLeft(heading, "#")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = whitespaceRegex.ReplaceAllString(s, "-")
	s = nonSlugRegex.ReplaceAllString(s, "")
	return s
}

// applyR110 ensures docs anchors resolve without network I/O.
func applyR110(canonicalRoot string, emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for emitterIndex := range emitters {
		e := &emitters[emitterIndex]
		for i := range e.meta.EnforcedBy {
			binding := &e.meta.EnforcedBy[i]
			if binding.Kind != "external" || binding.Docs == "" {
				continue
			}

			parts := strings.SplitN(binding.Docs, "#", 2)
			if len(parts) == 0 {
				continue
			}
			pathPart := parts[0]

			if strings.Contains(pathPart, "://") {
				continue
			}

			resolvedPath := filepath.Join(canonicalRoot, pathPart)
			if _, err := os.Stat(resolvedPath); err != nil {
				if os.IsNotExist(err) {
					*diagnostics = append(*diagnostics, Diagnostic{
						File:    e.path,
						Pointer: fmt.Sprintf("/enforced_by/%d/docs", i),
						Code:    "R-110",
						Message: fmt.Sprintf("docs path %s does not exist", pathPart),
					})
				}
				continue
			}

			if len(parts) > 1 {
				anchor := parts[1]
				if !hasAnchor(resolvedPath, anchor) {
					*diagnostics = append(*diagnostics, Diagnostic{
						File:    e.path,
						Pointer: fmt.Sprintf("/enforced_by/%d/docs", i),
						Code:    "R-110",
						Message: fmt.Sprintf("docs anchor #%s not found in %s", anchor, pathPart),
					})
				}
			}
		}
	}
}

func hasAnchor(path, anchor string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if atxHeadingRegex.MatchString(line) {
			if jigctlSlug(line) == anchor {
				return true
			}
		}
	}
	return false
}

// applyR111 ensures mapped rationale references resolve against the tree root.
func applyR111(canonicalRoot string, mapping map[string]string, emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for i := range emitters {
		e := &emitters[i]
		id := e.meta.Rationale
		if id == "" {
			continue
		}
		parts := strings.SplitN(id, "-", 2)
		if len(parts) < 2 {
			continue
		}
		pattern, mapped := mapping[parts[0]]
		if !mapped {
			continue
		}
		pattern = strings.ReplaceAll(pattern, "{id}", id)
		pattern = strings.ReplaceAll(pattern, "{rest}", parts[1])
		matches, globErr := filepath.Glob(filepath.Join(canonicalRoot, pattern))
		if globErr != nil || len(matches) == 0 {
			*diagnostics = append(*diagnostics, Diagnostic{
				File:    e.path,
				Pointer: "/rationale",
				Code:    "R-111",
				Message: fmt.Sprintf("rationale target %s does not exist", pattern),
			})
		}
	}
}

// applyR112 ensures the filename begins with its ID.
func applyR112(emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for i := range emitters {
		e := &emitters[i]
		if e.meta.ID == "" {
			continue
		}

		base := filepath.Base(e.path)
		pattern := "^" + regexp.QuoteMeta(e.meta.ID) + `-[a-z0-9]+(-[a-z0-9]+)*\.md$`

		matched, err := regexp.MatchString(pattern, base)
		if err == nil && !matched {
			*diagnostics = append(*diagnostics, Diagnostic{
				File:    e.path,
				Pointer: "/id",
				Code:    "R-112",
				Message: fmt.Sprintf("filename %s does not match id %s", base, e.meta.ID),
			})
		}
	}
}
