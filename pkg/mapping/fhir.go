package mapping

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sroopra/ghega/pkg/fhir"
)

// FHIRMapping configures which segments should be mapped to which FHIR resources.
type FHIRMapping struct {
	ResourceType string // "Patient", "Encounter", etc.
	Segment      string // "PID", "PV1", etc.
}

// FHIREngine transforms HL7v2 messages into FHIR Bundle resources.
type FHIREngine struct {
	mappings []FHIRMapping
}

// NewFHIREngine creates a new FHIR mapping engine.
func NewFHIREngine(mappings []FHIRMapping) *FHIREngine {
	return &FHIREngine{mappings: mappings}
}

// Apply parses the raw HL7v2 message and returns a FHIR Bundle containing mapped resources.
func (e *FHIREngine) Apply(rawHL7 []byte) (*fhir.Bundle, error) {
	msg, err := parseHL7(rawHL7)
	if err != nil {
		return nil, fmt.Errorf("parse hl7: %w", err)
	}

	bundle := &fhir.Bundle{
		ResourceType: "Bundle",
		Type:         "collection",
	}

	enabled := e.enabledMappings()

	if enabled["Patient"] {
		bundle.Entry = append(bundle.Entry, fhir.BundleEntry{Resource: mapPID(msg)})
	}
	if enabled["Encounter"] {
		bundle.Entry = append(bundle.Entry, fhir.BundleEntry{Resource: mapPV1(msg)})
	}
	if enabled["MessageHeader"] {
		bundle.Entry = append(bundle.Entry, fhir.BundleEntry{Resource: mapMSH(msg)})
	}
	if enabled["Observation"] {
		for _, obs := range mapOBX(msg) {
			bundle.Entry = append(bundle.Entry, fhir.BundleEntry{Resource: obs})
		}
	}
	if enabled["DiagnosticReport"] {
		for _, dr := range mapOBR(msg) {
			bundle.Entry = append(bundle.Entry, fhir.BundleEntry{Resource: dr})
		}
	}

	return bundle, nil
}

func (e *FHIREngine) enabledMappings() map[string]bool {
	if len(e.mappings) == 0 {
		return map[string]bool{
			"Patient":          true,
			"Encounter":        true,
			"Observation":      true,
			"DiagnosticReport": true,
			"MessageHeader":    true,
		}
	}
	m := make(map[string]bool, len(e.mappings))
	for _, mapping := range e.mappings {
		m[mapping.ResourceType] = true
	}
	return m
}

// mapPID maps PID segments to a Patient resource.
func mapPID(msg *hl7Message) *fhir.Patient {
	patient := &fhir.Patient{ResourceType: "Patient"}

	if id, _ := msg.getValue("PID-3.1"); id != "" {
		patient.Identifier = []fhir.Identifier{{Value: id}}
	}

	if name, _ := msg.getValue("PID-5"); name != "" {
		parts := strings.Split(name, "^")
		hn := fhir.HumanName{}
		if len(parts) > 0 {
			hn.Family = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			hn.Given = append(hn.Given, parts[1])
		}
		if len(parts) > 2 && parts[2] != "" {
			hn.Given = append(hn.Given, parts[2])
		}
		patient.Name = []fhir.HumanName{hn}
	}

	if bd, _ := msg.getValue("PID-7"); bd != "" {
		patient.BirthDate = formatHL7Date(bd)
	}

	if gender, _ := msg.getValue("PID-8"); gender != "" {
		patient.Gender = translateGender(gender)
	}

	if addr, _ := msg.getValue("PID-11"); addr != "" {
		parts := strings.Split(addr, "^")
		a := fhir.Address{}
		if len(parts) > 0 && parts[0] != "" {
			a.Line = append(a.Line, parts[0])
		}
		if len(parts) > 1 && parts[1] != "" {
			a.Line = append(a.Line, parts[1])
		}
		if len(parts) > 2 {
			a.City = parts[2]
		}
		if len(parts) > 3 {
			a.State = parts[3]
		}
		if len(parts) > 4 {
			a.PostalCode = parts[4]
		}
		patient.Address = []fhir.Address{a}
	}

	if phone, _ := msg.getValue("PID-13"); phone != "" {
		patient.Telecom = []fhir.ContactPoint{{System: "phone", Value: phone}}
	}

	return patient
}

// mapPV1 maps PV1 segments to an Encounter resource.
func mapPV1(msg *hl7Message) *fhir.Encounter {
	enc := &fhir.Encounter{ResourceType: "Encounter"}

	if cls, _ := msg.getValue("PV1-2"); cls != "" {
		enc.Class = translateEncounterClass(cls)
	}

	if vn, _ := msg.getValue("PV1-19"); vn != "" {
		enc.Identifier = []fhir.Identifier{{Value: vn}}
	}

	if start, _ := msg.getValue("PV1-44"); start != "" {
		if enc.Period == nil {
			enc.Period = &fhir.Period{}
		}
		enc.Period.Start = formatHL7DateTime(start)
	}

	if end, _ := msg.getValue("PV1-45"); end != "" {
		if enc.Period == nil {
			enc.Period = &fhir.Period{}
		}
		enc.Period.End = formatHL7DateTime(end)
	}

	return enc
}

// mapOBX maps all OBX segments to Observation resources.
func mapOBX(msg *hl7Message) []*fhir.Observation {
	var observations []*fhir.Observation

	for _, seg := range msg.segments {
		if seg.name != "OBX" {
			continue
		}

		obs := &fhir.Observation{ResourceType: "Observation"}

		valType, _ := seg.getField(2, 0)

		if code, _ := seg.getField(3, 0); code != "" {
			obs.Code = parseCodeableConcept(code)
		}

		if val, _ := seg.getField(5, 0); val != "" {
			switch valType {
			case "NM":
				obs.ValueQuantity = parseQuantity(val, "")
			case "CE":
				obs.ValueCodeableConcept = parseCodeableConcept(val)
			default:
				obs.ValueString = val
			}
		}

		if unit, _ := seg.getField(6, 0); unit != "" {
			if obs.ValueQuantity != nil {
				obs.ValueQuantity.Unit = unit
			}
		}

		if status, _ := seg.getField(11, 0); status != "" {
			obs.Status = translateObservationStatus(status)
		}

		if dt, _ := seg.getField(14, 0); dt != "" {
			obs.EffectiveDateTime = formatHL7DateTime(dt)
		}

		observations = append(observations, obs)
	}

	return observations
}

// mapOBR maps all OBR segments to DiagnosticReport resources.
func mapOBR(msg *hl7Message) []*fhir.DiagnosticReport {
	var reports []*fhir.DiagnosticReport

	for _, seg := range msg.segments {
		if seg.name != "OBR" {
			continue
		}

		dr := &fhir.DiagnosticReport{ResourceType: "DiagnosticReport"}

		if setID, _ := seg.getField(1, 0); setID != "" {
			dr.Identifier = append(dr.Identifier, fhir.Identifier{Value: setID})
		}

		if placer, _ := seg.getField(2, 0); placer != "" {
			dr.Identifier = append(dr.Identifier, fhir.Identifier{Value: placer})
		}

		if filler, _ := seg.getField(3, 0); filler != "" {
			dr.Identifier = append(dr.Identifier, fhir.Identifier{Value: filler})
		}

		if code, _ := seg.getField(4, 0); code != "" {
			dr.Code = parseCodeableConcept(code)
		}

		if dt, _ := seg.getField(7, 0); dt != "" {
			dr.EffectiveDateTime = formatHL7DateTime(dt)
		}

		if issued, _ := seg.getField(22, 0); issued != "" {
			dr.Issued = formatHL7DateTime(issued)
		}

		if status, _ := seg.getField(25, 0); status != "" {
			dr.Status = translateDiagnosticReportStatus(status)
		}

		reports = append(reports, dr)
	}

	return reports
}

// mapMSH maps MSH segments to a MessageHeader resource.
func mapMSH(msg *hl7Message) *fhir.MessageHeader {
	mh := &fhir.MessageHeader{ResourceType: "MessageHeader"}

	if name, _ := msg.getValue("MSH-3"); name != "" {
		if mh.Source == nil {
			mh.Source = &fhir.MessageHeaderSource{}
		}
		mh.Source.Name = name
	}

	if endpoint, _ := msg.getValue("MSH-4"); endpoint != "" {
		if mh.Source == nil {
			mh.Source = &fhir.MessageHeaderSource{}
		}
		mh.Source.Endpoint = endpoint
	}

	if name, _ := msg.getValue("MSH-5"); name != "" {
		mh.Destination = append(mh.Destination, fhir.MessageHeaderDestination{Name: name})
	}

	if endpoint, _ := msg.getValue("MSH-6"); endpoint != "" {
		if len(mh.Destination) > 0 {
			mh.Destination[0].Endpoint = endpoint
		} else {
			mh.Destination = append(mh.Destination, fhir.MessageHeaderDestination{Endpoint: endpoint})
		}
	}

	if id, _ := msg.getValue("MSH-10"); id != "" {
		mh.ID = id
	}

	if msgType, _ := msg.getValue("MSH-9.1"); msgType != "" {
		trigger, _ := msg.getValue("MSH-9.2")
		mh.EventCoding = &fhir.Coding{
			System: "http://hl7.org/fhir/message-events",
			Code:   msgType + "^" + trigger,
		}
	}

	return mh
}

// formatHL7Date converts YYYYMMDD to YYYY-MM-DD.
func formatHL7Date(s string) string {
	if len(s) >= 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}

// formatHL7DateTime converts YYYYMMDDHHMMSS to FHIR dateTime.
func formatHL7DateTime(s string) string {
	if len(s) >= 8 {
		result := s[:4] + "-" + s[4:6] + "-" + s[6:8]
		if len(s) >= 12 {
			result += "T" + s[8:10] + ":" + s[10:12]
			if len(s) >= 14 {
				result += ":" + s[12:14]
			}
			result += "Z"
		}
		return result
	}
	return s
}

func translateGender(g string) string {
	switch g {
	case "M":
		return "male"
	case "F":
		return "female"
	case "U":
		return "unknown"
	case "O":
		return "other"
	default:
		return strings.ToLower(g)
	}
}

func translateEncounterClass(cls string) *fhir.Coding {
	switch cls {
	case "I":
		return &fhir.Coding{System: "http://terminology.hl7.org/CodeSystem/v3-ActCode", Code: "IMP", Display: "inpatient"}
	case "O":
		return &fhir.Coding{System: "http://terminology.hl7.org/CodeSystem/v3-ActCode", Code: "AMB", Display: "outpatient"}
	case "E":
		return &fhir.Coding{System: "http://terminology.hl7.org/CodeSystem/v3-ActCode", Code: "EMER", Display: "emergency"}
	default:
		return &fhir.Coding{Code: strings.ToLower(cls)}
	}
}

func translateObservationStatus(s string) string {
	switch s {
	case "F":
		return "final"
	case "P":
		return "preliminary"
	case "C":
		return "corrected"
	default:
		return strings.ToLower(s)
	}
}

func translateDiagnosticReportStatus(s string) string {
	switch s {
	case "F":
		return "final"
	case "P":
		return "preliminary"
	case "C":
		return "corrected"
	default:
		return strings.ToLower(s)
	}
}

func parseCodeableConcept(s string) *fhir.CodeableConcept {
	parts := strings.Split(s, "^")
	cc := &fhir.CodeableConcept{}
	if len(parts) >= 2 {
		cc.Coding = []fhir.Coding{{
			Code:    parts[0],
			Display: parts[1],
		}}
		cc.Text = parts[1]
	} else {
		cc.Text = s
	}
	return cc
}

func parseQuantity(val, unit string) *fhir.Quantity {
	q := &fhir.Quantity{Unit: unit}
	if v, err := strconv.ParseFloat(val, 64); err == nil {
		q.Value = v
	}
	return q
}
