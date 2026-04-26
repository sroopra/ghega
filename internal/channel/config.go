// Package channel wires together the MLLP listener, message store, mapping
// engine, and HTTP sender into a minimal end-to-end integration channel.
package channel

import (
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

// SourceConfig describes the inbound side of a channel.
type SourceConfig struct {
	Type string `json:"type" yaml:"type"`
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

// DestinationConfig describes the outbound side of a channel.
type DestinationConfig struct {
	Type string `json:"type" yaml:"type"`
	URL  string `json:"url" yaml:"url"`
}

// MappingConfig holds mapping rules for a channel.
type MappingConfig struct {
	MessageType string           `json:"messageType,omitempty" yaml:"messageType,omitempty"`
	Mappings    []mapping.Mapping `json:"mappings,omitempty" yaml:"mappings,omitempty"`
}

// Config represents a channel definition loaded from YAML.
type Config struct {
	Name        string            `json:"name" yaml:"name"`
	Source      SourceConfig      `json:"source" yaml:"source"`
	Destination DestinationConfig `json:"destination" yaml:"destination"`
	Mapping     MappingConfig     `json:"mapping" yaml:"mapping"`
}

// LoadConfig reads a channel definition from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read channel config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse channel config: %w", err)
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("channel config missing required field: name")
	}
	if cfg.Source.Type == "" {
		cfg.Source.Type = "mllp"
	}
	if cfg.Destination.Type == "" {
		cfg.Destination.Type = "http"
	}

	return &cfg, nil
}
