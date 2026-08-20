package hcr

import (
	"fmt"
	"sort"

	"github.com/goccy/go-yaml"
)

// Service represents a discovered service directory. Its fields are unexported
// so no external caller can populate it with meaningful contents; the only way
// to obtain a populated Service is through the package's discovery mechanism.
// The zero value is constructible externally and safely yields an empty
// effective set. The signature of EffectiveSet guarantees safety: by taking a
// Service rather than a caller-supplied string, there is no "unknown service"
// code path to fail at runtime.
type Service struct {
	path           string
	repoRecords    []Record
	serviceRecords []Record
}

// Record represents a discovered record in the tree, identified by id and path,
// not necessarily schema-valid. Membership resolution is deliberately independent
// of schema validity so that a service's effective set does not silently shrink
// when one of its records is malformed.
type Record struct {
	ID   string
	Path string
}

// DiscoverTreeServices returns the services discovered in the tree, along with
// the records necessary to compute their effective sets.
func DiscoverTreeServices(root string) ([]Service, error) {
	services, err := discoverServices(root)
	if err != nil {
		return nil, err
	}

	indexed, err := indexTree(root)
	if err != nil {
		return nil, err
	}

	var repo []Record
	serviceMap := make(map[string][]Record)

	for _, idxRec := range indexed {
		var id string
		if fm, _, present := extractFrontmatter(idxRec.source); present {
			var meta struct {
				ID string `yaml:"id"`
			}
			if unmarshalErr := yaml.Unmarshal(fm, &meta); unmarshalErr == nil {
				id = meta.ID
			}
		}

		rec := Record{
			ID:   id,
			Path: idxRec.path,
		}

		if idxRec.service == "" {
			repo = append(repo, rec)
		} else {
			serviceMap[idxRec.service] = append(serviceMap[idxRec.service], rec)
		}
	}

	var out []Service
	for _, sp := range services {
		out = append(out, Service{
			path:           sp,
			repoRecords:    repo,
			serviceRecords: serviceMap[sp],
		})
	}
	return out, nil
}

// EffectiveSet returns the effective set of records for a given service.
//
// R-101, R-102 and R-103 are tree-global (decision D14) and MUST NOT be
// evaluated per effective set — doing so would re-report every repo-record
// finding once per service.
// R-109 is a purely local comparison.
// EffectiveSet exists at M1 because it is the substrate M2's binding execution
// needs, and because RULES.md already declares R-108. Its correctness is asserted
// directly by a unit test rather than inferred from a rule that happens to use it.
func EffectiveSet(s Service) []Record {
	out := make([]Record, 0, len(s.repoRecords)+len(s.serviceRecords))
	out = append(out, s.repoRecords...)
	out = append(out, s.serviceRecords...)

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out
}

// applyR109 ensures a record's scope agrees with its directory location.
// A record directly inside <root>/.hcr/ MUST declare scope: repo.
// A record directly inside <serviceDir>/.hcr/ MUST declare scope: service.
func applyR109(emitters []emitterRecord, diagnostics *[]Diagnostic) {
	for i := range emitters {
		e := &emitters[i]
		expected := "repo"
		if e.service != "" {
			expected = "service"
		}

		if e.meta.Scope != expected {
			*diagnostics = append(*diagnostics, Diagnostic{
				File:    e.path,
				Pointer: "/scope",
				Code:    "R-109",
				Message: fmt.Sprintf("scope %s does not match location (expected %s)", e.meta.Scope, expected),
			})
		}
	}
}
