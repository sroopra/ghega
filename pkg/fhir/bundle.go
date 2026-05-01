package fhir

import (
	"encoding/json"
	"fmt"
)

// ParseBundle parses a FHIR Bundle from JSON bytes.
func ParseBundle(data []byte) (*Bundle, error) {
	var b Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse bundle: %w", err)
	}
	return &b, nil
}

// BundleToJSON marshals a FHIR Bundle to JSON bytes.
func BundleToJSON(b *Bundle) ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("bundle is nil")
	}
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal bundle: %w", err)
	}
	return data, nil
}

// ValidateBundleType ensures the bundle type is one of the allowed values.
func ValidateBundleType(b *Bundle) error {
	if b == nil {
		return fmt.Errorf("bundle is nil")
	}
	switch b.Type {
	case "batch", "transaction", "search-set", "history":
		return nil
	default:
		return fmt.Errorf("invalid bundle type %q", b.Type)
	}
}

// IterateEntries iterates over bundle entries, calling fn for each.
// If fn returns an error, iteration stops and the error is returned.
func IterateEntries(b *Bundle, fn func(entry BundleEntry) error) error {
	if b == nil {
		return fmt.Errorf("bundle is nil")
	}
	for _, entry := range b.Entry {
		if err := fn(entry); err != nil {
			return err
		}
	}
	return nil
}
