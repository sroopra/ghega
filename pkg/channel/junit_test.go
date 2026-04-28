package channel

import (
	"strings"
	"testing"
	"time"
)

func TestToJUnit_Empty(t *testing.T) {
	xml := ToJUnit(nil)
	if !strings.Contains(xml, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("expected XML declaration")
	}
	if !strings.Contains(xml, `<testsuites tests="0" failures="0"`) {
		t.Errorf("expected empty testsuites, got:\n%s", xml)
	}
}

func TestToJUnit_AllPass(t *testing.T) {
	results := []TestResult{
		{Name: "test-a", Passed: true, Duration: 10 * time.Millisecond},
		{Name: "test-b", Passed: true, Duration: 20 * time.Millisecond},
	}
	xml := ToJUnit(results)

	if !strings.Contains(xml, `tests="2"`) {
		t.Errorf("expected tests=\"2\", got:\n%s", xml)
	}
	if !strings.Contains(xml, `failures="0"`) {
		t.Errorf("expected failures=\"0\", got:\n%s", xml)
	}
	if !strings.Contains(xml, `<testcase name="test-a"`) {
		t.Errorf("expected testcase test-a, got:\n%s", xml)
	}
	if !strings.Contains(xml, `<testcase name="test-b"`) {
		t.Errorf("expected testcase test-b, got:\n%s", xml)
	}
	if strings.Contains(xml, `<failure`) {
		t.Error("expected no failure elements")
	}
}

func TestToJUnit_WithFailures(t *testing.T) {
	results := []TestResult{
		{
			Name:     "fail-a",
			Passed:   false,
			Duration: 5 * time.Millisecond,
			Errors:   []string{`expected "key" = "a", got "b"`},
		},
		{
			Name:     "pass-b",
			Passed:   true,
			Duration: 15 * time.Millisecond,
		},
	}
	xml := ToJUnit(results)

	if !strings.Contains(xml, `tests="2"`) {
		t.Errorf("expected tests=\"2\", got:\n%s", xml)
	}
	if !strings.Contains(xml, `failures="1"`) {
		t.Errorf("expected failures=\"1\", got:\n%s", xml)
	}
	if !strings.Contains(xml, `<failure message="`) {
		t.Errorf("expected failure element, got:\n%s", xml)
	}
	if !strings.Contains(xml, `expected &quot;key&quot; = &quot;a&quot;, got &quot;b&quot;`) {
		t.Errorf("expected escaped failure message, got:\n%s", xml)
	}
}

func TestToJUnit_XMLEscaping(t *testing.T) {
	results := []TestResult{
		{
			Name:     "special-chars",
			Passed:   false,
			Duration: 1 * time.Millisecond,
			Errors:   []string{`a < b & c > d "e"`},
		},
	}
	xml := ToJUnit(results)

	if strings.Contains(xml, `a < b`) {
		t.Error("expected < to be escaped")
	}
	if !strings.Contains(xml, `a &lt; b &amp; c &gt; d &quot;e&quot;`) {
		t.Errorf("expected fully escaped message, got:\n%s", xml)
	}
}

func TestToJUnit_Structure(t *testing.T) {
	results := []TestResult{
		{Name: "suite-test", Passed: true, Duration: 1 * time.Millisecond},
	}
	xml := ToJUnit(results)

	if !strings.Contains(xml, `<testsuite name="ghega-channel-tests"`) {
		t.Errorf("expected testsuite name, got:\n%s", xml)
	}
	if !strings.Contains(xml, `</testsuite>`) {
		t.Error("expected closing testsuite tag")
	}
	if !strings.Contains(xml, `</testsuites>`) {
		t.Error("expected closing testsuites tag")
	}
}
