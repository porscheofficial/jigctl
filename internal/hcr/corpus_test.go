package hcr

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

const corpusRoot = "../../corpus"

var expectationPattern = regexp.MustCompile(`(?s)<!--\s*jig:expect\n(.*?)\n-->`)

type recordExpectation struct {
	Valid       *bool                   `yaml:"valid"`
	Covers      []string                `yaml:"covers"`
	Diagnostics []diagnosticExpectation `yaml:"diagnostics"`
	Deferred    []deferredExpectation   `yaml:"deferred"`
}

type diagnosticExpectation struct {
	Rule string `yaml:"rule"`
	At   string `yaml:"at"`
}

type deferredExpectation struct {
	Rule string `yaml:"rule"`
}

type corpusCounts struct {
	validated int
	reported  int
	missing   int
	noisy     int
	errored   int
	deferred  int
}

func (c corpusCounts) String() string {
	return fmt.Sprintf("validated=%d reported=%d missing=%d noisy=%d error=%d deferred=%d", c.validated, c.reported, c.missing, c.noisy, c.errored, c.deferred)
}

func TestCorpus(t *testing.T) {
	// Given
	paths, err := filepath.Glob(filepath.Join(corpusRoot, "records", "*.md"))
	mustNoError(t, err)
	if len(paths) == 0 {
		t.Fatal("no corpus records found")
	}
	rulesSource, err := os.ReadFile(filepath.Join(corpusRoot, "RULES.md"))
	mustNoError(t, err)
	rulePattern := regexp.MustCompile(`\|\s*(R-\d+)\s*\|`)
	rules := make(map[string]struct{})
	for _, match := range rulePattern.FindAllStringSubmatch(string(rulesSource), -1) {
		rules[match[1]] = struct{}{}
	}

	// When
	counts := corpusCounts{}
	cited := make(map[string]struct{})
	for _, path := range paths {
		evaluateCorpusRecord(t, path, &counts, cited)
	}
	t.Log(counts.String())
	assertExpectationsFrozen(t)

	// Then
	for rule := range cited {
		_, ok := rules[rule]
		if !ok {
			t.Fatalf("cited rule absent from RULES.md: %s", rule)
		}
	}
	mustEqual(t, 0, counts.errored)
	mustEqual(t, 0, counts.missing)
	mustEqual(t, 0, counts.noisy)
	mustEqual(t, len(paths), counts.validated+counts.reported+counts.missing+counts.noisy+counts.errored)
}

func evaluateCorpusRecord(t *testing.T, path string, counts *corpusCounts, cited map[string]struct{}) {
	t.Helper()
	source, readErr := os.ReadFile(path)
	if readErr != nil {
		counts.errored++
		return
	}
	expectation, validShape := parseRecordExpectation(source)
	if !validShape {
		counts.errored++
		return
	}
	for _, rule := range expectation.Covers {
		cited[rule] = struct{}{}
	}
	for _, item := range expectation.Deferred {
		cited[item.Rule] = struct{}{}
	}
	counts.deferred += len(expectation.Deferred)

	diagnostics, validationErr := ValidateRecord(path, source)
	if validationErr != nil {
		counts.errored++
		return
	}
	classifyCorpusResult(t, path, &expectation, diagnostics, counts)
}

func parseRecordExpectation(source []byte) (recordExpectation, bool) {
	blocks := expectationPattern.FindAllSubmatch(source, -1)
	if len(blocks) != 1 {
		return recordExpectation{}, false
	}
	var raw map[string]any
	if decodeErr := yaml.Unmarshal(blocks[0][1], &raw); decodeErr != nil {
		return recordExpectation{}, false
	}
	allowed := map[string]struct{}{"valid": {}, "covers": {}, "diagnostics": {}, "deferred": {}}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return recordExpectation{}, false
		}
	}
	var expectation recordExpectation
	if decodeErr := yaml.Unmarshal(blocks[0][1], &expectation); decodeErr != nil || expectation.Valid == nil {
		return recordExpectation{}, false
	}
	if (*expectation.Valid && raw["diagnostics"] != nil) || (!*expectation.Valid && raw["diagnostics"] == nil) {
		return recordExpectation{}, false
	}
	if !*expectation.Valid && (len(expectation.Covers) != 1 || len(expectation.Diagnostics) != 1 || expectation.Diagnostics[0].Rule != expectation.Covers[0]) {
		return recordExpectation{}, false
	}
	return expectation, true
}

func classifyCorpusResult(t *testing.T, path string, expectation *recordExpectation, diagnostics []Diagnostic, counts *corpusCounts) {
	t.Helper()
	if *expectation.Valid {
		if len(diagnostics) == 0 {
			counts.validated++
		} else {
			counts.noisy++
		}
		return
	}
	if len(diagnostics) == 0 {
		counts.missing++
		return
	}
	if len(diagnostics) == 1 && atOK(diagnostics[0].Pointer, expectation.Diagnostics[0].At) {
		mustEqual(t, path, diagnostics[0].File)
		mustEqual(t, "schema", diagnostics[0].Code)
		counts.reported++
		return
	}
	counts.noisy++
}

// 2020-12 does not specify where a "property must not be here" violation is
// reported: check-jsonschema says the containing object, Go santhosh-tekuri and
// Rust boon/jsonschema say the offending property. Both conform, so accept the
// asserted pointer or one segment below. Do NOT tighten to == (breaks 9 fixtures).
func atOK(got, want string) bool {
	if got == want {
		return true
	}
	if !strings.HasPrefix(got, want+"/") {
		return false
	}
	return !strings.Contains(got[len(want)+1:], "/")
}

func TestIssue156_pruned_walk_reports_one_diagnostic_at_asserted_pointer(t *testing.T) {
	// santhosh-tekuri/jsonschema issue #156 is closed as invalid: broad
	// unevaluatedProperties errors conform to 2020-12, so pruning is our guard.
	paths := []string{"r003-unknown-key.md", "r011-unknown-binding-key.md"}
	wants := []string{"", "/enforced_by/0"}
	for index, name := range paths {
		path := filepath.Join(corpusRoot, "records", name)
		source, err := os.ReadFile(path)
		mustNoError(t, err)

		diagnostics, err := ValidateRecord(path, source)

		mustNoError(t, err)
		if len(diagnostics) != 1 {
			t.Fatalf("want one diagnostic, got %d", len(diagnostics))
			return
		}
		mustEqual(t, path, diagnostics[0].File)
		if !atOK(diagnostics[0].Pointer, wants[index]) {
			t.Fatalf("want asserted pointer %q (or conforming child), got %q", wants[index], diagnostics[0].Pointer)
		}
		mustEqual(t, "schema", diagnostics[0].Code)
	}
}
