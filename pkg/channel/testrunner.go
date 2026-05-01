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

	// If ExpectedJSON is populated, perform JSON deep comparison.
	if fixture.ExpectedJSON != "" {
		var actualObj map[string]any
		actualBytes, _ := json.Marshal(actual)
		if err := json.Unmarshal(actualBytes, &actualObj); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("unmarshal actual output as JSON: %v", err))
			return result, nil
		}
		errs := deepCompare("", actualObj, fixture.ExpectedObject)
		if len(errs) > 0 {
			result.Passed = false
			result.Errors = append(result.Errors, errs...)
		}
		return result, nil
	}

	// Compare expected vs actual (flat string comparison, backward compatible).
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

// deepCompare recursively compares two JSON-decoded values and returns a list
// of human-readable path-based differences.
func deepCompare(path string, actual, expected any) []string {
	var errs []string

	switch e := expected.(type) {
	case map[string]any:
		a, ok := actual.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: expected object, got %T", path, actual))
			return errs
		}
		for k, ev := range e {
			kpath := k
			if path != "" {
				kpath = path + "." + k
			}
			av, ok := a[k]
			if !ok {
				errs = append(errs, fmt.Sprintf("%s: expected %v, got missing key", kpath, formatValue(ev)))
				continue
			}
			errs = append(errs, deepCompare(kpath, av, ev)...)
		}
		for k, av := range a {
			if _, ok := e[k]; !ok {
				kpath := k
				if path != "" {
					kpath = path + "." + k
				}
				errs = append(errs, fmt.Sprintf("%s: unexpected extra key, got %v", kpath, formatValue(av)))
			}
		}
	case []any:
		a, ok := actual.([]any)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: expected array, got %T", path, actual))
			return errs
		}
		if len(a) != len(e) {
			errs = append(errs, fmt.Sprintf("%s: expected array of length %d, got %d", path, len(e), len(a)))
		}
		minLen := len(a)
		if len(e) < minLen {
			minLen = len(e)
		}
		for i := 0; i < minLen; i++ {
			ipath := fmt.Sprintf("%s[%d]", path, i)
			errs = append(errs, deepCompare(ipath, a[i], e[i])...)
		}
	default:
		if actual != expected {
			errs = append(errs, fmt.Sprintf("%s: expected %v, got %v", path, formatValue(expected), formatValue(actual)))
		}
	}

	return errs
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
