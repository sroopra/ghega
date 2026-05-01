package mapping

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sroopra/ghega/pkg/fhir"
)

// FHIREngine converts HL7v2 messages to FHIR R4 Bundle resources.
type FHIREngine struct{}

// NewFHIREngine creates a new FHIR mapping engine.
func NewFHIREngine() *FHIREngine {
	return &FHIREngine{}
}

// Apply parses an HL7v2 message and maps segments to FHIR resources.
func (e *FHIREngine) Apply(rawHL7 []byte) (*fhir.Bundle, error) {
	msg, err := parseHL7(rawHL7)
	if err != nil {
		return nil, fmt.Errorf("parse hl7: %w", err)
	}

	bundle := &fhir.Bundle{
		ResourceType: "Bundle",
		Type:         "collection",
	}

	if patient := mapPID(msg); patient != nil {
		bundle.Entry = append(bundle.Entry, bundleEntry(patient))
	}
	if encounter := mapPV1(msg); encounter != nil {
		bundle.Entry = append(bundle.Entry, bundleEntry(encounter))
	}
	if obs := mapOBX(msg); obs != nil {
		bundle.Entry = append(bundle.Entry, bundleEntry(obs))
	}
	if report := mapOBR(msg); report != nil {
		bundle.Entry = append(bundle.Entry, bundleEntry(report))
	}
	if header := mapMSH(msg); header != nil {
		bundle.Entry = append(bundle.Entry, bundleEntry(header))
	}

	return bundle, nil
}

func bundleEntry(resource any) fhir.BundleEntry {
	b, _ := json.Marshal(resource)
	return fhir.BundleEntry{Resource: b}
}

// ---------------------------------------------------------------------------
// PID -> Patient
// ---------------------------------------------------------------------------

func mapPID(msg *hl7Message) *fhir.Patient {
	patient := &fhir.Patient{ResourceType: "Patient"}

	if mrn, _ := msg.getValue("PID-3.1"); mrn != "" {
		patient.Identifier = append(patient.Identifier, fhir.Identifier{
			Value: mrn,
		})
	}

	family, _ := msg.getValue("PID-5.1")
	given, _ := msg.getValue("PID-5.2")
	if family != "" || given != "" {
		patient.Name = append(patient.Name, fhir.HumanName{
			Family: family,
			Given:  []string{given},
		})
	}

	if bd, _ := msg.getValue("PID-7"); bd != "" {
		patient.BirthDate = formatDate(bd)
	}

	if gender, _ := msg.getValue("PID-8"); gender != "" {
		patient.Gender = translateGender(gender)
	}

	if addrLine, _ := msg.getValue("PID-11.1"); addrLine != "" {
		city, _ := msg.getValue("PID-11.3")
		state, _ := msg.getValue("PID-11.4")
		zip, _ := msg.getValue("PID-11.5")
		patient.Address = append(patient.Address, fhir.Address{
			Line:       []string{addrLine},
			City:       city,
			State:      state,
			PostalCode: zip,
		})
	}

	if phone, _ := msg.getValue("PID-13.1"); phone != "" {
		patient.Telecom = append(patient.Telecom, fhir.ContactPoint{
			System: "phone",
			Value:  phone,
		})
	}

	return patient
}

func translateGender(g string) string {
	switch strings.ToUpper(g) {
	case "M":
		return "male"
	case "F":
		return "female"
	case "U":
		return "unknown"
	case "O":
		return "other"
	default:
		return "unknown"
	}
}

func formatDate(s string) string {
	if len(s) == 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:]
	}
	return s
}

// ---------------------------------------------------------------------------
// PV1 -> Encounter
// ---------------------------------------------------------------------------

func mapPV1(msg *hl7Message) *fhir.Encounter {
	enc := &fhir.Encounter{ResourceType: "Encounter", Status: "in-progress"}

	if classCode, _ := msg.getValue("PV1-2"); classCode != "" {
		enc.Class = &fhir.Coding{Code: classCode}
	}

	if loc, _ := msg.getValue("PV1-3.1"); loc != "" {
		enc.Location = append(enc.Location, fhir.EncounterLocation{
			Location: &fhir.Reference{Reference: "Location/" + loc},
			Status:   "active",
		})
	}

	if participant, _ := msg.getValue("PV1-7.1"); participant != "" {
		enc.Participant = append(enc.Participant, fhir.EncounterParticipant{
			Individual: &fhir.Reference{Reference: "Practitioner/" + participant},
		})
	}

	if vn, _ := msg.getValue("PV1-19"); vn != "" {
		enc.Identifier = append(enc.Identifier, fhir.Identifier{Value: vn})
	}

	if start, _ := msg.getValue("PV1-44"); start != "" {
		enc.Period = &fhir.Period{Start: formatDateTime(start)}
	}
	if end, _ := msg.getValue("PV1-45"); end != "" {
		if enc.Period == nil {
			enc.Period = &fhir.Period{}
		}
		enc.Period.End = formatDateTime(end)
	}

	return enc
}

func formatDateTime(s string) string {
	if len(s) >= 14 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:8] + "T" + s[8:10] + ":" + s[10:12] + ":" + s[12:14]
	}
	if len(s) == 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:]
	}
	return s
}

// ---------------------------------------------------------------------------
// OBX -> Observation
// ---------------------------------------------------------------------------

func mapOBX(msg *hl7Message) *fhir.Observation {
	obs := &fhir.Observation{ResourceType: "Observation", Status: "final"}

	valType, _ := msg.getValue("OBX-2")
	code, _ := msg.getValue("OBX-3.1")
	value, _ := msg.getValue("OBX-5")
	unit, _ := msg.getValue("OBX-6")
	interp, _ := msg.getValue("OBX-8")
	status, _ := msg.getValue("OBX-11")
	effDT, _ := msg.getValue("OBX-14")

	if code != "" {
		obs.Code = &fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: code}},
		}
	}

	if value != "" {
		switch strings.ToUpper(valType) {
		case "NM":
			if v, err := parseFloat(value); err == nil {
				obs.ValueQuantity = &fhir.Quantity{Value: &v, Unit: unit}
			} else {
				obs.ValueString = value
			}
		default:
			obs.ValueString = value
		}
	}

	if interp != "" {
		obs.Interpretation = append(obs.Interpretation, fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: interp}},
		})
	}

	if status != "" {
		obs.Status = strings.ToLower(status)
	}

	if effDT != "" {
		obs.EffectiveDateTime = formatDateTime(effDT)
	}

	return obs
}

// ---------------------------------------------------------------------------
// OBR -> DiagnosticReport
// ---------------------------------------------------------------------------

func mapOBR(msg *hl7Message) *fhir.DiagnosticReport {
	report := &fhir.DiagnosticReport{ResourceType: "DiagnosticReport"}

	if id, _ := msg.getValue("OBR-1"); id != "" {
		report.Identifier = append(report.Identifier, fhir.Identifier{Value: id})
	}

	if code, _ := msg.getValue("OBR-4.1"); code != "" {
		report.Code = &fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: code}},
		}
	}

	if eff, _ := msg.getValue("OBR-7"); eff != "" {
		report.EffectiveDateTime = formatDateTime(eff)
	}

	if issued, _ := msg.getValue("OBR-22"); issued != "" {
		report.Issued = formatDateTime(issued)
	}

	if status, _ := msg.getValue("OBR-25"); status != "" {
		report.Status = strings.ToLower(status)
	}

	return report
}

// ---------------------------------------------------------------------------
// MSH -> MessageHeader
// ---------------------------------------------------------------------------

func mapMSH(msg *hl7Message) *fhir.MessageHeader {
	header := &fhir.MessageHeader{ResourceType: "MessageHeader"}

	if name, _ := msg.getValue("MSH-3"); name != "" {
		header.Source = &fhir.MessageHeaderSource{
			Name:     name,
			Endpoint: mustGet(msg, "MSH-4"),
		}
	}

	if destName, _ := msg.getValue("MSH-5"); destName != "" {
		header.Destination = append(header.Destination, fhir.MessageHeaderDestination{
			Name:     destName,
			Endpoint: mustGet(msg, "MSH-6"),
		})
	}

	if ts, _ := msg.getValue("MSH-7"); ts != "" {
		header.Timestamp = formatDateTime(ts)
	}

	if msgType, _ := msg.getValue("MSH-9.1"); msgType != "" {
		trigger, _ := msg.getValue("MSH-9.2")
		header.EventCoding = &fhir.Coding{
			Code:    msgType + "^" + trigger,
			System:  "http://hl7.org/fhir/message-events",
			Display: msgType + " " + trigger,
		}
	}

	return header
}

func mustGet(msg *hl7Message, path string) string {
	v, _ := msg.getValue(path)
	return v
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
