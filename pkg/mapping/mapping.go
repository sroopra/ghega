// Package mapping provides a minimal typed mapping engine for transforming
// HL7v2 messages into destination payloads.
package mapping

// TransformType defines the kind of transformation to apply.
type TransformType string

const (
	TransformCopy      TransformType = "copy"
	TransformUppercase TransformType = "uppercase"
	TransformLowercase TransformType = "lowercase"
	TransformStatic    TransformType = "static"
	TransformCEL       TransformType = "cel"
)

// Mapping represents a single field mapping.
type Mapping struct {
	Source     string        `json:"source" yaml:"source"`
	Target     string        `json:"target" yaml:"target"`
	Transform  TransformType `json:"transform,omitempty" yaml:"transform,omitempty"`
	Value      string        `json:"value,omitempty" yaml:"value,omitempty"`
	Expression string        `json:"expression,omitempty" yaml:"expression,omitempty"`
}

// Engine applies a set of typed mappings to an input HL7v2 message.
type Engine struct {
	Mappings []Mapping
}

// NewEngine creates a new mapping engine with the given mappings.
func NewEngine(mappings []Mapping) *Engine {
	return &Engine{Mappings: mappings}
}
