// Package migration converts parsed Mirth channel definitions into Ghega
// Channel structs. It does not execute or interpret JavaScript.
package migration

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sroopra/ghega/pkg/channel"
	"github.com/sroopra/ghega/pkg/mirthxml"
)

// ConversionResult holds the output of converting a single Mirth channel.
type ConversionResult struct {
	Channel            channel.Channel
	Warnings           []string
	OriginalSourceType string
	OriginalDestTypes  []string
}

// ConvertChannel converts a parsed Mirth XML channel into a Ghega Channel.
// It returns warnings for unsupported features or connector types that could
// not be fully mapped.
func ConvertChannel(mch *mirthxml.Channel) (*ConversionResult, error) {
	if mch == nil {
		return nil, fmt.Errorf("mirth channel is nil")
	}

	result := &ConversionResult{
		Warnings:          make([]string, 0),
		OriginalDestTypes: make([]string, 0),
	}

	// Map basic metadata.
	gch := channel.Channel{
		Name:        sanitizeName(mch.Name),
		Description: mch.Description,
	}
	if gch.Name != mch.Name {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("channel name sanitized from %q to %q", mch.Name, gch.Name))
	}

	// Convert source connector.
	src, srcType, srcWarnings := convertSource(mch.SourceConnector)
	gch.Source = src
	result.OriginalSourceType = srcType
	result.Warnings = append(result.Warnings, srcWarnings...)

	// Convert destination connectors.
	dest, destWarnings := convertDestinations(mch.DestinationConnectors)
	gch.Destination = dest
	result.Warnings = append(result.Warnings, destWarnings...)
	for _, d := range mch.DestinationConnectors {
		result.OriginalDestTypes = append(result.OriginalDestTypes, d.Properties.Class)
	}

	// Do not attempt to convert JavaScript transformers or filters in this bead.
	if hasScripts(mch) {
		result.Warnings = append(result.Warnings,
			"JavaScript transformers/filters are present but not converted in this step")
	}

	result.Channel = gch
	return result, nil
}

func hasScripts(mch *mirthxml.Channel) bool {
	for _, step := range mch.SourceConnector.Transformer.Steps {
		if strings.TrimSpace(step.Script) != "" {
			return true
		}
	}
	for _, rule := range mch.SourceConnector.Filter.Rules {
		if strings.TrimSpace(rule.Script) != "" {
			return true
		}
	}
	for _, d := range mch.DestinationConnectors {
		for _, step := range d.Transformer.Steps {
			if strings.TrimSpace(step.Script) != "" {
				return true
			}
		}
		for _, rule := range d.Filter.Rules {
			if strings.TrimSpace(rule.Script) != "" {
				return true
			}
		}
	}
	if strings.TrimSpace(mch.PreprocessorScript) != "" {
		return true
	}
	if strings.TrimSpace(mch.PostprocessorScript) != "" {
		return true
	}
	if strings.TrimSpace(mch.DeployScript) != "" {
		return true
	}
	if strings.TrimSpace(mch.UndeployScript) != "" {
		return true
	}
	return false
}

// convertSource maps a Mirth SourceConnector to a Ghega Source.
func convertSource(src mirthxml.SourceConnector) (channel.Source, string, []string) {
	var warnings []string
	class := src.Properties.Class

	switch class {
	case "com.mirth.connect.connectors.tcp.TcpListenerProperties":
		var props mirthxml.TcpListenerProperties
		_ = src.Properties.UnmarshalInto(&props)
		return channel.Source{
			Type: "mllp",
			Config: map[string]any{
				"host": props.ListenerConnectorProperties.Host,
				"port": props.ListenerConnectorProperties.Port,
			},
		}, class, warnings

	case "com.mirth.connect.connectors.http.HttpListenerProperties":
		var props mirthxml.HttpListenerProperties
		_ = src.Properties.UnmarshalInto(&props)
		return channel.Source{
			Type: "http",
			Config: map[string]any{
				"host": props.ListenerConnectorProperties.Host,
				"port": props.ListenerConnectorProperties.Port,
			},
		}, class, warnings

	case "com.mirth.connect.connectors.file.FileReceiverProperties":
		var props mirthxml.FileReceiverProperties
		_ = src.Properties.UnmarshalInto(&props)
		return channel.Source{
			Type: "file",
			Config: map[string]any{
				"path": buildFilePath(props.Scheme, props.Host, props.Directory),
			},
		}, class, warnings

	case "com.mirth.connect.connectors.jdbc.DatabaseReaderProperties":
		warnings = append(warnings,
			"Database Reader source is mapped to type 'db' but configuration is not fully extracted")
		return channel.Source{Type: "db"}, class, warnings

	default:
		warnings = append(warnings,
			fmt.Sprintf("unsupported source connector type %q", class))
		return channel.Source{}, class, warnings
	}
}

// convertDestinations selects the primary destination from a list of Mirth
// destination connectors and converts it. Additional destinations generate
// warnings because Ghega channels currently support a single destination.
func convertDestinations(dests []mirthxml.DestinationConnector) (channel.Destination, []string) {
	var warnings []string
	if len(dests) == 0 {
		warnings = append(warnings, "no destination connectors found")
		return channel.Destination{}, warnings
	}

	primary := dests[0]
	for _, d := range dests {
		if d.Enabled {
			primary = d
			break
		}
	}

	if len(dests) > 1 {
		warnings = append(warnings,
			fmt.Sprintf("Mirth channel has %d destinations; only the primary (%q) was converted", len(dests), primary.Name))
	}

	dest, warn := convertDestination(primary)
	warnings = append(warnings, warn...)
	return dest, warnings
}

// convertDestination maps a single Mirth DestinationConnector to a Ghega Destination.
func convertDestination(dest mirthxml.DestinationConnector) (channel.Destination, []string) {
	var warnings []string
	class := dest.Properties.Class

	switch class {
	case "com.mirth.connect.connectors.http.HttpDispatcherProperties":
		var props mirthxml.HttpDispatcherProperties
		_ = dest.Properties.UnmarshalInto(&props)
		cfg := map[string]any{
			"url":    buildURL(props.Host, props.Port),
			"method": props.Method,
		}
		if props.Method == "" {
			cfg["method"] = "POST"
			warnings = append(warnings, "HTTP destination method defaulted to POST")
		}
		return channel.Destination{Type: "http", Config: cfg}, warnings

	case "com.mirth.connect.connectors.file.FileDispatcherProperties":
		var props mirthxml.FileDispatcherProperties
		_ = dest.Properties.UnmarshalInto(&props)
		return channel.Destination{
			Type: "file",
			Config: map[string]any{
				"path": buildFilePath(props.Scheme, props.Host, props.Directory),
			},
		}, warnings

	case "com.mirth.connect.connectors.jdbc.DatabaseWriterProperties":
		warnings = append(warnings,
			"Database Writer destination is mapped to type 'db' but configuration is not fully extracted")
		return channel.Destination{Type: "db"}, warnings

	default:
		warnings = append(warnings,
			fmt.Sprintf("unsupported destination connector type %q", class))
		return channel.Destination{}, warnings
	}
}

var invalidNameChars = regexp.MustCompile(`[^a-z0-9-]+`)

// sanitizeName converts a Mirth channel name into a Ghega-compatible name.
// Ghega names must match ^[a-z0-9-]+$.
func sanitizeName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = invalidNameChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	// Collapse multiple hyphens.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		s = "migrated-channel"
	}
	return s
}

// buildURL constructs a simple HTTP URL from host and port.
func buildURL(host string, port int) string {
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 80
	}
	if port == 80 {
		return fmt.Sprintf("http://%s/", host)
	}
	if port == 443 {
		return fmt.Sprintf("https://%s/", host)
	}
	return fmt.Sprintf("http://%s:%d/", host, port)
}

// buildFilePath constructs a file path from Mirth file connector properties.
func buildFilePath(scheme, host, directory string) string {
	if directory == "" {
		return "/"
	}
	return directory
}
