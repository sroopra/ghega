package channel

import (
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

	// Compare expected vs actual.
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
