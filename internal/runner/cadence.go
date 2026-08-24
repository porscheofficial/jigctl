package runner

import (
	"errors"
	"fmt"
	"strings"

	"github.com/porscheofficial/jigctl/internal/hcr"
)

// CadenceSet is the set of cadences one invocation selects for.
// explicit records whether --cadence was supplied, which is what
// distinguishes an invoker's deselection from an author's exclusion.
type CadenceSet struct {
	values   map[string]struct{}
	explicit bool
}

// DefaultCadenceSet reproduces the pre-flag behaviour: {on-change, ci},
// explicit == false.
func DefaultCadenceSet() CadenceSet {
	return CadenceSet{
		values: map[string]struct{}{
			"on-change": {},
			"ci":        {},
		},
		explicit: false,
	}
}

// ParseCadenceSet builds the set for one invocation. supplied is the
// caller's Flags().Changed("cadence"), NOT a test on raw being empty:
// `--cadence=` is supplied-but-empty and must be an error, not a
// silent fall back to the default.
//
//	supplied == false            -> DefaultCadenceSet(), raw ignored
//	supplied == true, raw == ""  -> error
//	supplied == true             -> parsed set, explicit == true
func ParseCadenceSet(raw string, supplied bool) (CadenceSet, error) {
	if !supplied {
		return DefaultCadenceSet(), nil
	}
	if raw == "" {
		return CadenceSet{}, errors.New("--cadence requires a value")
	}

	tokens := strings.Split(raw, ",")
	values := make(map[string]struct{})

	validTokens := map[string]struct{}{
		"on-change":  {},
		"ci":         {},
		"scheduled":  {},
		"production": {},
	}

	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "all" {
			if len(tokens) > 1 {
				return CadenceSet{}, errors.New(`"all" cannot be combined with other cadences`)
			}
			for v := range validTokens {
				values[v] = struct{}{}
			}
			return CadenceSet{values: values, explicit: true}, nil
		}

		if _, ok := validTokens[t]; !ok {
			return CadenceSet{}, fmt.Errorf("unknown cadence %q", t)
		}
		values[t] = struct{}{}
	}

	return CadenceSet{values: values, explicit: true}, nil
}

func (c CadenceSet) Explicit() bool {
	return c.explicit
}

func (c CadenceSet) Selects(binding *hcr.ExecutableBinding) bool {
	if len(binding.Cadence) == 0 {
		return false
	}
	for _, cadence := range binding.Cadence {
		if _, ok := c.values[cadence]; ok {
			return true
		}
	}
	return false
}
