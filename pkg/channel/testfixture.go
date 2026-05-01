package channel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestFixture is a fully-resolved test case ready to be executed.
type TestFixture struct {
	Name          string            `json:"name" yaml:"name"`
	Description   string            `json:"description,omitempty" yaml:"description,omitempty"`
	Input         string            `json:"input" yaml:"input"`
	Expected      map[string]string `json:"expected" yaml:"expected"`
	ExpectedJSON  string            `json:"expectedJson,omitempty" yaml:"expectedJson,omitempty"`
	ExpectedObject map[string]any   `json:"-" yaml:"-"`
}

// LoadTestFixtures resolves a slice of Test definitions into TestFixtures.
// If Input looks like a file path (contains "/" or ends in ".hl7" or ".json"), it is
// read from disk relative to the channel directory; otherwise it is treated
// as inline HL7.
func LoadTestFixtures(channelPath string, tests []Test) ([]TestFixture, error) {
	channelDir := filepath.Dir(channelPath)
	fixtures := make([]TestFixture, 0, len(tests))

	for _, tt := range tests {
		input := tt.Input
		if looksLikeFilePath(input) {
			data, err := loadFixtureFile(channelDir, tt.Name, input)
			if err != nil {
				return nil, err
			}
			input = string(data)
		}

		expectedJSON := tt.ExpectedJSON
		if looksLikeFilePath(expectedJSON) {
			data, err := loadFixtureFile(channelDir, tt.Name, expectedJSON)
			if err != nil {
				return nil, err
			}
			expectedJSON = string(data)
		}

		var expectedObject map[string]any
		if expectedJSON != "" {
			if err := json.Unmarshal([]byte(expectedJSON), &expectedObject); err != nil {
				return nil, fmt.Errorf("load fixture %q: unmarshal expected JSON: %w", tt.Name, err)
			}
		}

		fixtures = append(fixtures, TestFixture{
			Name:           tt.Name,
			Description:    tt.Description,
			Input:          input,
			Expected:       tt.Expected,
			ExpectedJSON:   expectedJSON,
			ExpectedObject: expectedObject,
		})
	}

	return fixtures, nil
}

func loadFixtureFile(channelDir, testName, input string) ([]byte, error) {
	p := filepath.Join(channelDir, input)
	absP, err := filepath.Abs(p)
	if err != nil {
		return nil, fmt.Errorf("load fixture %q: resolve path %q: %w", testName, p, err)
	}
	absChannelDir, err := filepath.Abs(channelDir)
	if err != nil {
		return nil, fmt.Errorf("load fixture %q: resolve channel dir %q: %w", testName, channelDir, err)
	}
	rel, err := filepath.Rel(absChannelDir, absP)
	if err != nil {
		return nil, fmt.Errorf("load fixture %q: path %q is outside channel directory", testName, p)
	}
	if strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("load fixture %q: path traversal detected in %q", testName, p)
	}
	data, err := os.ReadFile(absP)
	if err != nil {
		return nil, fmt.Errorf("load fixture %q: read %q: %w", testName, p, err)
	}
	return data, nil
}

func looksLikeFilePath(s string) bool {
	return strings.Contains(s, "/") || strings.HasSuffix(s, ".hl7") || strings.HasSuffix(s, ".json")
}
