package mapping

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/fhir"
)

func buildADT_A01() string {
	hl7Type := "ADT^A01"
	segments := []string{
		fmt.Sprintf("MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|%s||%s|12345|P|2.5", time.Now().UTC().Format("20060102150405"), hl7Type),
		"EVN|A01|20240101000000|||",
		"PID|1||SYNTH-MRN-001^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||",
		"PV1|1|I|GHEGA_WARD^GHEGA_ROOM^1||||||||||||||||||||||||||||||||||||||||||20240101000000",
	}
	return strings.Join(segments, "\r") + "\r"
}

func buildORU_R01() string {
	hl7Type := "ORU^R01"
	segments := []string{
		fmt.Sprintf("MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|%s||%s|12346|P|2.5", time.Now().UTC().Format("20060102150405"), hl7Type),
		"PID|1||SYNTH-MRN-002^^^GHEGA_FACILITY^MR||TESTPATIENT^SECOND||19850101|F|||456 TEST AVENUE^^TESTVILLE^TS^67890||||||||||||||||||||||",
		"OBR|1|PLAC001|FILL001|24323-8^CBC|||20240101120000|||||||||||||||20240101130000|||F",
		"OBX|1|NM|24323-8^HGB|1|13.5|g/dL|12.0-16.0|N|F|||20240101120000",
	}
	return strings.Join(segments, "\r") + "\r"
}

func TestFHIREngine_ADT_A01(t *testing.T) {
	eng := NewFHIREngine()
	bundle, err := eng.Apply([]byte(buildADT_A01()))
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("ResourceType = %q, want Bundle", bundle.ResourceType)
	}
	if bundle.Type != "collection" {
		t.Errorf("Type = %q, want collection", bundle.Type)
	}

	// Should contain Patient, Encounter, MessageHeader
	if len(bundle.Entry) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(bundle.Entry))
	}

	var patient *fhir.Patient
	var encounter *fhir.Encounter
	var header *fhir.MessageHeader

	for _, entry := range bundle.Entry {
		var raw map[string]any
		if err := json.Unmarshal(entry.Resource, &raw); err != nil {
			t.Fatalf("unmarshal entry: %v", err)
		}
		rt, _ := raw["resourceType"].(string)
		switch rt {
		case "Patient":
			patient = &fhir.Patient{}
			json.Unmarshal(entry.Resource, patient)
		case "Encounter":
			encounter = &fhir.Encounter{}
			json.Unmarshal(entry.Resource, encounter)
		case "MessageHeader":
			header = &fhir.MessageHeader{}
			json.Unmarshal(entry.Resource, header)
		}
	}

	if patient == nil {
		t.Fatal("Patient resource not found")
	}
	if len(patient.Name) == 0 || patient.Name[0].Family != "TESTPATIENT" {
		t.Errorf("Patient.Name[0].Family = %q, want TESTPATIENT", patient.Name[0].Family)
	}
	if len(patient.Identifier) == 0 || patient.Identifier[0].Value != "SYNTH-MRN-001" {
		t.Errorf("Patient.Identifier[0].Value = %q, want SYNTH-MRN-001", patient.Identifier[0].Value)
	}
	if patient.Gender != "male" {
		t.Errorf("Patient.Gender = %q, want male", patient.Gender)
	}
	if patient.BirthDate != "1980-01-01" {
		t.Errorf("Patient.BirthDate = %q, want 1980-01-01", patient.BirthDate)
	}

	if encounter == nil {
		t.Fatal("Encounter resource not found")
	}
	if encounter.Class == nil || encounter.Class.Code != "I" {
		t.Errorf("Encounter.Class.Code = %q, want I", encounter.Class.Code)
	}

	if header == nil {
		t.Fatal("MessageHeader resource not found")
	}
	if header.EventCoding == nil || !strings.Contains(header.EventCoding.Code, "ADT") {
		t.Errorf("MessageHeader.EventCoding.Code = %q, want ADT*", header.EventCoding.Code)
	}
}

func TestFHIREngine_ORU_R01(t *testing.T) {
	eng := NewFHIREngine()
	bundle, err := eng.Apply([]byte(buildORU_R01()))
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(bundle.Entry) < 4 {
		t.Fatalf("expected at least 4 entries, got %d", len(bundle.Entry))
	}

	var obs *fhir.Observation
	var report *fhir.DiagnosticReport

	for _, entry := range bundle.Entry {
		var raw map[string]any
		if err := json.Unmarshal(entry.Resource, &raw); err != nil {
			t.Fatalf("unmarshal entry: %v", err)
		}
		rt, _ := raw["resourceType"].(string)
		switch rt {
		case "Observation":
			obs = &fhir.Observation{}
			json.Unmarshal(entry.Resource, obs)
		case "DiagnosticReport":
			report = &fhir.DiagnosticReport{}
			json.Unmarshal(entry.Resource, report)
		}
	}

	if obs == nil {
		t.Fatal("Observation resource not found")
	}
	if obs.ValueQuantity == nil || *obs.ValueQuantity.Value != 13.5 {
		t.Errorf("Observation.ValueQuantity.Value = %v, want 13.5", obs.ValueQuantity.Value)
	}
	if obs.Status != "final" {
		t.Errorf("Observation.Status = %q, want final", obs.Status)
	}

	if report == nil {
		t.Fatal("DiagnosticReport resource not found")
	}
	if report.Status != "f" {
		t.Errorf("DiagnosticReport.Status = %q, want f", report.Status)
	}
	if len(report.Identifier) == 0 {
		t.Fatal("DiagnosticReport.Identifier is empty, want at least one identifier from OBR-1")
	}
}

func TestFHIREngine_MissingPID(t *testing.T) {
	// Minimal HL7 with only MSH and EVN — no PID segment.
	hl7 := "MSH|^~\\&|SENDER|FACILITY|RECEIVER|FACILITY|20240101000000||ADT^A01|12345|P|2.5\rEVN|A01|20240101000000|||\r"
	eng := NewFHIREngine()
	bundle, err := eng.Apply([]byte(hl7))
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Without a PID segment, no Patient resource should be produced.
	for _, entry := range bundle.Entry {
		var raw map[string]any
		json.Unmarshal(entry.Resource, &raw)
		if raw["resourceType"] == "Patient" {
			t.Error("expected no Patient resource when PID segment is missing")
			break
		}
	}

	// Should still have a MessageHeader from the MSH segment.
	var foundHeader bool
	for _, entry := range bundle.Entry {
		var raw map[string]any
		json.Unmarshal(entry.Resource, &raw)
		if raw["resourceType"] == "MessageHeader" {
			foundHeader = true
			break
		}
	}
	if !foundHeader {
		t.Error("expected MessageHeader resource from MSH segment")
	}
}

func TestFHIREngine_RoundTrip(t *testing.T) {
	eng := NewFHIREngine()
	bundle, err := eng.Apply([]byte(buildADT_A01()))
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	data, err := fhir.BundleToJSON(bundle)
	if err != nil {
		t.Fatalf("BundleToJSON failed: %v", err)
	}

	parsed, err := fhir.ParseBundle(data)
	if err != nil {
		t.Fatalf("ParseBundle failed: %v", err)
	}

	if parsed.ResourceType != "Bundle" {
		t.Errorf("round-trip ResourceType = %q, want Bundle", parsed.ResourceType)
	}
	if len(parsed.Entry) != len(bundle.Entry) {
		t.Errorf("round-trip entries = %d, want %d", len(parsed.Entry), len(bundle.Entry))
	}
}
