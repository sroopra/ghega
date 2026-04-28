# ADR-009: PHI-Safe Logger Contract

## Status

Accepted

## Context

Ghega processes healthcare messages that may contain Protected Health Information (PHI). Structured and unstructured logs are a common source of PHI leaks. Establishing a logger contract that prevents payload bytes from entering log streams is essential for HIPAA compliance and operational safety.

## Decision

The Ghega logger contract prohibits emitting payload bytes in any log output. `Envelope.String()` and `PayloadRef.String()` must never include raw payload data. All log methods in `internal/logging/` accept only metadata fields (message IDs, channel IDs, status codes). Using `fmt.Printf` or `slog` with `Envelope` or `PayloadRef` directly is forbidden. Tests prove that synthetic payload bytes never appear in captured log output.

## Consequences

- Log analysis and alerting are safe to run without PHI scrubbing pipelines.
- Developers must be trained to avoid passing envelope or payload objects directly to formatters.
- Any new connector or parser must include log-safety tests.
- Violations are caught at the type level (String() contract) and by test assertions.

## Alternatives considered

- Rely on log scrubbing pipelines (e.g., Fluentd regex filters) to redact PHI after emission. Rejected (for now) because post-hoc scrubbing is unreliable and delays detection.

## References

- HIPAA logging requirements (internal)
- `pkg/payloadref/` implementation
