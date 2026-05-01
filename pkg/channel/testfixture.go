package channel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestFixture is a fully-resolved test case ready to be executed.
type TestFixture struct {
	Name         string            `json:"name" yaml:"name"`
	Description  string            `json:"description,omitempty" yaml:"description,omitempty"`
	Input        string            `json:"input" yaml:"input"`
	Expected     map[string]string `json:"expected" yaml:"expected"`
	ExpectedJSON string            `json:"expectedJSON,omitempty" yaml:"expectedJSON,omitempty"`
}

// LoadTestFixtures resolves a slice of Test definitions into TestFixtures.
// If Input looks like a file path (contains "/" or ends in ".hl7"), it is
// read from disk relative to the channel directory; otherwise it is treated
// as inline HL7.
func LoadTestFixtures(channelPath string, tests []Test) ([]TestFixture, error) {
	channelDir := filepath.Dir(channelPath)
	fixtures := make([]TestFixture, 0, len(tests))

	for _, tt := range tests {
		input := tt.Input
		if looksLikeFilePath(input) {
			p := filepath.Join(channelDir, input)
			absP, err := filepath.Abs(p)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: resolve path %q: %w", tt.Name, p, err)
			}
			absChannelDir, err := filepath.Abs(channelDir)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: resolve channel dir %q: %w", tt.Name, channelDir, err)
			}
			rel, err := filepath.Rel(absChannelDir, absP)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: path %q is outside channel directory", tt.Name, p)
			}
			if strings.HasPrefix(rel, "..") {
				return nil, fmt.Errorf("load fixture %q: path traversal detected in %q", tt.Name, p)
			}
			data, err := os.ReadFile(absP)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: read %q: %w", tt.Name, p, err)
			}
			input = string(data)
		}

		expectedJSON := tt.ExpectedJSON
		if looksLikeFilePath(expectedJSON) {
			p := filepath.Join(channelDir, expectedJSON)
			absP, err := filepath.Abs(p)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: resolve expectedJSON path %q: %w", tt.Name, p, err)
			}
			absChannelDir, err := filepath.Abs(channelDir)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: resolve channel dir %q: %w", tt.Name, channelDir, err)
			}
			rel, err := filepath.Rel(absChannelDir, absP)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: expectedJSON path %q is outside channel directory", tt.Name, p)
			}
			if strings.HasPrefix(rel, "..") {
				return nil, fmt.Errorf("load fixture %q: path traversal detected in expectedJSON %q", tt.Name, p)
			}
			data, err := os.ReadFile(absP)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: read expectedJSON %q: %w", tt.Name, p, err)
			}
			expectedJSON = string(data)
		}

		fixtures = append(fixtures, TestFixture{
			Name:         tt.Name,
			Description:  tt.Description,
			Input:        input,
			Expected:     tt.Expected,
			ExpectedJSON: expectedJSON,
		})
	}

	return fixtures, nil
}

func looksLikeFilePath(s string) bool {
	return strings.Contains(s, "/") || strings.HasSuffix(s, ".hl7") || strings.HasSuffix(s, ".json")
}
