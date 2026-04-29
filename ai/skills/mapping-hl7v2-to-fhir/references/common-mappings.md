# Common HL7v2-to-FHIR Mappings

This document provides field-level mapping details for the most commonly
encountered HL7v2 segments in Ghega channels.

## PID to Patient

| HL7v2 Field | FHIR Path | Notes |
|-------------|-----------|-------|
| PID-3 | Patient.identifier | May repeat; map each identifier type |
| PID-5 | Patient.name | Family name and given names |
| PID-7 | Patient.birthDate | Format as FHIR date (`YYYY-MM-DD`) |
| PID-8 | Patient.gender | Translate using HL70001 table |
| PID-11 | Patient.address | Street, city, state, postal code |
| PID-13 | Patient.telecom | Home phone, email, etc. |
| PID-16 | Patient.maritalStatus | Translate using HL70002 table |
| PID-17 | Patient.religion | Translate using HL70006 table |
| PID-22 | Patient.communication | Language preference |

### Synthetic Example

HL7v2 PID segment:

```
PID|1||SYNTH-MRN-001^^^GHEGA-TEST||TESTPATIENT^ONE||19800101|M|||123 TEST LANE^^TESTVILLE^TS^12345|||||||||||||||||||||||||||
```

Resulting FHIR Patient (abbreviated):

```json
{
  "resourceType": "Patient",
  "identifier": [{ "value": "SYNTH-MRN-001" }],
  "name": [{ "family": "TESTPATIENT", "given": ["ONE"] }],
  "birthDate": "1980-01-01",
  "gender": "male"
}
```

## PV1 to Encounter

| HL7v2 Field | FHIR Path | Notes |
|-------------|-----------|-------|
| PV1-2 | Encounter.class | Inpatient, outpatient, emergency |
| PV1-3 | Encounter.location | Assigned physical location |
| PV1-7 | Encounter.participant | Attending practitioner |
| PV1-17 | Encounter.admitSource | Admission source code |
| PV1-19 | Encounter.identifier | Visit number |
| PV1-44 | Encounter.period.start | Admission date/time |
| PV1-45 | Encounter.period.end | Discharge date/time |

## OBX to Observation

| HL7v2 Field | FHIR Path | Notes |
|-------------|-----------|-------|
| OBX-2 | Observation.value[x] type | Determines value type (NM, ST, CE, etc.) |
| OBX-3 | Observation.code | LOINC or local code |
| OBX-5 | Observation.value[x] | Actual result value |
| OBX-6 | Observation.valueQuantity.unit | Unit of measure |
| OBX-7 | Observation.referenceRange | Reference range string |
| OBX-8 | Observation.interpretation | Abnormal flag translation |
| OBX-11 | Observation.status | Final, preliminary, corrected |
| OBX-14 | Observation.effectiveDateTime | Observation timestamp |
| OBX-16 | Observation.performer | Responsible practitioner |

## OBR to DiagnosticReport

| HL7v2 Field | FHIR Path | Notes |
|-------------|-----------|-------|
| OBR-1 | DiagnosticReport.identifier | Set ID |
| OBR-2 | DiagnosticReport.identifier | Placer order number |
| OBR-3 | DiagnosticReport.identifier | Filler order number |
| OBR-4 | DiagnosticReport.code | Orderable test code |
| OBR-7 | DiagnosticReport.effectiveDateTime | Observation date/time |
| OBR-16 | DiagnosticReport.requester | Ordering provider |
| OBR-22 | DiagnosticReport.issued | Report release timestamp |
| OBR-25 | DiagnosticReport.status | Result status translation |

## MSH to MessageHeader

| HL7v2 Field | FHIR Path | Notes |
|-------------|-----------|-------|
| MSH-3 | MessageHeader.source.name | Sending application |
| MSH-4 | MessageHeader.source.endpoint | Sending facility |
| MSH-5 | MessageHeader.destination.name | Receiving application |
| MSH-6 | MessageHeader.destination.endpoint | Receiving facility |
| MSH-7 | MessageHeader.timestamp | Message creation time |
| MSH-9 | MessageHeader.eventCoding | Message type and trigger event |
| MSH-10 | MessageHeader.id | Message control ID |

## Repeating Fields

When an HL7v2 field repeats (e.g., `PID-3` with multiple identifiers):

- Map each repetition to a separate element in the FHIR array
- Preserve the identifier type code if present
- If the target FHIR element is not repeatable, use an extension or
  document the truncation behavior

## Null Values

HL7v2 `""` (two consecutive quotes) indicates an explicit null.
Map this to:

- FHIR `dataAbsentReason` extension with value `unknown`
- Or omit the element if the profile permits

## Safety Reminders

- All examples must use synthetic data
- Do not include real patient names, real MRNs, or real addresses
- Use `GHEGA-TEST-FACILITY` for facility names in examples
- Use `TESTPATIENT` for patient family names in examples
