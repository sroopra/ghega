# ADR-006: Embedded UI in Production Deployment

## Status

Accepted

## Context

Ghega can be deployed as a single binary or as separate services. A key question is whether the Console (TypeScript/React UI) should be embedded into the Go server binary for production, served as a separate container, or both. Embedding simplifies small deployments; separation improves scaling and independent release cadence.

## Decision

The Ghega Console UI is served as static assets by the Go HTTP server (`ghega serve`). The built UI is embedded or served from the filesystem alongside the binary. For large-scale deployments, the UI may be served from a CDN or static file server with the Go API as the BFF backend.

## Consequences

- Single-binary deployments are simple: `./ghega serve` serves both API and UI.
- Large deployments can separate UI and API for independent scaling.
- UI build artifacts must be included in the container image or deployment package.
- No hidden production UI state: all UI configuration comes from the API.

## Alternatives considered

- Serve the Console from a standalone static file server or CDN, with the Go API as a separate backend. Rejected (for now) because it complicates the out-of-box experience for small teams.

## References

- ADR-004 UI Exception — TypeScript and React Permitted for the Console
- Deployment topology discussion (internal)
