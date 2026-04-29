---
name: mapping-hl7v2-to-fhir
description: >
  Use when a user wants to map or convert HL7v2 messages to FHIR resources.
  Use when someone asks about HL7 to FHIR, v2 to FHIR mapping,
  convert HL7 to FHIR, or FHIR mapping.
  Use when the topic involves translating HL7v2 segments into FHIR
  resources such as Patient, Encounter, or Observation.
license: Apache-2.0
---

# Mapping HL7v2 to FHIR

This skill guides the mapping of HL7v2 messages to FHIR resources in Ghega.
Mapping is a common step when migrating from legacy HL7v2 interfaces to
modern FHIR-based interoperability.

## When to Use

- Converting inbound HL7v2 messages to FHIR resources
- Building transformation logic for ADT, ORU, or MDM message types
- Mapping PID, PV1, OBX, and other segments to FHIR structures
- Documenting deviations from standard HL7 v2-to-FHIR guidance

## Key Concepts

- **HL7v2 Segment**: A logical group of fields (e.g., `PID`, `PV1`, `OBX`)
- **FHIR Resource**: A structured JSON or XML representation of healthcare data
- **Cardinality**: HL7v2 fields may repeat; FHIR elements have defined cardinality
- **Code Systems**: HL7v2 tables and FHIR ValueSets require translation

## Standard Mapping Guidance

The HL7 v2-to-FHIR Implementation Guide provides canonical mappings for
common message types. Ghega channels should follow this guidance as the
default, documenting any deviations explicitly.

## Common Segment Mappings

| HL7v2 Segment | FHIR Resource | Notes |
|---------------|---------------|-------|
| PID | Patient | Demographics and identifiers |
| PV1 | Encounter | Visit context and location |
| OBX | Observation | Lab results, vitals, measurements |
| OBR | DiagnosticReport or ServiceRequest | Order or report grouping |
| MSH | MessageHeader | Messaging metadata |
| EVN | Provenance or Encounter | Event context |
| NK1 | RelatedPerson | Next of kin or contact |
| AL1 | AllergyIntolerance | Allergy information |

## Deviations

When a Ghega channel deviates from the standard HL7 v2-to-FHIR guidance:

1. Document the deviation in the channel README or mapping spec
2. Provide a business justification
3. Include a mapping decision table showing the standard vs chosen approach

## Code System Translation

HL7v2 tables (e.g., `HL70001` for administrative sex) often map to FHIR
ValueSets. Use the official HL7 terminology server or a local mapping table.

| HL7v2 Table | FHIR ValueSet | Example Mapping |
|-------------|---------------|-----------------|
| HL70001 | AdministrativeGender | M -> male, F -> female, U -> unknown |
| HL70002 | MaritalStatus | S -> S, M -> M, U -> UNK |
| HL70064 | v2.0078 (Interpretation) | H -> high, L -> low, N -> normal |

## Safety

Never include real patient data (PHI) in mapping examples or test fixtures.
Always use synthetic test data. See the references for safe example patterns.

## References

- See [references/hl7v2-to-fhir-mapping.md](references/hl7v2-to-fhir-mapping.md) for mapping guidance and standards.
- See [references/common-mappings.md](references/common-mappings.md) for detailed field-level mappings.
