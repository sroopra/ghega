# ADR-011: FHIR Support

## Status

Accepted

## Context

Phase 6 adds FHIR R4 support to Ghega. Healthcare interoperability is moving from HL7v2 to FHIR, and a Mirth alternative must support both. The key question was whether to import a heavy external FHIR library or build minimal custom types.

## Decision

Implement minimal custom FHIR R4 Go structs rather than importing an external library. Add `fhir` as a first-class source and destination type in channel YAML. Extend the mapping engine with a separate HL7v2-to-FHIR path (`FHIREngine`) rather than overloading the flat `map[string]string` engine.

## Consequences

- Lighter dependency tree with no external FHIR library.
- Custom types require manual updates when the FHIR spec changes.
- Consistent with the Go-native philosophy established in ADR-002.
- Channels can declare `source.type: fhir` and `destination.type: fhir`.
- The test runner supports JSON deep comparison via `ExpectedJSON` for FHIR output validation.

## Details

### Supported FHIR types

- `Bundle` (batch, transaction, searchset, history, collection)
- `Patient`, `Encounter`, `Observation`, `DiagnosticReport`, `MessageHeader`
- `OperationOutcome` for error responses
- Supporting types: `HumanName`, `Identifier`, `CodeableConcept`, `Coding`, `Quantity`, `Reference`, `Period`, `Address`, `ContactPoint`

### Content type

Canonical: `application/fhir+json; fhirVersion=4.0`

### Mapping

- PID → Patient
- PV1 → Encounter
- OBX → Observation
- OBR → DiagnosticReport
- MSH → MessageHeader

## References

- `pkg/fhir/` — minimal FHIR R4 types
- `pkg/mapping/fhir.go` — HL7v2-to-FHIR mapping engine
- `internal/fhirserver/` — FHIR REST source connector
- `pkg/fhirsender/` — FHIR destination connector
- `ai/skills/mapping-hl7v2-to-fhir/references/common-mappings.md`
