package fhir

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

// syntheticBundleJSON is a synthetic FHIR Bundle with no PHI.
var syntheticBundleJSON = []byte(`{
  "resourceType": "Bundle",
  "type": "batch",
  "entry": [
    {
      "resource": {
        "resourceType": "Patient",
        "identifier": [
          {
            "system": "http://ghega.test/mrn",
            "value": "SYNTH-MRN-001"
          }
        ],
        "name": [
          {
            "family": "TESTPATIENT",
            "given": ["SYNTH"]
          }
        ],
        "birthDate": "1990-01-01",
        "gender": "unknown",
        "address": [
          {
            "city": "Testville",
            "country": "Testland"
          }
        ],
        "telecom": [
          {
            "system": "phone",
            "value": "555-TEST"
          }
        ],
        "maritalStatus": {
          "coding": [
            {
              "system": "http://terminology.hl7.org/CodeSystem/v3-MaritalStatus",
              "code": "U",
              "display": "unmarried"
            }
          ]
        }
      },
      "request": {
        "method": "POST",
        "url": "Patient"
      }
    },
    {
      "resource": {
        "resourceType": "Observation",
        "status": "final",
        "code": {
          "coding": [
            {
              "system": "http://loinc.org",
              "code": "8867-4",
              "display": "Heart rate"
            }
          ]
        },
        "subject": {
          "reference": "Patient/SYNTH-MRN-001"
        },
        "effectiveDateTime": "2024-01-01T00:00:00Z",
        "valueQuantity": {
          "value": 72,
          "unit": "beats/min"
        }
      }
    }
  ]
}`)

func TestParseBundle(t *testing.T) {
	b, err := ParseBundle(syntheticBundleJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %q", b.ResourceType)
	}
	if b.Type != "batch" {
		t.Errorf("expected type batch, got %q", b.Type)
	}
	if len(b.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(b.Entry))
	}

	// Verify first entry is a Patient.
	var patient Patient
	if err := json.Unmarshal(b.Entry[0].Resource, &patient); err != nil {
		t.Fatalf("failed to unmarshal patient: %v", err)
	}
	if patient.ResourceType != "Patient" {
		t.Errorf("expected resourceType Patient, got %q", patient.ResourceType)
	}
	if len(patient.Identifier) != 1 || patient.Identifier[0].Value != "SYNTH-MRN-001" {
		t.Errorf("unexpected identifier: %+v", patient.Identifier)
	}
	if len(patient.Name) != 1 || patient.Name[0].Family != "TESTPATIENT" {
		t.Errorf("unexpected name: %+v", patient.Name)
	}
	if patient.Gender != "unknown" {
		t.Errorf("expected gender unknown, got %q", patient.Gender)
	}
	if len(patient.Address) != 1 || patient.Address[0].City != "Testville" {
		t.Errorf("unexpected address: %+v", patient.Address)
	}
	if len(patient.Telecom) != 1 || patient.Telecom[0].Value != "555-TEST" {
		t.Errorf("unexpected telecom: %+v", patient.Telecom)
	}
	if patient.MaritalStatus == nil || len(patient.MaritalStatus.Coding) != 1 {
		t.Errorf("unexpected maritalStatus: %+v", patient.MaritalStatus)
	}

	// Verify request on first entry.
	if b.Entry[0].Request == nil {
		t.Fatal("expected request on first entry")
	}
	if b.Entry[0].Request.Method != "POST" {
		t.Errorf("expected method POST, got %q", b.Entry[0].Request.Method)
	}
	if b.Entry[0].Request.URL != "Patient" {
		t.Errorf("expected url Patient, got %q", b.Entry[0].Request.URL)
	}

	// Verify second entry is an Observation.
	var obs Observation
	if err := json.Unmarshal(b.Entry[1].Resource, &obs); err != nil {
		t.Fatalf("failed to unmarshal observation: %v", err)
	}
	if obs.ResourceType != "Observation" {
		t.Errorf("expected resourceType Observation, got %q", obs.ResourceType)
	}
	if obs.Status != "final" {
		t.Errorf("expected status final, got %q", obs.Status)
	}
	if obs.ValueQuantity == nil || obs.ValueQuantity.Value == nil || *obs.ValueQuantity.Value != 72 {
		t.Errorf("unexpected valueQuantity: %+v", obs.ValueQuantity)
	}
}

func TestValidateBundleType(t *testing.T) {
	validTypes := []string{"document", "message", "transaction", "transaction-response", "batch", "batch-response", "history", "searchset", "collection"}
	for _, typ := range validTypes {
		b := &Bundle{Type: typ}
		if err := ValidateBundleType(b); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", typ, err)
		}
	}

	invalidTypes := []string{"", "search-set", "invalid"}
	for _, typ := range invalidTypes {
		b := &Bundle{Type: typ}
		if err := ValidateBundleType(b); err == nil {
			t.Errorf("expected %q to be invalid", typ)
		}
	}

	if err := ValidateBundleType(nil); err == nil {
		t.Error("expected error for nil bundle")
	}
}

func TestIterateEntries(t *testing.T) {
	b, err := ParseBundle(syntheticBundleJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	err = IterateEntries(b, func(entry BundleEntry) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 entries processed, got %d", count)
	}

	expectedErr := errors.New("stop iteration")
	count = 0
	err = IterateEntries(b, func(entry BundleEntry) error {
		count++
		if count == 1 {
			return expectedErr
		}
		return nil
	})
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if count != 1 {
		t.Errorf("expected 1 entry processed before error, got %d", count)
	}

	if err := IterateEntries(nil, func(entry BundleEntry) error { return nil }); err == nil {
		t.Error("expected error for nil bundle")
	}
}

func TestRoundTrip(t *testing.T) {
	original := &Bundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entry: []BundleEntry{
			{
				Resource: json.RawMessage(`{"resourceType":"MessageHeader","eventCoding":{"system":"http://ghega.test/events","code":"adt-a01"},"source":{"name":"GHEGA-TEST-FACILITY"},"timestamp":"2024-01-01T00:00:00Z"}`),
				Request: &BundleEntryRequest{
					Method: "POST",
					URL:    "MessageHeader",
				},
			},
		},
	}

	data, err := BundleToJSON(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := ParseBundle(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.ResourceType != original.ResourceType {
		t.Errorf("resourceType mismatch: %q vs %q", parsed.ResourceType, original.ResourceType)
	}
	if parsed.Type != original.Type {
		t.Errorf("type mismatch: %q vs %q", parsed.Type, original.Type)
	}
	if len(parsed.Entry) != len(original.Entry) {
		t.Fatalf("entry count mismatch: %d vs %d", len(parsed.Entry), len(original.Entry))
	}

	if !bytes.Equal(parsed.Entry[0].Resource, original.Entry[0].Resource) {
		t.Errorf("resource mismatch: %s vs %s", parsed.Entry[0].Resource, original.Entry[0].Resource)
	}
	if parsed.Entry[0].Request.Method != original.Entry[0].Request.Method {
		t.Errorf("request method mismatch")
	}
	if parsed.Entry[0].Request.URL != original.Entry[0].Request.URL {
		t.Errorf("request URL mismatch")
	}
}

func TestBundleToJSON_NilBundle(t *testing.T) {
	_, err := BundleToJSON(nil)
	if err == nil {
		t.Fatal("expected error for nil bundle")
	}
}
