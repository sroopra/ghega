package channel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestFixture is a fully-resolved test case ready to be executed.
type TestFixture struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Input       string            `json:"input" yaml:"input"`
	Expected    map[string]string `json:"expected" yaml:"expected"`
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
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, fmt.Errorf("load fixture %q: read %q: %w", tt.Name, p, err)
			}
			input = string(data)
		}

		fixtures = append(fixtures, TestFixture{
			Name:        tt.Name,
			Description: tt.Description,
			Input:       input,
			Expected:    tt.Expected,
		})
	}

	return fixtures, nil
}

func looksLikeFilePath(s string) bool {
	return strings.Contains(s, "/") || strings.HasSuffix(s, ".hl7")
}
