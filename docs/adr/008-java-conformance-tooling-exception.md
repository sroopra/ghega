# ADR-008: Java Conformance Tooling Exception

## Status

Proposed

## Context

Healthcare integration requires conformance with standards such as HL7v2 and FHIR. Some conformance testing tools and reference implementations are Java-based (e.g., HAPI FHIR). The runtime prohibits Java execution (ADR-003), but an exception may be needed for offline build-time or CI-time conformance validation.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Rewrite or wrap conformance tools in Go. Rejected (for now) because it is prohibitively expensive and duplicates mature, certified tooling.

## References

- ADR-003 Runtime Prohibition on Java and JavaScript Execution
- Conformance testing requirements (internal)
