// Package fhir provides minimal FHIR R4 types used by Ghega connectors.
// These are intentionally lightweight to avoid heavy external dependencies.
package fhir

import "encoding/json"

// Bundle is a FHIR Bundle resource.
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type"`
	Entry        []BundleEntry `json:"entry"`
}

// BundleEntry represents a single entry in a Bundle.
type BundleEntry struct {
	FullURL  string          `json:"fullUrl,omitempty"`
	Resource json.RawMessage `json:"resource"`
}

// OperationOutcome is returned when FHIR processing fails.
type OperationOutcome struct {
	ResourceType string          `json:"resourceType"`
	Issue        []OutcomeIssue `json:"issue"`
}

// OutcomeIssue represents a single issue inside an OperationOutcome.
type OutcomeIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Diagnostics string `json:"diagnostics,omitempty"`
}

// GenericResource captures the resourceType and id of any FHIR resource.
type GenericResource struct {
	ResourceType string `json:"resourceType"`
	ID           string `json:"id,omitempty"`
}
