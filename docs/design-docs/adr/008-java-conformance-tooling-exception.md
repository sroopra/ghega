# ADR-008: Java Conformance Tooling Exception

## Status

Accepted

## Context

Healthcare integration requires conformance with standards such as HL7v2 and FHIR. Some conformance testing tools and reference implementations are Java-based (e.g., HAPI FHIR). The runtime prohibits Java execution (ADR-003), but an exception may be needed for offline build-time or CI-time conformance validation.

## Decision

Java-based tools are permitted ONLY for external conformance testing, validation, and CI-time checks. They may never be required runtime dependencies of the Ghega engine. The runtime image must not contain a JVM. Conformance tools run in separate CI steps or containers.

## Consequences

- The engine runtime remains free of JVM dependencies.
- CI can still validate FHIR conformance using HAPI or similar tools in isolated steps.
- Users who need conformance validation must run it as a separate process.
- The project must maintain pure-Go parsers for production message processing.

## Alternatives considered

- Rewrite or wrap conformance tools in Go. Rejected (for now) because it is prohibitively expensive and duplicates mature, certified tooling.

## References

- ADR-003 Runtime Prohibition on Java and JavaScript Execution
- Conformance testing requirements (internal)
