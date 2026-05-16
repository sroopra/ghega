# ADR-007: SPA + BFF Authentication Pattern

## Status

Accepted

## Context

The Ghega Console is a single-page application (SPA) that communicates with the Go backend. Authentication must protect both the static assets and the API. The pattern chosen affects security, UX, and implementation complexity.

## Decision

Ghega uses a Backend-for-Frontend (BFF) authentication pattern. The Go server acts as the BFF: it validates sessions or tokens, proxies to identity providers if needed, and serves the SPA. The SPA does not store tokens in localStorage. Production authentication uses OIDC with session cookies; a placeholder middleware validates bearer tokens in development.

## Consequences

- The Go server controls authentication state, reducing XSS risk.
- Session cookies require CSRF protection for mutating API calls.
- The BFF can inject channel-scoped permissions into API responses.
- OIDC integration is deferred to Phase 5; Phase 2–3 use placeholder auth.

## Alternatives considered

- Stateless JWTs stored in browser localStorage. Rejected (for now) due to XSS risk and lack of session revocation.

## References

- Security architecture guidelines (internal)
