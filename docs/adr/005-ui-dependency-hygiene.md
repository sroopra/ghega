# ADR-005: UI Dependency Hygiene

## Status

Proposed

## Context

Because the Console is a TypeScript/React application (ADR-004), it is vulnerable to npm supply-chain risks, license contamination, and large bundle sizes. Establishing strict dependency hygiene early prevents technical debt and security exposure in the UI layer.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Adopt a "few dependencies, mostly handwritten" philosophy (e.g., minimal UI frameworks, custom components). Rejected (for now) because it would slow initial development; may be revisited if bundle size becomes critical.

## References

- ADR-004 UI Exception — TypeScript and React Permitted for the Console
