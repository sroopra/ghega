// Package fhir provides minimal FHIR R4 types used by Ghega.
package fhir

// Bundle represents a FHIR Bundle resource.
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

// BundleEntry represents a single entry in a Bundle.
type BundleEntry struct {
	Resource interface{} `json:"resource"`
}

// Patient represents a FHIR Patient resource.
type Patient struct {
	ResourceType string          `json:"resourceType"`
	Identifier   []Identifier    `json:"identifier,omitempty"`
	Name         []HumanName     `json:"name,omitempty"`
	Gender       string          `json:"gender,omitempty"`
	BirthDate    string          `json:"birthDate,omitempty"`
	Address      []Address       `json:"address,omitempty"`
	Telecom      []ContactPoint  `json:"telecom,omitempty"`
}

// Encounter represents a FHIR Encounter resource.
type Encounter struct {
	ResourceType string       `json:"resourceType"`
	Identifier   []Identifier `json:"identifier,omitempty"`
	Class        *Coding      `json:"class,omitempty"`
	Period       *Period      `json:"period,omitempty"`
	Location     []EncounterLocation `json:"location,omitempty"`
	Participant  []EncounterParticipant `json:"participant,omitempty"`
}

// EncounterLocation represents a location association in an Encounter.
type EncounterLocation struct {
	Location *Reference `json:"location,omitempty"`
}

// EncounterParticipant represents a participant in an Encounter.
type EncounterParticipant struct {
	Individual *Reference `json:"individual,omitempty"`
}

// Observation represents a FHIR Observation resource.
type Observation struct {
	ResourceType       string          `json:"resourceType"`
	Status             string          `json:"status,omitempty"`
	Code               *CodeableConcept `json:"code,omitempty"`
	Subject            *Reference      `json:"subject,omitempty"`
	ValueString        string          `json:"valueString,omitempty"`
	ValueQuantity      *Quantity       `json:"valueQuantity,omitempty"`
	ValueCodeableConcept *CodeableConcept `json:"valueCodeableConcept,omitempty"`
	EffectiveDateTime  string          `json:"effectiveDateTime,omitempty"`
	Component          []ObservationComponent `json:"component,omitempty"`
}

// ObservationComponent represents a component of an Observation.
type ObservationComponent struct {
	Code        *CodeableConcept `json:"code,omitempty"`
	ValueString string           `json:"valueString,omitempty"`
	ValueQuantity *Quantity      `json:"valueQuantity,omitempty"`
}

// DiagnosticReport represents a FHIR DiagnosticReport resource.
type DiagnosticReport struct {
	ResourceType      string           `json:"resourceType"`
	Identifier        []Identifier     `json:"identifier,omitempty"`
	Status            string           `json:"status,omitempty"`
	Code              *CodeableConcept `json:"code,omitempty"`
	Subject           *Reference       `json:"subject,omitempty"`
	EffectiveDateTime string           `json:"effectiveDateTime,omitempty"`
	Issued            string           `json:"issued,omitempty"`
	Result            []Reference      `json:"result,omitempty"`
}

// MessageHeader represents a FHIR MessageHeader resource.
type MessageHeader struct {
	ResourceType string                     `json:"resourceType"`
	ID           string                     `json:"id,omitempty"`
	EventCoding  *Coding                    `json:"eventCoding,omitempty"`
	Source       *MessageHeaderSource       `json:"source,omitempty"`
	Destination  []MessageHeaderDestination `json:"destination,omitempty"`
}

// MessageHeaderSource represents the source of a message.
type MessageHeaderSource struct {
	Name     string `json:"name,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

// MessageHeaderDestination represents the destination of a message.
type MessageHeaderDestination struct {
	Name     string `json:"name,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

// Identifier represents a FHIR Identifier.
type Identifier struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
	Type   *CodeableConcept `json:"type,omitempty"`
}

// HumanName represents a FHIR HumanName.
type HumanName struct {
	Family string   `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
}

// Address represents a FHIR Address.
type Address struct {
	Text       string   `json:"text,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
}

// ContactPoint represents a FHIR ContactPoint.
type ContactPoint struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
}

// CodeableConcept represents a FHIR CodeableConcept.
type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

// Coding represents a FHIR Coding.
type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// Quantity represents a FHIR Quantity.
type Quantity struct {
	Value  float64 `json:"value,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	System string  `json:"system,omitempty"`
	Code   string  `json:"code,omitempty"`
}

// Period represents a FHIR Period.
type Period struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// Reference represents a FHIR Reference.
type Reference struct {
	Reference string `json:"reference,omitempty"`
	Display   string `json:"display,omitempty"`
}
