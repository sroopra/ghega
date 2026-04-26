# ADR-007: SPA + BFF Authentication Pattern

## Status

Proposed

## Context

The Ghega Console is a single-page application (SPA) that communicates with the Go backend. Authentication must protect both the static assets and the API. The pattern chosen (e.g., session cookies, JWT in localStorage, OAuth2/OIDC proxy) affects security, UX, and implementation complexity.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Stateless JWTs stored in browser localStorage. Rejected (for now) due to XSS risk and lack of session revocation.

## References

- Security architecture guidelines (internal)
