package runner

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
	"github.com/porscheofficial/jigctl/internal/hcr"
)

var (
	errPointerMalformed = errors.New("pointer malformed")
	errPointerNotFound  = errors.New("pointer not found")
)

// ConfigAssert evaluates a config-assert binding.
//
//nolint:gocritic // matches sibling evaluators
func ConfigAssert(plan hcr.Plan, target hcr.Target, binding hcr.ExecutableBinding) *Verdict {
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

	if binding.Op == "matches" {
		expectedStr, ok := binding.Value.(string)
		if !ok {
			return NewBlockedVerdict(&report, ReasonPatternInvalid)
		}
		if _, err := regexp.Compile(expectedStr); err != nil {
			return NewBlockedVerdict(&report, ReasonPatternInvalid)
		}
	}

	doc, blockedVerdict := readAndParseFile(&report, plan.Root, binding.File)
	if blockedVerdict != nil {
		return blockedVerdict
	}

	val, err := resolveJSONPointer(doc, binding.Path)
	if err != nil {
		if errors.Is(err, errPointerMalformed) {
			return NewBlockedVerdict(&report, ReasonPointerMalformed)
		}
		if binding.Op == "absent" {
			return NewCompletedVerdict(&report) // Pass
		}
		report.Findings = []Finding{{
			Locus:    Locus{File: binding.File, Pointer: binding.Path},
			Severity: binding.Severity,
		}}
		return NewCompletedVerdict(&report)
	}

	if binding.Op == "absent" {
		report.Findings = []Finding{{
			Locus:    Locus{File: binding.File, Pointer: binding.Path},
			Severity: binding.Severity,
		}}
		return NewCompletedVerdict(&report)
	}

	if evaluateComparison(val, binding.Op, binding.Value) {
		report.Findings = []Finding{{
			Locus:    Locus{File: binding.File, Pointer: binding.Path},
			Severity: binding.Severity,
		}}
	}

	return NewCompletedVerdict(&report)
}

func readAndParseFile(report *VerdictReport, root, file string) (interface{}, *Verdict) {
	resolvedFile, err := confine(root, file)
	if err != nil {
		return nil, NewOperationalVerdict(report, ReasonPathEscapesRoot)
	}

	content, err := os.ReadFile(resolvedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewBlockedVerdict(report, ReasonInputMissing)
		}
		return nil, NewBlockedVerdict(report, ReasonInputUnreadable)
	}

	var doc interface{}
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".json":
		err = json.Unmarshal(content, &doc)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(content, &doc)
	case ".toml":
		err = toml.Unmarshal(content, &doc)
	default:
		return nil, NewBlockedVerdict(report, ReasonFormatUnsupported)
	}

	if err != nil {
		return nil, NewBlockedVerdict(report, ReasonInputMalformed)
	}

	return doc, nil
}

func resolveJSONPointer(doc interface{}, pointer string) (interface{}, error) {
	if pointer == "" {
		return doc, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, errPointerMalformed
	}
	parts := strings.Split(pointer, "/")[1:]

	current := doc
	for _, part := range parts {
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		next, err := resolvePointerPart(current, part)
		if err != nil {
			return nil, err
		}
		current = next
	}
	return current, nil
}

func resolvePointerPart(current interface{}, part string) (interface{}, error) {
	switch v := current.(type) {
	case map[string]interface{}:
		next, ok := v[part]
		if !ok {
			return nil, errPointerNotFound
		}
		return next, nil
	case map[interface{}]interface{}:
		for mk, mv := range v {
			if mkStr, ok := mk.(string); ok && mkStr == part {
				return mv, nil
			}
		}
		return nil, errPointerNotFound
	case []interface{}:
		return resolvePointerArrayPart(v, part)
	}
	return nil, errPointerMalformed
}

func resolvePointerArrayPart(v []interface{}, part string) (interface{}, error) {
	if part == "-" {
		return nil, errPointerNotFound
	}
	idx, err := strconv.Atoi(part)
	if err != nil || idx < 0 || idx >= len(v) || (part != "0" && strings.HasPrefix(part, "0")) {
		if err != nil || (part != "0" && strings.HasPrefix(part, "0")) {
			return nil, errPointerMalformed
		}
		return nil, errPointerNotFound
	}
	return v[idx], nil
}

func evaluateComparison(actual interface{}, op string, expected interface{}) bool {
	if _, ok := actual.(time.Time); ok {
		return true // type mismatch / unsupported
	}

	switch op {
	case "equals":
		return !isEqual(actual, expected)
	case "gte", "lte":
		actualNum, aOk := toFloat64(actual)
		expectedNum, eOk := toFloat64(expected)
		if !aOk || !eOk {
			return true // type mismatch
		}
		if op == "gte" {
			return !(actualNum >= expectedNum)
		}
		return !(actualNum <= expectedNum)
	case "matches":
		actualStr, ok := actual.(string)
		if !ok {
			return true
		}
		expectedStr, ok := expected.(string)
		if !ok {
			return true
		}
		matched, err := regexp.MatchString(expectedStr, actualStr)
		if err != nil {
			return true
		}
		return !matched
	default:
		return true // unknown op
	}
}

func isEqual(a, b interface{}) bool {
	if a == b {
		return true
	}
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum == bNum
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}
