// Package fhir provides minimal, lightweight FHIR R4 types for Ghega.
//
// These structs are intentionally custom and Go-native rather than importing a
// heavy external FHIR library. JSON tags match the FHIR R4 JSON serialization
// specification.
package fhir

import "encoding/json"

// ---------------------------------------------------------------------------
// Core resource types
// ---------------------------------------------------------------------------

// Bundle is a FHIR R4 Bundle resource.
type Bundle struct {
	ResourceType string        `json:"resourceType,omitempty"`
	Type         string        `json:"type,omitempty"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

// BundleEntry represents an entry in a Bundle.
type BundleEntry struct {
	Resource json.RawMessage      `json:"resource,omitempty"`
	Request  *BundleEntryRequest  `json:"request,omitempty"`
	Response *BundleEntryResponse `json:"response,omitempty"`
}

// BundleEntryRequest carries the request details for a Bundle entry.
type BundleEntryRequest struct {
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

// BundleEntryResponse carries the response details for a Bundle entry.
type BundleEntryResponse struct {
	Status   string `json:"status,omitempty"`
	Location string `json:"location,omitempty"`
}

// Patient is a minimal FHIR R4 Patient resource.
type Patient struct {
	ResourceType  string           `json:"resourceType,omitempty"`
	Identifier    []Identifier     `json:"identifier,omitempty"`
	Name          []HumanName      `json:"name,omitempty"`
	BirthDate     string           `json:"birthDate,omitempty"`
	Gender        string           `json:"gender,omitempty"`
	Address       []Address        `json:"address,omitempty"`
	Telecom       []ContactPoint   `json:"telecom,omitempty"`
	MaritalStatus *CodeableConcept `json:"maritalStatus,omitempty"`
}

// Encounter is a minimal FHIR R4 Encounter resource.
type Encounter struct {
	ResourceType string                 `json:"resourceType,omitempty"`
	Identifier   []Identifier           `json:"identifier,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Class        *Coding                `json:"class,omitempty"`
	Subject      *Reference             `json:"subject,omitempty"`
	Period       *Period                `json:"period,omitempty"`
	Location     []EncounterLocation    `json:"location,omitempty"`
	Participant  []EncounterParticipant `json:"participant,omitempty"`
}

// EncounterLocation represents a location association in an Encounter.
type EncounterLocation struct {
	Location *Reference `json:"location,omitempty"`
	Status   string     `json:"status,omitempty"`
}

// EncounterParticipant represents a participant in an Encounter.
type EncounterParticipant struct {
	Type       []CodeableConcept `json:"type,omitempty"`
	Individual *Reference        `json:"individual,omitempty"`
}

// Observation is a minimal FHIR R4 Observation resource.
type Observation struct {
	ResourceType      string            `json:"resourceType,omitempty"`
	Status            string            `json:"status,omitempty"`
	Code              *CodeableConcept  `json:"code,omitempty"`
	Subject           *Reference        `json:"subject,omitempty"`
	EffectiveDateTime string            `json:"effectiveDateTime,omitempty"`
	ValueString       string            `json:"valueString,omitempty"`
	ValueQuantity     *Quantity         `json:"valueQuantity,omitempty"`
	Interpretation    []CodeableConcept `json:"interpretation,omitempty"`
}

// DiagnosticReport is a minimal FHIR R4 DiagnosticReport resource.
type DiagnosticReport struct {
	ResourceType      string           `json:"resourceType,omitempty"`
	Identifier        []Identifier     `json:"identifier,omitempty"`
	Status            string           `json:"status,omitempty"`
	Code              *CodeableConcept `json:"code,omitempty"`
	Subject           *Reference       `json:"subject,omitempty"`
	EffectiveDateTime string           `json:"effectiveDateTime,omitempty"`
	Issued            string           `json:"issued,omitempty"`
	Result            []Reference      `json:"result,omitempty"`
}

// MessageHeader is a minimal FHIR R4 MessageHeader resource.
type MessageHeader struct {
	ResourceType string                     `json:"resourceType,omitempty"`
	EventCoding  *Coding                    `json:"eventCoding,omitempty"`
	Source       *MessageHeaderSource       `json:"source,omitempty"`
	Destination  []MessageHeaderDestination `json:"destination,omitempty"`
	Timestamp    string                     `json:"timestamp,omitempty"`
}

// MessageHeaderSource represents the source of a message.
type MessageHeaderSource struct {
	Name     string        `json:"name,omitempty"`
	Endpoint string        `json:"endpoint,omitempty"`
	Contact  *ContactPoint `json:"contact,omitempty"`
}

// MessageHeaderDestination represents a destination of a message.
type MessageHeaderDestination struct {
	Name     string `json:"name,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

// OperationOutcome is a minimal FHIR R4 OperationOutcome resource.
type OperationOutcome struct {
	ResourceType string                  `json:"resourceType,omitempty"`
	Issue        []OperationOutcomeIssue `json:"issue,omitempty"`
}

// OperationOutcomeIssue represents a single issue inside an OperationOutcome.
type OperationOutcomeIssue struct {
	Severity    string           `json:"severity,omitempty"`
	Code        string           `json:"code,omitempty"`
	Diagnostics string           `json:"diagnostics,omitempty"`
	Details     *CodeableConcept `json:"details,omitempty"`
}

// ---------------------------------------------------------------------------
// Supporting types
// ---------------------------------------------------------------------------

// HumanName represents a person's name.
type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Text   string   `json:"text,omitempty"`
	Family string   `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
	Prefix []string `json:"prefix,omitempty"`
	Suffix []string `json:"suffix,omitempty"`
}

// Identifier represents an identifier.
type Identifier struct {
	System string           `json:"system,omitempty"`
	Value  string           `json:"value,omitempty"`
	Type   *CodeableConcept `json:"type,omitempty"`
}

// CodeableConcept represents a concept with codes.
type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

// Coding represents a coding in a code system.
type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// Quantity represents a measured or measurable amount.
type Quantity struct {
	Value  *float64 `json:"value,omitempty"`
	Unit   string   `json:"unit,omitempty"`
	System string   `json:"system,omitempty"`
	Code   string   `json:"code,omitempty"`
}

// Reference represents a reference to another resource.
type Reference struct {
	Reference string `json:"reference,omitempty"`
	Type      string `json:"type,omitempty"`
	Display   string `json:"display,omitempty"`
}

// Period represents a time period.
type Period struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// Address represents an address.
type Address struct {
	Use        string   `json:"use,omitempty"`
	Type       string   `json:"type,omitempty"`
	Text       string   `json:"text,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	District   string   `json:"district,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}

// ContactPoint represents a contact detail.
type ContactPoint struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
	Use    string `json:"use,omitempty"`
	Rank   *int   `json:"rank,omitempty"`
}
