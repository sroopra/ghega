// Package mirthxml parses Mirth Connect channel XML exports into Go structs.
// It does not execute or interpret JavaScript — it only unmarshals XML.
package mirthxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Channel represents a Mirth Connect channel export.
type Channel struct {
	XMLName               xml.Name               `xml:"channel"`
	Version               string                 `xml:"version,attr,omitempty"`
	ID                    string                 `xml:"id"`
	Name                  string                 `xml:"name"`
	Description           string                 `xml:"description"`
	Enabled               bool                   `xml:"enabled"`
	Revision              int                    `xml:"revision"`
	SourceConnector       SourceConnector        `xml:"sourceConnector"`
	DestinationConnectors []DestinationConnector `xml:"destinationConnectors>connector"`
	Properties            []ChannelProperty      `xml:"properties>property"`
	PreprocessorScript    string                 `xml:"preprocessingScript"`
	PostprocessorScript   string                 `xml:"postprocessingScript"`
	DeployScript          string                 `xml:"deployScript"`
	UndeployScript        string                 `xml:"undeployScript"`
}

// SourceConnector represents the inbound side of a Mirth channel.
type SourceConnector struct {
	XMLName     xml.Name    `xml:"sourceConnector"`
	Name        string      `xml:"name"`
	Enabled     bool        `xml:"enabled"`
	Properties  Properties  `xml:"properties"`
	Transformer Transformer `xml:"transformer"`
	Filter      Filter      `xml:"filter"`
	MetaData    MetaDataMap `xml:"metaData"`
}

// DestinationConnector represents one outbound side of a Mirth channel.
type DestinationConnector struct {
	XMLName     xml.Name    `xml:"connector"`
	Name        string      `xml:"name"`
	Enabled     bool        `xml:"enabled"`
	Properties  Properties  `xml:"properties"`
	Transformer Transformer `xml:"transformer"`
	Filter      Filter      `xml:"filter"`
	MetaData    MetaDataMap `xml:"metaData"`
}

// Properties holds the raw inner XML of a Mirth connector's properties
// element, along with the Java class attribute that identifies the connector type.
type Properties struct {
	Class string `xml:"class,attr"`
	Raw   []byte `xml:",innerxml"`
}

// UnmarshalInto unmarshals the raw inner XML of Properties into the provided
// value. Callers can use connector-specific property structs defined in this package.
func (p Properties) UnmarshalInto(v interface{}) error {
	if len(p.Raw) == 0 {
		return nil
	}
	// Wrap the raw fragment in a root element so that nested path tags
	// (e.g. xml:"listenerConnectorProperties>host") resolve correctly.
	wrapped := append([]byte("<properties>"), p.Raw...)
	wrapped = append(wrapped, []byte("</properties>")...)
	return xml.Unmarshal(wrapped, v)
}

// Transformer holds the ordered steps applied to a message.
type Transformer struct {
	XMLName xml.Name `xml:"transformer"`
	Steps   []Step   `xml:"steps>step"`
}

// Step represents a single transformer step.
type Step struct {
	XMLName          xml.Name `xml:"step"`
	SequenceNumber   int      `xml:"sequenceNumber"`
	Name             string   `xml:"name"`
	Script           string   `xml:"script"`
	Type             string   `xml:"type"`
	DataClass        string   `xml:"dataClass"`
}

// Filter holds the ordered rules evaluated against a message.
type Filter struct {
	XMLName xml.Name `xml:"filter"`
	Rules   []Rule   `xml:"rules>rule"`
}

// Rule represents a single filter rule.
type Rule struct {
	XMLName        xml.Name `xml:"rule"`
	SequenceNumber int      `xml:"sequenceNumber"`
	Name           string   `xml:"name"`
	Script         string   `xml:"script"`
	Operator       string   `xml:"operator"`
}

// MetaDataMap captures optional metadata entries on connectors.
type MetaDataMap struct {
	XMLName    xml.Name    `xml:"metaData"`
	MapEntries []MapEntry  `xml:"map>entry"`
}

// MapEntry is a generic key/value pair used in Mirth metadata.
type MapEntry struct {
	XMLName xml.Name `xml:"entry"`
	Strings []string `xml:"string"`
}

// Key returns the first string element (the map key) if present.
func (m MapEntry) Key() string {
	if len(m.Strings) > 0 {
		return m.Strings[0]
	}
	return ""
}

// Value returns the second string element (the map value) if present.
func (m MapEntry) Value() string {
	if len(m.Strings) > 1 {
		return m.Strings[1]
	}
	return ""
}

// ChannelProperty captures top-level channel properties.
type ChannelProperty struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// CodeTemplate represents a Mirth Code Template.
type CodeTemplate struct {
	XMLName     xml.Name `xml:"codeTemplate"`
	ID          string   `xml:"id"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	Code        string   `xml:"code"`
	Context     string   `xml:"context"`
	Type        string   `xml:"type"`
}

// CodeTemplateLibrary is the root of a Mirth code-template export.
type CodeTemplateLibrary struct {
	XMLName       xml.Name       `xml:"codeTemplateLibrary"`
	ID            string         `xml:"id"`
	Name          string         `xml:"name"`
	Description   string         `xml:"description"`
	CodeTemplates []CodeTemplate `xml:"codeTemplates>codeTemplate"`
}

// ---------------------------------------------------------------------------
// Typed property structs for common Mirth connectors.
// ---------------------------------------------------------------------------

// ListenerConnectorProperties is embedded in many listener-type connectors.
type ListenerConnectorProperties struct {
	Host string `xml:"host"`
	Port int    `xml:"port"`
}

// TcpListenerProperties is the property set for MLLP/TCP listeners.
type TcpListenerProperties struct {
	ListenerConnectorProperties ListenerConnectorProperties `xml:"listenerConnectorProperties"`
}

// HttpListenerProperties is the property set for HTTP listeners.
type HttpListenerProperties struct {
	ListenerConnectorProperties ListenerConnectorProperties `xml:"listenerConnectorProperties"`
}

// HttpDispatcherProperties is the property set for HTTP sender connectors.
type HttpDispatcherProperties struct {
	Host   string `xml:"host"`
	Port   int    `xml:"port"`
	Method string `xml:"method"`
	Secure bool   `xml:"secure"`
}

// FileReceiverProperties is the property set for file-reader connectors.
type FileReceiverProperties struct {
	Scheme    string `xml:"scheme"`
	Host      string `xml:"host"`
	Directory string `xml:"directory"`
}

// FileDispatcherProperties is the property set for file-writer connectors.
type FileDispatcherProperties struct {
	Scheme    string `xml:"scheme"`
	Host      string `xml:"host"`
	Directory string `xml:"directory"`
}

// DatabaseReaderProperties is the property set for database-reader (source) connectors.
type DatabaseReaderProperties struct {
	Driver   string `xml:"driver"`
	URL      string `xml:"url"`
	Username string `xml:"username"`
	Password string `xml:"password"`
	Query    string `xml:"query"`
}

// DatabaseWriterProperties is the property set for database-writer (destination) connectors.
type DatabaseWriterProperties struct {
	Driver   string `xml:"driver"`
	URL      string `xml:"url"`
	Username string `xml:"username"`
	Password string `xml:"password"`
}

// SftpReceiverProperties is the property set for SFTP reader connectors.
type SftpReceiverProperties struct {
	Host     string `xml:"host"`
	Port     int    `xml:"port"`
	Username string `xml:"username"`
	Password string `xml:"password"`
	Path     string `xml:"path"`
}

// SftpDispatcherProperties is the property set for SFTP writer connectors.
type SftpDispatcherProperties struct {
	Host     string `xml:"host"`
	Port     int    `xml:"port"`
	Username string `xml:"username"`
	Password string `xml:"password"`
	Path     string `xml:"path"`
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

// ParseChannel unmarshals a single Mirth channel from XML bytes.
func ParseChannel(data []byte) (*Channel, error) {
	var ch Channel
	if err := xml.Unmarshal(data, &ch); err != nil {
		return nil, fmt.Errorf("unmarshal channel: %w", err)
	}
	return &ch, nil
}

// ParseChannelFile reads and unmarshals a single Mirth channel XML file.
func ParseChannelFile(path string) (*Channel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read channel file %s: %w", path, err)
	}
	ch, err := ParseChannel(data)
	if err != nil {
		return nil, fmt.Errorf("parse channel file %s: %w", path, err)
	}
	return ch, nil
}

// ParseChannelsFromDir walks a directory and parses every file ending in .xml
// as a Mirth channel. Files that fail to parse are skipped without error;
// the function returns only channels that parsed successfully.
func ParseChannelsFromDir(dir string) ([]*Channel, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var channels []*Channel
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xml" {
			continue
		}
		ch, err := ParseChannelFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			// Skip files that are not valid channel XML.
			continue
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

// ParseCodeTemplate unmarshals a single Mirth Code Template from XML bytes.
func ParseCodeTemplate(data []byte) (*CodeTemplate, error) {
	var ct CodeTemplate
	if err := xml.Unmarshal(data, &ct); err != nil {
		return nil, fmt.Errorf("unmarshal code template: %w", err)
	}
	return &ct, nil
}

// ParseCodeTemplateFile reads and unmarshals a single Mirth code-template XML file.
func ParseCodeTemplateFile(path string) (*CodeTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read code template file %s: %w", path, err)
	}
	ct, err := ParseCodeTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("parse code template file %s: %w", path, err)
	}
	return ct, nil
}

// ParseCodeTemplatesFromDir walks a directory and parses every file ending in
// .xml as a Mirth code template. Files that fail to parse are skipped.
func ParseCodeTemplatesFromDir(dir string) ([]*CodeTemplate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var templates []*CodeTemplate
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xml" {
			continue
		}
		ct, err := ParseCodeTemplateFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		templates = append(templates, ct)
	}
	return templates, nil
}

// ParseCodeTemplateLibrary unmarshals a code-template library XML document.
func ParseCodeTemplateLibrary(data []byte) (*CodeTemplateLibrary, error) {
	var lib CodeTemplateLibrary
	if err := xml.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("unmarshal code template library: %w", err)
	}
	return &lib, nil
}
