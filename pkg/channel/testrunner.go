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

	// JSON deep comparison takes precedence when ExpectedJSON is set.
	if fixture.ExpectedJSON != "" {
		actualJSON, err := json.Marshal(actual)
		if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("marshal actual output: %v", err))
			return result, nil
		}
		var actualObject map[string]any
		if err := json.Unmarshal(actualJSON, &actualObject); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("unmarshal actual output: %v", err))
			return result, nil
		}
		diffs := jsonDiff("", actualObject, fixture.ExpectedObject)
		if len(diffs) > 0 {
			result.Passed = false
			result.Errors = append(result.Errors, diffs...)
		}
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

	return result, nil
}

// jsonDiff recursively compares two JSON-compatible values and returns a list
// of human-readable differences with JSON-path-style keys.
func jsonDiff(path string, actual, expected any) []string {
	var diffs []string

	switch exp := expected.(type) {
	case map[string]any:
		act, ok := actual.(map[string]any)
		if !ok {
			return append(diffs, fmt.Sprintf("%s: expected object, got %T", path, actual))
		}
		for k, v := range exp {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}
			av, exists := act[k]
			if !exists {
				diffs = append(diffs, fmt.Sprintf("%s: expected %v, got missing key", childPath, v))
				continue
			}
			diffs = append(diffs, jsonDiff(childPath, av, v)...)
		}
		for k := range act {
			if _, exists := exp[k]; !exists {
				childPath := k
				if path != "" {
					childPath = path + "." + k
				}
				diffs = append(diffs, fmt.Sprintf("%s: unexpected key with value %v", childPath, act[k]))
			}
		}
	case []any:
		act, ok := actual.([]any)
		if !ok {
			return append(diffs, fmt.Sprintf("%s: expected array, got %T", path, actual))
		}
		if len(act) != len(exp) {
			diffs = append(diffs, fmt.Sprintf("%s: expected array of length %d, got %d", path, len(exp), len(act)))
		}
		minLen := len(act)
		if len(exp) < minLen {
			minLen = len(exp)
		}
		for i := 0; i < minLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			diffs = append(diffs, jsonDiff(childPath, act[i], exp[i])...)
		}
	default:
		if actual != expected {
			diffs = append(diffs, fmt.Sprintf("%s: expected %v, got %v", path, expected, actual))
		}
	}

	return diffs
}
