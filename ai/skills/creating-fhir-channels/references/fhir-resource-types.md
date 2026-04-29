# FHIR Resource Types Reference

This document lists commonly used FHIR resource types in Ghega channels.

## Patient Administration

| Resource | Description | Common Use Case |
|----------|-------------|-----------------|
| Patient | Demographics and identifiers | ADT registrations, master patient index |
| Encounter | Visit or appointment details | ADT admissions, discharges, transfers |
| Practitioner | Healthcare provider | Provider directory sync |
| Organization | Facility or department | Facility registry updates |
| Location | Physical place of care | Room or bed assignments |

## Clinical

| Resource | Description | Common Use Case |
|----------|-------------|-----------------|
| Observation | Measurements and findings | Lab results, vitals |
| DiagnosticReport | Grouped observations | Lab reports, imaging results |
| Condition | Problem or diagnosis | Problem list updates |
| Procedure | Performed procedures | Surgical reports |
| MedicationRequest | Prescription or order | Pharmacy orders |
| AllergyIntolerance | Allergy or intolerance | Allergy list updates |

## Infrastructure

| Resource | Description | Common Use Case |
|----------|-------------|-----------------|
| Bundle | Collection of resources | Batch uploads, search results |
| Composition | Structured document | Discharge summaries |
| MessageHeader | Messaging metadata | FHIR messaging protocol |
| OperationOutcome | Error or warning details | Validation failures |

## Synthetic Data Guidelines

When referencing resource examples:

- Use `Patient.identifier.value` values such as `TEST-001`, `SYNTH-002`
- Use `Patient.name` values such as `TESTPATIENT,ONE` or `SYNTHETIC,TWO`
- Use `Organization.name` values such as `GHEGA-TEST-FACILITY`
- Never use real names, real MRNs, or real addresses

## Version Compatibility

| Version | Base URL Convention | JSON Content-Type |
|---------|---------------------|-------------------|
| DSTU2 | `/baseDstu2` | `application/fhir+json; fhirVersion=1.0` |
| R4 | `/R4` | `application/fhir+json; fhirVersion=4.0` |

Prefer R4 for new channels unless the target system only supports DSTU2.
