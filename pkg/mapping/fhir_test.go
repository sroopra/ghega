package mapping

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/fhir"
)

func buildSegment(name string, fields ...string) string {
	parts := append([]string{name}, fields...)
	return strings.Join(parts, "|")
}

func TestFHIREngineApply_ADT_A01(t *testing.T) {
	// Construct PV1 with exactly 44 data fields so PV1-44 is the admit datetime.
	pv1Fields := make([]string, 44)
	pv1Fields[0] = "1"
	pv1Fields[1] = "I"
	pv1Fields[2] = "GHEGA_WARD^GHEGA_ROOM^1"
	pv1Fields[43] = "20240101000000"
	pv1 := buildSegment("PV1", pv1Fields...)

	// Synthetic ADT^A01 message with no PHI.
	raw := []byte(
		"MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
			"EVN|A01|20240101120000|\r" +
			"PID|1||SYNTHETIC_MRN_001^^^GHEGA_FACILITY^MR||TESTPATIENT^SYNTHETIC||19800101|M|||123 SYNTHETIC STREET^^SYNTHETIC CITY^ST^12345||555-0100||||||||||||||||||||\r" +
			pv1 + "\r",
	)

	eng := NewFHIREngine(nil)
	bundle, err := eng.Apply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("bundle.ResourceType = %q, want %q", bundle.ResourceType, "Bundle")
	}
	if bundle.Type != "collection" {
		t.Errorf("bundle.Type = %q, want %q", bundle.Type, "collection")
	}

	// Should contain Patient, Encounter, and MessageHeader.
	if len(bundle.Entry) < 3 {
		t.Fatalf("expected at least 3 entries, got %d", len(bundle.Entry))
	}

	var patient *fhir.Patient
	var encounter *fhir.Encounter
	var msgHeader *fhir.MessageHeader
	for _, entry := range bundle.Entry {
		switch r := entry.Resource.(type) {
		case *fhir.Patient:
			patient = r
		case *fhir.Encounter:
			encounter = r
		case *fhir.MessageHeader:
			msgHeader = r
		}
	}

	if patient == nil {
		t.Fatal("expected Patient resource in bundle")
	}
	if len(patient.Identifier) != 1 || patient.Identifier[0].Value != "SYNTHETIC_MRN_001" {
		t.Errorf("patient.Identifier = %v, want [{Value: SYNTHETIC_MRN_001}]", patient.Identifier)
	}
	if len(patient.Name) != 1 || patient.Name[0].Family != "TESTPATIENT" {
		t.Errorf("patient.Name[0].Family = %q, want %q", patient.Name[0].Family, "TESTPATIENT")
	}
	if len(patient.Name[0].Given) != 1 || patient.Name[0].Given[0] != "SYNTHETIC" {
		t.Errorf("patient.Name[0].Given = %v, want [SYNTHETIC]", patient.Name[0].Given)
	}
	if patient.Gender != "male" {
		t.Errorf("patient.Gender = %q, want %q", patient.Gender, "male")
	}
	if patient.BirthDate != "1980-01-01" {
		t.Errorf("patient.BirthDate = %q, want %q", patient.BirthDate, "1980-01-01")
	}

	if encounter == nil {
		t.Fatal("expected Encounter resource in bundle")
	}
	if encounter.Class == nil || encounter.Class.Code != "IMP" {
		t.Errorf("encounter.Class = %v, want IMP coding", encounter.Class)
	}
	if encounter.Period == nil || encounter.Period.Start != "2024-01-01T00:00:00Z" {
		t.Errorf("encounter.Period.Start = %q, want %q", encounter.Period.Start, "2024-01-01T00:00:00Z")
	}

	if msgHeader == nil {
		t.Fatal("expected MessageHeader resource in bundle")
	}
	if msgHeader.ID != "MSG001" {
		t.Errorf("msgHeader.ID = %q, want %q", msgHeader.ID, "MSG001")
	}
	if msgHeader.EventCoding == nil || msgHeader.EventCoding.Code != "ADT^A01" {
		t.Errorf("msgHeader.EventCoding = %v, want code ADT^A01", msgHeader.EventCoding)
	}
	if msgHeader.Source == nil || msgHeader.Source.Name != "GHEGA_SENDER" {
		t.Errorf("msgHeader.Source.Name = %q, want %q", msgHeader.Source.Name, "GHEGA_SENDER")
	}
	if len(msgHeader.Destination) != 1 || msgHeader.Destination[0].Name != "GHEGA_RECEIVER" {
		t.Errorf("msgHeader.Destination = %v, want [{Name: GHEGA_RECEIVER}]", msgHeader.Destination)
	}
}

func TestFHIREngineApply_ORU_R01(t *testing.T) {
	// Construct OBR with exactly 25 data fields so OBR-22 and OBR-25 are correct.
	obrFields := make([]string, 25)
	obrFields[0] = "1"
	obrFields[1] = "PLACER001"
	obrFields[2] = "FILLER001"
	obrFields[3] = "24323-8^Complete Blood Count"
	obrFields[6] = "20240101100000"
	obrFields[21] = "20240101103000"
	obrFields[24] = "F"
	obr := buildSegment("OBR", obrFields...)

	// Synthetic ORU^R01 message with no PHI.
	raw := []byte(
		"MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ORU^R01|MSG002|P|2.5\r" +
			"PID|1||SYNTHETIC_MRN_002^^^GHEGA_FACILITY^MR||TESTPATIENT^SECOND||19850101|F|||456 SYNTHETIC AVE^^SYNTHETIC TOWN^ST^67890||||||||||||||||||||\r" +
			obr + "\r" +
			"OBX|1|NM|24323-8^WBC|1|7.2|10*3/uL|||||F|||20240101100000\r" +
			"OBX|2|ST|24323-8^RBC|2|4.5|10*6/uL|||||F|||20240101100000\r",
	)

	eng := NewFHIREngine(nil)
	bundle, err := eng.Apply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var patient *fhir.Patient
	var diagnosticReport *fhir.DiagnosticReport
	var observations []*fhir.Observation
	for _, entry := range bundle.Entry {
		switch r := entry.Resource.(type) {
		case *fhir.Patient:
			patient = r
		case *fhir.DiagnosticReport:
			diagnosticReport = r
		case *fhir.Observation:
			observations = append(observations, r)
		}
	}

	if patient == nil {
		t.Fatal("expected Patient resource in bundle")
	}
	if patient.Gender != "female" {
		t.Errorf("patient.Gender = %q, want %q", patient.Gender, "female")
	}

	if diagnosticReport == nil {
		t.Fatal("expected DiagnosticReport resource in bundle")
	}
	if diagnosticReport.Status != "final" {
		t.Errorf("diagnosticReport.Status = %q, want %q", diagnosticReport.Status, "final")
	}
	if diagnosticReport.Code == nil || diagnosticReport.Code.Text != "Complete Blood Count" {
		t.Errorf("diagnosticReport.Code = %v, want Complete Blood Count", diagnosticReport.Code)
	}

	if len(observations) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(observations))
	}

	// First OBX has NM value type → valueQuantity.
	obs1 := observations[0]
	if obs1.ValueQuantity == nil {
		t.Errorf("obs1.ValueQuantity = nil, want non-nil")
	} else {
		if obs1.ValueQuantity.Value != 7.2 {
			t.Errorf("obs1.ValueQuantity.Value = %v, want 7.2", obs1.ValueQuantity.Value)
		}
		if obs1.ValueQuantity.Unit != "10*3/uL" {
			t.Errorf("obs1.ValueQuantity.Unit = %q, want %q", obs1.ValueQuantity.Unit, "10*3/uL")
		}
	}
	if obs1.Code == nil || obs1.Code.Text != "WBC" {
		t.Errorf("obs1.Code = %v, want WBC", obs1.Code)
	}

	// Second OBX has ST value type → valueString.
	obs2 := observations[1]
	if obs2.ValueString != "4.5" {
		t.Errorf("obs2.ValueString = %q, want %q", obs2.ValueString, "4.5")
	}
	if obs2.Code == nil || obs2.Code.Text != "RBC" {
		t.Errorf("obs2.Code = %v, want RBC", obs2.Code)
	}
}

func TestFHIREngineApply_MissingSegments(t *testing.T) {
	// Minimal message with only MSH and EVN — no PID, PV1, OBX, OBR.
	raw := []byte(
		"MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG003|P|2.5\r" +
			"EVN|A01|20240101120000|\r",
	)

	eng := NewFHIREngine(nil)
	bundle, err := eng.Apply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still contain Patient, Encounter, MessageHeader (even if empty).
	var hasPatient, hasEncounter, hasMessageHeader bool
	for _, entry := range bundle.Entry {
		switch r := entry.Resource.(type) {
		case *fhir.Patient:
			hasPatient = true
			if r.ResourceType != "Patient" {
				t.Errorf("patient.ResourceType = %q, want %q", r.ResourceType, "Patient")
			}
		case *fhir.Encounter:
			hasEncounter = true
			if r.ResourceType != "Encounter" {
				t.Errorf("encounter.ResourceType = %q, want %q", r.ResourceType, "Encounter")
			}
		case *fhir.MessageHeader:
			hasMessageHeader = true
			if r.ResourceType != "MessageHeader" {
				t.Errorf("messageHeader.ResourceType = %q, want %q", r.ResourceType, "MessageHeader")
			}
		}
	}

	if !hasPatient {
		t.Error("expected empty Patient resource even when PID segment is missing")
	}
	if !hasEncounter {
		t.Error("expected empty Encounter resource even when PV1 segment is missing")
	}
	if !hasMessageHeader {
		t.Error("expected MessageHeader resource even when only MSH is present")
	}
}

func TestFHIREngineApply_EmptyMessage(t *testing.T) {
	eng := NewFHIREngine(nil)
	_, err := eng.Apply([]byte{})
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestFHIREngineApply_MarshalToJSON(t *testing.T) {
	raw := []byte(
		"MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG004|P|2.5\r" +
			"PID|1||SYNTHETIC_MRN_004^^^GHEGA_FACILITY^MR||TESTPATIENT^FOUR||19900101|U|||789 SYNTHETIC BLVD^^SYNTHETIC CITY^ST^11111||||||||||||||||||||\r",
	)

	eng := NewFHIREngine(nil)
	bundle, err := eng.Apply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("failed to marshal bundle to JSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal bundle JSON: %v", err)
	}

	if result["resourceType"] != "Bundle" {
		t.Errorf("resourceType = %v, want Bundle", result["resourceType"])
	}
	if result["type"] != "collection" {
		t.Errorf("type = %v, want collection", result["type"])
	}

	entries, ok := result["entry"].([]any)
	if !ok || len(entries) == 0 {
		t.Fatal("expected non-empty entry array")
	}

	// Verify Patient gender "U" maps to "unknown".
	for _, e := range entries {
		entry := e.(map[string]any)
		res := entry["resource"].(map[string]any)
		if res["resourceType"] == "Patient" {
			if res["gender"] != "unknown" {
				t.Errorf("patient gender = %v, want unknown", res["gender"])
			}
		}
	}
}

func TestFHIREngineApply_WithMappings(t *testing.T) {
	// Use explicit mappings to limit resources produced.
	raw := []byte(
		"MSH|^~\\&|GHEGA_SENDER|GHEGA_FACILITY|GHEGA_RECEIVER|GHEGA_FACILITY|20240101120000||ADT^A01|MSG005|P|2.5\r" +
			"PID|1||SYNTHETIC_MRN_005^^^GHEGA_FACILITY^MR||TESTPATIENT^FIVE||19950101|M|||||||||||||||||||||||||\r" +
			"PV1|1|O||||||||||||||||||||||||||||||||||||||||||\r",
	)

	eng := NewFHIREngine([]FHIRMapping{
		{ResourceType: "Patient", Segment: "PID"},
		{ResourceType: "MessageHeader", Segment: "MSH"},
	})
	bundle, err := eng.Apply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var hasPatient, hasMessageHeader, hasEncounter bool
	for _, entry := range bundle.Entry {
		switch r := entry.Resource.(type) {
		case *fhir.Patient:
			hasPatient = true
			if r.Identifier[0].Value != "SYNTHETIC_MRN_005" {
				t.Errorf("patient identifier = %q, want %q", r.Identifier[0].Value, "SYNTHETIC_MRN_005")
			}
		case *fhir.MessageHeader:
			hasMessageHeader = true
		case *fhir.Encounter:
			hasEncounter = true
		}
	}

	if !hasPatient {
		t.Error("expected Patient resource")
	}
	if !hasMessageHeader {
		t.Error("expected MessageHeader resource")
	}
	if hasEncounter {
		t.Error("did not expect Encounter resource when mapping is not enabled")
	}
}
