package channel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sroopra/ghega/pkg/mapping"
)

// TestResult holds the outcome of a single fixture run.
type TestResult struct {
	Name     string
	Passed   bool
	Errors   []string
	Warnings []string
	Duration time.Duration
	Actual   map[string]string
}

// RunTest executes a single TestFixture through the mapping engine and
// compares the actual output with the expected values.
func RunTest(fixture TestFixture, mappings []mapping.Mapping) (*TestResult, error) {
	eng := mapping.NewEngine(mappings)

	start := time.Now()
	actual, err := eng.Apply([]byte(fixture.Input))
	duration := time.Since(start)

	result := &TestResult{
		Name:     fixture.Name,
		Passed:   true,
		Duration: duration,
		Actual:   actual,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("mapping engine error: %v", err))
		return result, nil
	}

	// Compare expected vs actual (flat string comparison).
	for key, expectedVal := range fixture.Expected {
		actualVal, ok := actual[key]
		if !ok {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("expected %q = %q, got missing key", key, expectedVal))
			continue
		}
		if actualVal != expectedVal {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("expected %q = %q, got %q", key, expectedVal, actualVal))
		}
	}

	// Warn about extra keys.
	for key := range actual {
		if _, ok := fixture.Expected[key]; !ok {
			result.Warnings = append(result.Warnings, fmt.Sprintf("extra key %q = %q", key, actual[key]))
		}
	}

	// Compare expected vs actual (JSON deep comparison).
	if fixture.ExpectedJSON != "" {
		actualJSON, err := json.Marshal(actual)
		if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("marshal actual output to JSON: %v", err))
		} else {
			var actualObj, expectedObj map[string]any
			if err := json.Unmarshal(actualJSON, &actualObj); err != nil {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("parse actual output as JSON: %v", err))
			} else if err := json.Unmarshal([]byte(fixture.ExpectedJSON), &expectedObj); err != nil {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("parse expected JSON: %v", err))
			} else {
				diffs := compareJSON(expectedObj, actualObj, "")
				if len(diffs) > 0 {
					result.Passed = false
					result.Errors = append(result.Errors, diffs...)
				}
			}
		}
	}

	return result, nil
}

// compareJSON performs a deep comparison of two JSON-compatible values and
// returns a slice of human-readable difference messages with JSON paths.
func compareJSON(expected, actual any, path string) []string {
	var diffs []string

	switch exp := expected.(type) {
	case map[string]any:
		act, ok := actual.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("expected object at %s, got %T", pathOrRoot(path), actual)}
		}
		for k, v := range exp {
			childPath := joinPath(path, k)
			av, ok := act[k]
			if !ok {
				diffs = append(diffs, fmt.Sprintf("missing key %s", childPath))
			} else {
				diffs = append(diffs, compareJSON(v, av, childPath)...)
			}
		}
		for k := range act {
			if _, ok := exp[k]; !ok {
				diffs = append(diffs, fmt.Sprintf("extra key %s", joinPath(path, k)))
			}
		}
	case []any:
		act, ok := actual.([]any)
		if !ok {
			return []string{fmt.Sprintf("expected array at %s, got %T", pathOrRoot(path), actual)}
		}
		if len(exp) != len(act) {
			diffs = append(diffs, fmt.Sprintf("expected array length %d at %s, got %d", len(exp), pathOrRoot(path), len(act)))
		}
		maxLen := len(exp)
		if len(act) > maxLen {
			maxLen = len(act)
		}
		for i := 0; i < maxLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", pathOrRoot(path), i)
			if i < len(exp) && i < len(act) {
				diffs = append(diffs, compareJSON(exp[i], act[i], childPath)...)
			} else if i >= len(exp) {
				diffs = append(diffs, fmt.Sprintf("extra array element at %s", childPath))
			} else {
				diffs = append(diffs, fmt.Sprintf("missing array element at %s", childPath))
			}
		}
	default:
		if expected != actual {
			diffs = append(diffs, fmt.Sprintf("expected %s = %v, got %v", pathOrRoot(path), expected, actual))
		}
	}

	return diffs
}

func pathOrRoot(path string) string {
	if path == "" {
		return "<root>"
	}
	return path
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}
