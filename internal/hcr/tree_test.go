package hcr

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

type treeExpectations struct {
	Diagnostics []treeDiagnosticExpectation `yaml:"diagnostics"`
}

type treeDiagnosticExpectation struct {
	File string `yaml:"file"`
	At   string `yaml:"at"`
	Rule string `yaml:"rule"`
}

func TestTree(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join(corpusRoot, "fixtures", "*"))
	mustNoError(t, err)
	for _, fixture := range fixtures {
		info, statErr := os.Stat(fixture)
		if statErr != nil {
			t.Fatal(statErr)
			return
		}
		if !info.IsDir() {
			continue
		}
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			expectationPath := filepath.Join(fixture, "expect.yaml")
			source, readErr := os.ReadFile(expectationPath)
			if os.IsNotExist(readErr) {
				// Correct and temporary: Wave 3 installs the harness before Wave 4
				// creates expect.yaml fixtures. Vacuous success must not be guarded.
				t.Skip("fixture has no expect.yaml yet")
			}
			mustNoError(t, readErr)
			var expected treeExpectations
			mustNoError(t, yaml.Unmarshal(source, &expected))

			diagnostics, validationErr := ValidateTree(fixture)

			mustNoError(t, validationErr)
			actualKeys := make([]string, 0, len(diagnostics))
			absFixture, absErr := filepath.Abs(fixture)
			mustNoError(t, absErr)
			for _, diagnostic := range diagnostics {
				relative, relativeErr := filepath.Rel(absFixture, diagnostic.File)
				mustNoError(t, relativeErr)
				actualKeys = append(actualKeys, strings.Join([]string{filepath.ToSlash(relative), diagnostic.Pointer, diagnostic.Code}, "\x00"))
			}
			expectedKeys := make([]string, 0, len(expected.Diagnostics))
			for _, diagnostic := range expected.Diagnostics {
				expectedKeys = append(expectedKeys, strings.Join([]string{diagnostic.File, diagnostic.At, diagnostic.Rule}, "\x00"))
			}
			sort.Strings(actualKeys)
			sort.Strings(expectedKeys)
			mustDeepEqual(t, expectedKeys, actualKeys)
		})
	}
}

func TestDiscoverServices_finds_fixture_services_and_none_at_repo_root(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join(corpusRoot, "fixtures", "multi-service"))
	mustNoError(t, err)
	repoRoot, err := filepath.Abs(filepath.Join(corpusRoot, ".."))
	mustNoError(t, err)

	fixtureServices, err := discoverServices(fixtureRoot)
	mustNoError(t, err)
	repoServices, err := discoverServices(repoRoot)

	mustNoError(t, err)
	mustEqual(t, 2, len(fixtureServices))
	mustEqual(t, 0, len(repoServices))
}

func TestCanonicalServices_deduplicates_before_overlap_check(t *testing.T) {
	root := t.TempDir()
	service := filepath.Join(root, "services", "api")

	services, err := canonicalServices(root, []string{service, service})

	mustNoError(t, err)
	mustDeepEqual(t, []string{service}, services)
}

func TestCanonicalServices_reports_overlapping_matches(t *testing.T) {
	root := t.TempDir()
	service := filepath.Join(root, "services", "api")
	nested := filepath.Join(service, "nested")

	_, err := canonicalServices(root, []string{nested, service})

	mustEqual(t, "overlapping service_globs matches: "+service+", "+nested, err.Error())
}
