// Package migration classifies JavaScript/E4X transformer code from Mirth
// channels into typed patterns and decides what can be auto-converted to Ghega
// mappings versus what needs a rewrite task.
package migration

import (
	"regexp"
	"strings"

	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/mirthxml"
)

// PatternCategory identifies the kind of construct found in a transformer step.
type PatternCategory string

const (
	CategoryFieldAssignment     PatternCategory = "field_assignment"
	CategoryConditional         PatternCategory = "conditional"
	CategoryE4XManipulation     PatternCategory = "e4x_manipulation"
	CategoryLoop                PatternCategory = "loop"
	CategoryDestinationDispatch PatternCategory = "destination_dispatch"
	CategoryLogger              PatternCategory = "logger"
	CategoryExternalCall        PatternCategory = "external_call"
)

// Disposition indicates whether a pattern can be auto-converted.
type Disposition string

const (
	DispositionAutoConvertible Disposition = "auto_convertible"
	DispositionNeedsRewrite    Disposition = "needs_rewrite"
	DispositionUnsupported     Disposition = "unsupported"
)

// RewriteTask describes work required for a pattern that cannot be auto-converted.
type RewriteTask struct {
	Severity    string `json:"severity" yaml:"severity"`
	Description string `json:"description" yaml:"description"`
}

// ClassifiedPattern is the result of analysing a single pattern occurrence.
type ClassifiedPattern struct {
	Category    PatternCategory  `json:"category" yaml:"category"`
	Disposition Disposition      `json:"disposition" yaml:"disposition"`
	Description string           `json:"description" yaml:"description"`
	Mapping     *mapping.Mapping `json:"mapping,omitempty" yaml:"mapping,omitempty"`
	RewriteTask *RewriteTask     `json:"rewrite_task,omitempty" yaml:"rewrite_task,omitempty"`
}

// ClassificationResult is the aggregate output for one transformer step.
type ClassificationResult struct {
	Step     mirthxml.Step       `json:"step" yaml:"step"`
	Patterns []ClassifiedPattern `json:"patterns" yaml:"patterns"`
	Status   string              `json:"status" yaml:"status"`
	Warnings []string            `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

var (
	assignRe       = regexp.MustCompile(`(?m)^\s*(msg(?:\['[^']+'\])+)\s*=\s*(.*?)\s*;?\s*$`)
	conditionalRe  = regexp.MustCompile(`(?m)\b(if|switch)\b`)
	e4xRe          = regexp.MustCompile(`(?m)(new\s+XML\s*\(|\.{2}\s*\*|\.\s*@|\belements\s*\(|\bdescendants\s*\(|\bchildren\s*\(|\bnamespace\s*\()`)
	loopRe         = regexp.MustCompile(`(?m)\b(for\s*\(|for\s+each\s*\(|while\s*\()`)
	destDispatchRe = regexp.MustCompile(`(?m)\b(destinationSet|router\s*\.\s*routeMessage|responseMap)\b`)
	loggerRe       = regexp.MustCompile(`(?m)\b(logger\s*\.\s*(info|debug|error|warn|trace)|print\s*\(|debug\s*\()`)
	msgAccessRe    = regexp.MustCompile(`msg(?:\['[^']+'\])+`)
	stringLiteralRe = regexp.MustCompile(`^["'](.*)["']$`)
	staticValueRe  = regexp.MustCompile(`^\d+(\.\d+)?$|^(true|false|null)$`)
	externalCallRe = regexp.MustCompile(`(?m)([a-zA-Z_$][a-zA-Z0-9_$]*(?:\s*\.\s*[a-zA-Z_$][a-zA-Z0-9_$]*)*)\s*\(`)
)

var jsKeywords = map[string]bool{
	"if": true, "switch": true, "for": true, "while": true, "return": true,
	"function": true, "var": true, "let": true, "const": true, "new": true,
	"typeof": true, "instanceof": true, "void": true, "delete": true,
	"try": true, "catch": true, "finally": true, "throw": true,
	"true": true, "false": true, "null": true, "undefined": true,
	"this": true, "else": true, "do": true, "in": true, "of": true,
}

var knownCallRoots = map[string]bool{
	"msg": true, "logger": true, "destinationSet": true, "router": true,
	"responseMap": true, "print": true, "debug": true, "XML": true,
	// JavaScript built-ins — not external libraries, but will need Go equivalents
	"JSON": true, "Date": true, "Array": true, "Math": true, "String": true, "Number": true,
	"parseInt": true, "parseFloat": true, "isNaN": true, "isFinite": true,
	"encodeURI": true, "decodeURI": true, "encodeURIComponent": true, "decodeURIComponent": true,
	"escape": true, "unescape": true, "eval": true,
}

// ClassifyTransformerStep analyses a single Mirth transformer step and returns
// a classification result.  It never executes JavaScript; only regex-based
// static analysis is performed.
func ClassifyTransformerStep(step mirthxml.Step) ClassificationResult {
	result := ClassificationResult{
		Step:     step,
		Patterns: []ClassifiedPattern{},
	}

	if strings.TrimSpace(step.Script) == "" {
		result.Status = "auto_converted"
		return result
	}

	lines := strings.Split(step.Script, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		patterns := classifyLine(line)
		result.Patterns = append(result.Patterns, patterns...)
	}

	result.Status = deriveStatus(result.Patterns)
	return result
}

func classifyLine(line string) []ClassifiedPattern {
	var patterns []ClassifiedPattern

	if loopRe.MatchString(line) {
		patterns = append(patterns, ClassifiedPattern{
			Category:    CategoryLoop,
			Disposition: DispositionNeedsRewrite,
			Description: "Loop construct detected; must be rewritten as a typed mapping or custom Go logic.",
			RewriteTask: &RewriteTask{
				Severity:    "high",
				Description: "Replace loop with an equivalent mapping or custom transformation.",
			},
		})
	}

	if conditionalRe.MatchString(line) {
		patterns = append(patterns, ClassifiedPattern{
			Category:    CategoryConditional,
			Disposition: DispositionNeedsRewrite,
			Description: "Conditional logic detected; must be rewritten as a typed mapping or custom Go logic.",
			RewriteTask: &RewriteTask{
				Severity:    "high",
				Description: "Replace conditional with an equivalent mapping or custom transformation.",
			},
		})
	}

	if loggerRe.MatchString(line) {
		patterns = append(patterns, ClassifiedPattern{
			Category:    CategoryLogger,
			Disposition: DispositionNeedsRewrite,
			Description: "Logger or debug statement detected; replace with structured logging.",
			RewriteTask: &RewriteTask{
				Severity:    "low",
				Description: "Replace logger/debug statement with structured logging call.",
			},
		})
	}

	if destDispatchRe.MatchString(line) {
		patterns = append(patterns, ClassifiedPattern{
			Category:    CategoryDestinationDispatch,
			Disposition: DispositionNeedsRewrite,
			Description: "Destination dispatch call detected; must be wired into Ghega destination configuration.",
			RewriteTask: &RewriteTask{
				Severity:    "medium",
				Description: "Replace destination dispatch with Ghega destination routing configuration.",
			},
		})
	}

	if e4xRe.MatchString(line) {
		patterns = append(patterns, ClassifiedPattern{
			Category:    CategoryE4XManipulation,
			Disposition: DispositionUnsupported,
			Description: "E4X/XML manipulation detected; not yet supported by Ghega.",
			RewriteTask: &RewriteTask{
				Severity:    "high",
				Description: "Rewrite E4X/XML manipulation using Ghega's supported transformation approach.",
			},
		})
	}

	if matches := assignRe.FindAllStringSubmatch(line, -1); len(matches) > 0 {
		for _, m := range matches {
			lhs := m[1]
			rhs := strings.TrimSpace(m[2])
			segments := extractBracketChain(lhs)
			if len(segments) == 0 {
				continue
			}
			targetPath := toHL7Path(segments[len(segments)-1])

			cp := ClassifiedPattern{
				Category: CategoryFieldAssignment,
			}

			if srcAccess := msgAccessRe.FindString(rhs); srcAccess != "" && srcAccess == rhs {
				srcSegments := extractBracketChain(rhs)
				srcPath := ""
				if len(srcSegments) > 0 {
					srcPath = toHL7Path(srcSegments[len(srcSegments)-1])
				}
				cp.Disposition = DispositionAutoConvertible
				cp.Description = "Simple field assignment from another message field."
				cp.Mapping = &mapping.Mapping{
					Source:    srcPath,
					Target:    targetPath,
					Transform: mapping.TransformCopy,
				}
			} else if isStaticValue(rhs) {
				val := extractStaticValue(rhs)
				cp.Disposition = DispositionAutoConvertible
				cp.Description = "Simple field assignment from a static value."
				cp.Mapping = &mapping.Mapping{
					Target:    targetPath,
					Transform: mapping.TransformStatic,
					Value:     val,
				}
			} else {
				cp.Disposition = DispositionNeedsRewrite
				cp.Description = "Field assignment with a complex right-hand side expression."
				cp.RewriteTask = &RewriteTask{
					Severity:    "medium",
					Description: "Rewrite field assignment with complex expression into a typed mapping or custom Go code.",
				}
			}

			patterns = append(patterns, cp)
		}
	}

	if len(patterns) == 0 {
		seen := make(map[string]bool)
		extMatches := externalCallRe.FindAllStringSubmatch(line, -1)
		for _, em := range extMatches {
			chain := strings.ReplaceAll(em[1], " ", "")
			parts := strings.Split(chain, ".")
			name := parts[0]
			if jsKeywords[name] || knownCallRoots[name] {
				continue
			}
			if seen[chain] {
				continue
			}
			seen[chain] = true
			patterns = append(patterns, ClassifiedPattern{
				Category:    CategoryExternalCall,
				Disposition: DispositionUnsupported,
				Description: "External function call or library usage detected: " + chain + "().",
				RewriteTask: &RewriteTask{
					Severity:    "high",
					Description: "Replace external function call '" + chain + "()' with a Ghega-native equivalent or custom Go code.",
				},
			})
		}
	}

	return patterns
}

func extractBracketChain(s string) []string {
	re := regexp.MustCompile(`\['([^']+)'\]`)
	matches := re.FindAllStringSubmatch(s, -1)
	var out []string
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func toHL7Path(s string) string {
	if idx := strings.Index(s, "."); idx >= 0 {
		return s[:idx] + "-" + s[idx+1:]
	}
	return s
}

func isStaticValue(s string) bool {
	return staticValueRe.MatchString(s) || stringLiteralRe.MatchString(s)
}

func extractStaticValue(s string) string {
	if m := stringLiteralRe.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

func deriveStatus(patterns []ClassifiedPattern) string {
	if len(patterns) == 0 {
		return "auto_converted"
	}
	hasAuto := false
	hasRewrite := false
	hasUnsupported := false
	for _, p := range patterns {
		switch p.Disposition {
		case DispositionAutoConvertible:
			hasAuto = true
		case DispositionNeedsRewrite:
			hasRewrite = true
		case DispositionUnsupported:
			hasUnsupported = true
		}
	}
	if hasUnsupported {
		if hasAuto || hasRewrite {
			return "mixed"
		}
		return "unsupported"
	}
	if hasRewrite {
		if hasAuto {
			return "mixed"
		}
		return "needs_rewrite"
	}
	return "auto_converted"
}
