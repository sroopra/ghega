// Package channel provides the schema types for Ghega channels.
package channel

import (
	"github.com/sroopra/ghega/pkg/mapping"
)

// Channel represents a deployable, testable Ghega integration channel.
type Channel struct {
	Name        string          `json:"name" yaml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Source      Source          `json:"source" yaml:"source"`
	Destination Destination     `json:"destination" yaml:"destination"`
	Mappings    []mapping.Mapping `json:"mappings" yaml:"mappings"`
	Tests       []Test          `json:"tests,omitempty" yaml:"tests,omitempty"`
	Policies    Policies        `json:"policies,omitempty" yaml:"policies,omitempty"`
}

// Source describes where messages are received from.
type Source struct {
	Type   string         `json:"type" yaml:"type"`
	Config map[string]any `json:"config,omitempty" yaml:"config,omitempty"`
}

// Destination describes where transformed messages are sent to.
type Destination struct {
	Type   string         `json:"type" yaml:"type"`
	Config map[string]any `json:"config,omitempty" yaml:"config,omitempty"`
}

// Test defines a single test fixture for a channel.
type Test struct {
	Name         string            `json:"name" yaml:"name"`
	Description  string            `json:"description,omitempty" yaml:"description,omitempty"`
	Input        string            `json:"input" yaml:"input"`
	Expected     map[string]string `json:"expected" yaml:"expected"`
	ExpectedJSON string            `json:"expectedJSON,omitempty" yaml:"expectedJSON,omitempty"`
}

// Policies holds governance and runtime constraints for a channel.
type Policies struct {
	Network NetworkPolicy `json:"network,omitempty" yaml:"network,omitempty"`
	Payload PayloadPolicy `json:"payload,omitempty" yaml:"payload,omitempty"`
	Time    TimePolicy    `json:"time,omitempty" yaml:"time,omitempty"`
}

// NetworkPolicy constrains outbound network access.
type NetworkPolicy struct {
	AllowedHosts []string `json:"allowedHosts,omitempty" yaml:"allowedHosts,omitempty"`
}

// PayloadPolicy constrains message size.
type PayloadPolicy struct {
	MaxSizeBytes int64 `json:"maxSizeBytes,omitempty" yaml:"maxSizeBytes,omitempty"`
}

// TimePolicy constrains processing duration.
type TimePolicy struct {
	MaxProcessingSeconds int `json:"maxProcessingSeconds,omitempty" yaml:"maxProcessingSeconds,omitempty"`
}
