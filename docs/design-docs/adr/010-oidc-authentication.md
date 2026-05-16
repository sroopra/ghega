# ADR-010: OIDC Authentication

## Status

Accepted

## Context

Phase 5 requires production-grade authentication for the Ghega Console and API. The Console is a single-page application (SPA) served by the Go backend, and both the static assets and the API must be protected. The chosen pattern must balance security, user experience, and implementation complexity.

## Decision

Use OIDC with session cookies in a Backend-for-Frontend (BFF) pattern. The Go backend handles all token exchange; the SPA never sees OIDC tokens. CSRF double-submit cookies protect mutating API calls. A dev mode (`GHEGA_AUTH_ENABLED=false`) injects a fake developer user for local development.

## Consequences

- XSS risk is reduced because no tokens are stored in localStorage.
- Session revocation is possible because sessions are server-side.
- CSRF tokens are required for POST, PUT, and DELETE requests.
- Additional environment variables are needed to configure the OIDC provider and session signing secret.

## Alternatives considered

- Stateless JWT in localStorage. Rejected due to XSS risk and lack of session revocation.

## References

- ADR-007: SPA + BFF Authentication Pattern
- Security architecture guidelines (internal)
