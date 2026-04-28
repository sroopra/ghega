package channel

import (
	"fmt"
	"strings"
	"time"
)

// ToJUnit generates a minimal JUnit XML report from a slice of TestResults.
func ToJUnit(results []TestResult) string {
	var b strings.Builder
	now := time.Now().Format(time.RFC3339)

	failures := 0
	for _, r := range results {
		if !r.Passed {
			failures++
		}
	}

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<testsuites tests="%d" failures="%d" time="%.3f" timestamp="%s">`+"\n",
		len(results), failures, totalDuration(results).Seconds(), now)
	fmt.Fprintf(&b, `  <testsuite name="ghega-channel-tests" tests="%d" failures="%d" time="%.3f">`+"\n",
		len(results), failures, totalDuration(results).Seconds())

	for _, r := range results {
		name := xmlEscape(r.Name)
		fmt.Fprintf(&b, `    <testcase name="%s" time="%.3f">`, name, r.Duration.Seconds())
		if !r.Passed {
			b.WriteString("\n")
			msg := xmlEscape(strings.Join(r.Errors, "; "))
			fmt.Fprintf(&b, `      <failure message="%s"></failure>`, msg)
			b.WriteString("\n    ")
		}
		b.WriteString("</testcase>\n")
	}

	b.WriteString("  </testsuite>\n")
	b.WriteString("</testsuites>\n")

	return b.String()
}

func totalDuration(results []TestResult) time.Duration {
	var total time.Duration
	for _, r := range results {
		total += r.Duration
	}
	return total
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
