// Package channel provides the end-to-end integration channel for Ghega.
// It wires together the MLLP listener, message store, mapping engine, and
// HTTP sender into a single runnable unit.
package channel

import (
	"fmt"
	"os"

	"github.com/sroopra/ghega/pkg/mapping"
	"gopkg.in/yaml.v3"
)

// SourceConfig describes the inbound side of a channel.
type SourceConfig struct {
	Type string `yaml:"type"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DestinationConfig describes the outbound side of a channel.
type DestinationConfig struct {
	Type    string `yaml:"type"`
	URL     string `yaml:"url"`
	Method  string `yaml:"method,omitempty"`
	Timeout int    `yaml:"timeout,omitempty"`
}

// MappingConfig holds mapping rules for a channel.
type MappingConfig struct {
	MessageType string          `yaml:"messageType"`
	Fields      []mapping.Mapping `yaml:"fields"`
}

// Config is the top-level channel definition loaded from YAML.
type Config struct {
	Name        string            `yaml:"name"`
	Source      SourceConfig      `yaml:"source"`
	Destination DestinationConfig `yaml:"destination"`
	Mapping     MappingConfig     `yaml:"mapping"`
}

// Validate checks that the configuration is usable.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("channel name is required")
	}
	if c.Source.Type != "mllp" {
		return fmt.Errorf("unsupported source type: %s", c.Source.Type)
	}
	if c.Destination.Type != "http" {
		return fmt.Errorf("unsupported destination type: %s", c.Destination.Type)
	}
	if c.Destination.URL == "" {
		return fmt.Errorf("destination URL is required")
	}
	return nil
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

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate channel config: %w", err)
	}

	return &cfg, nil
}
