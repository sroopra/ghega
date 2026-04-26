# ADR-009: PHI-Safe Logger Contract

## Status

Proposed

## Context

Ghega processes healthcare messages that may contain Protected Health Information (PHI). Structured and unstructured logs are a common source of PHI leaks. Establishing a logger contract that prevents payload bytes from entering log streams is essential for HIPAA compliance and operational safety.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Rely on log scrubbing pipelines (e.g., Fluentd regex filters) to redact PHI after emission. Rejected (for now) because post-hoc scrubbing is unreliable and delays detection.

## References

- HIPAA logging requirements (internal)
- `pkg/payloadref/` implementation
