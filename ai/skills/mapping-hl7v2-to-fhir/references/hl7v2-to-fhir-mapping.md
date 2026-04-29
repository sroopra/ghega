# HL7 v2-to-FHIR Mapping Guidance

This document provides guidance for mapping HL7v2 messages to FHIR resources
in Ghega channels, including references to official standards and recommended
practices.

## Official Standards

The primary reference for HL7 v2-to-FHIR mapping is the HL7 FHIR
Implementation Guide for v2-to-FHIR conversion. This guide defines canonical
mappings for common message types including ADT, ORU, MDM, and PPR.

Key principles from the standard:

- Each HL7v2 segment maps to one or more FHIR resources
- Repeating fields in HL7v2 may produce multiple FHIR elements or resources
- Code values must be translated using defined value set mappings
- Temporal values should use FHIR `dateTime` or `instant` formats

## Message-Level Mapping

| HL7v2 Message Type | Primary FHIR Resources | Secondary Resources |
|--------------------|------------------------|---------------------|
| ADT^A01 | Patient, Encounter | Location, Practitioner, Organization |
| ADT^A08 | Patient, Encounter | Location, Practitioner |
| ORU^R01 | Patient, DiagnosticReport, Observation | Practitioner, Organization |
| MDM^T02 | Patient, Composition, DocumentReference | Encounter, Practitioner |
| PPR^PC1 | Patient, Condition, Encounter | Practitioner, Organization |

## Mapping Approach

Ghega channels should implement mapping in discrete steps:

1. Parse the HL7v2 message into segments and fields
2. Map each segment to its target FHIR resource(s)
3. Translate coded values using mapping tables
4. Assemble resources into a Bundle or individual outputs
5. Validate the resulting FHIR against the relevant profile

## Deviations and Customization

Not all HL7v2 implementations follow the standard exactly. Common reasons for
deviation include:

- Custom Z-segments that carry domain-specific data
- Non-standard use of standard segments (e.g., using PV1 for non-visit data)
- Local code systems not present in the HL7 tables

When deviating:

- Document the deviation in the channel specification
- Maintain a mapping decision log
- Prefer extending standard resources over inventing new ones

## Version Considerations

| FHIR Version | Mapping Stability | Recommendation |
|--------------|-------------------|----------------|
| R4 | Stable | Use for all new channels |
| DSTU2 | Deprecated | Migrate to R4 if possible |
| R5 | Evolving | Evaluate before production use |

## Synthetic Data for Testing

Use the following patterns when creating mapping test fixtures:

- Patient names: `TESTPATIENT,ONE` or `SYNTHETIC,TWO`
- MRNs: `SYNTH-MRN-001`
- Dates: `1980-01-01` or `2020-03-15T10:30:00Z`
- Facilities: `GHEGA-TEST-FACILITY`

Never use real patient identifiers, real names, or real addresses in tests.
