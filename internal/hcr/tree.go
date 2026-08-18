package hcr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type treeConfig struct {
	ServiceGlobs []string `toml:"service_globs"`
}

type indexedRecord struct {
	path    string
	service string
	source  []byte
}

// discoverServices returns canonical service directories. Nested services are
// rejected because overlapping service ownership is unsupported at M1.
func discoverServices(root string) ([]string, error) {
	canonicalRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("canonicalize tree root %s: %w", root, err)
	}
	var config treeConfig
	if _, decodeErr := toml.DecodeFile(filepath.Join(canonicalRoot, "jig.toml"), &config); decodeErr != nil {
		return nil, fmt.Errorf("read jig.toml: %w", decodeErr)
	}
	matched := make([]string, 0)
	for _, pattern := range config.ServiceGlobs {
		paths, globErr := filepath.Glob(filepath.Join(canonicalRoot, pattern))
		if globErr != nil {
			return nil, fmt.Errorf("expand service glob %q: %w", pattern, globErr)
		}
		for _, path := range paths {
			info, statErr := os.Stat(filepath.Join(path, ".hcr"))
			if statErr != nil {
				if os.IsNotExist(statErr) {
					continue
				}
				return nil, fmt.Errorf("inspect service %s: %w", path, statErr)
			}
			if info.IsDir() {
				matched = append(matched, path)
			}
		}
	}
	return canonicalServices(canonicalRoot, matched)
}

func canonicalServices(root string, matched []string) ([]string, error) {
	deduplicated := make(map[string]struct{}, len(matched))
	for _, path := range matched {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("canonicalize service %s: %w", path, err)
		}
		deduplicated[filepath.Clean(absolute)] = struct{}{}
	}
	services := make([]string, 0, len(deduplicated))
	for path := range deduplicated {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, fmt.Errorf("relativize service %s: %w", path, err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("service glob escapes tree root: %s", path)
		}
		services = append(services, path)
	}
	sort.Strings(services)
	for index, ancestor := range services {
		for _, descendant := range services[index+1:] {
			relative, err := filepath.Rel(ancestor, descendant)
			if err != nil {
				return nil, fmt.Errorf("compare service paths %s and %s: %w", ancestor, descendant, err)
			}
			if relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("overlapping service_globs matches: %s, %s", ancestor, descendant)
			}
		}
	}
	return services, nil
}

func indexTree(root string) ([]indexedRecord, error) {
	canonicalRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("canonicalize tree root %s: %w", root, err)
	}
	services, err := discoverServices(canonicalRoot)
	if err != nil {
		return nil, err
	}
	records := make([]indexedRecord, 0)
	locations := append([]string{canonicalRoot}, services...)
	for _, location := range locations {
		paths, globErr := filepath.Glob(filepath.Join(location, ".hcr", "*.md"))
		if globErr != nil {
			return nil, fmt.Errorf("index records under %s: %w", location, globErr)
		}
		service := ""
		if location != canonicalRoot {
			service = location
		}
		for _, path := range paths {
			source, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil, fmt.Errorf("read record %s: %w", path, readErr)
			}
			records = append(records, indexedRecord{path: path, service: service, source: source})
		}
	}
	return records, nil
}

// ValidateTree validates each directly indexed record once. Wave 4 extends
// this indexed layer with META rules while preserving this operational contract.
func ValidateTree(root string) ([]Diagnostic, error) {
	records, err := indexTree(root)
	if err != nil {
		return nil, err
	}
	diagnostics := make([]Diagnostic, 0)
	for _, record := range records {
		recordDiagnostics, validationErr := schemaDiagnostics(record.path, record.source)
		if validationErr != nil {
			return nil, validationErr
		}
		diagnostics = append(diagnostics, recordDiagnostics...)
	}
	sortDiagnostics(diagnostics)
	return diagnostics, nil
}
