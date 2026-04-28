# ADR-005: UI Dependency Hygiene

## Status

Accepted

## Context

Because the Console is a TypeScript/React application (ADR-004), it is vulnerable to npm supply-chain risks, license contamination, and large bundle sizes. Establishing strict dependency hygiene early prevents technical debt and security exposure in the UI layer.

## Decision

All UI dependencies must have clearly compatible open-source licenses. `npm audit` and a custom dependency audit script (`scripts/ui-dependency-audit.sh`) are run in CI. Unused dependencies must be removed. Major framework upgrades require explicit review.

## Consequences

- Bundle size and attack surface are kept under control.
- License compatibility is verified automatically in CI.
- Adding a UI dependency requires justification and audit update.
- Development velocity is slightly slower due to audit gate.

## Alternatives considered

- Adopt a "few dependencies, mostly handwritten" philosophy (e.g., minimal UI frameworks, custom components). Rejected (for now) because it would slow initial development; may be revisited if bundle size becomes critical.

## References

- ADR-004 UI Exception — TypeScript and React Permitted for the Console
