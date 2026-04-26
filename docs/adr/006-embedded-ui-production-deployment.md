# ADR-006: Embedded UI in Production Deployment

## Status

Proposed

## Context

Ghega can be deployed as a single binary or as separate services. A key question is whether the Console (TypeScript/React UI) should be embedded into the Go server binary for production, served as a separate container, or both. Embedding simplifies small deployments; separation improves scaling and independent release cadence.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Serve the Console from a standalone static file server or CDN, with the Go API as a separate backend. Rejected (for now) because it complicates the out-of-box experience for small teams.

## References

- ADR-004 UI Exception — TypeScript and React Permitted for the Console
- Deployment topology discussion (internal)
