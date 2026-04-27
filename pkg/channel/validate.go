package channel

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Field   string
	Message string
}

var (
	namePattern       = regexp.MustCompile(`^[a-z0-9-]+$`)
	hl7PathPattern    = regexp.MustCompile(`^[A-Z]{3}-\d+(\.\d+)?$`)
	validSourceTypes  = map[string]bool{"mllp": true, "http": true, "file": true, "sftp": true, "db": true}
	validDestTypes    = map[string]bool{"http": true, "file": true, "sftp": true, "db": true}
)

// ValidateYAML parses YAML into a Channel and returns validation errors.
func ValidateYAML(data []byte) (*Channel, []ValidationError) {
	var ch Channel
	var errs []ValidationError

	if err := yaml.Unmarshal(data, &ch); err != nil {
		errs = append(errs, ValidationError{Field: "", Message: fmt.Sprintf("invalid YAML: %v", err)})
		return nil, errs
	}

	errs = append(errs, validateChannel(&ch)...)
	return &ch, errs
}

func validateChannel(ch *Channel) []ValidationError {
	var errs []ValidationError

	if ch.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	} else if !namePattern.MatchString(ch.Name) {
		errs = append(errs, ValidationError{Field: "name", Message: "name must match ^[a-z0-9-]+$"})
	}

	if ch.Source.Type == "" {
		errs = append(errs, ValidationError{Field: "source.type", Message: "source.type is required"})
	} else if !validSourceTypes[ch.Source.Type] {
		errs = append(errs, ValidationError{Field: "source.type", Message: fmt.Sprintf("source.type must be one of: mllp, http, file, sftp, db (got %q)", ch.Source.Type)})
	}

	if ch.Destination.Type == "" {
		errs = append(errs, ValidationError{Field: "destination.type", Message: "destination.type is required"})
	} else if !validDestTypes[ch.Destination.Type] {
		errs = append(errs, ValidationError{Field: "destination.type", Message: fmt.Sprintf("destination.type must be one of: http, file, sftp, db (got %q)", ch.Destination.Type)})
	}

	if len(ch.Mappings) == 0 {
		errs = append(errs, ValidationError{Field: "mappings", Message: "mappings must not be empty"})
	}

	for i, m := range ch.Mappings {
		if m.Source == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("mappings[%d].source", i), Message: "mapping source is required"})
		} else if !hl7PathPattern.MatchString(m.Source) {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("mappings[%d].source", i), Message: fmt.Sprintf("mapping source %q must match a valid HL7 path pattern", m.Source)})
		}
		if m.Target == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("mappings[%d].target", i), Message: "mapping target is required"})
		}
	}

	for i, test := range ch.Tests {
		if test.Name == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("tests[%d].name", i), Message: "test name is required"})
		}
		if test.Input == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("tests[%d].input", i), Message: "test input is required"})
		}
	}

	return errs
}
