# ADR-003: Runtime Prohibition on Java and JavaScript Execution

## Status

Proposed

## Context

To keep the engine lightweight, secure, and simple to deploy, the core runtime must avoid embedding a JVM or a JS engine. This boundary protects the project from runtime bloat, complex sandboxing, and supply-chain risks associated with multi-language runtimes.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Allow embedded JavaScript for user scripting (e.g., via Goja or Otto). Rejected (for now) because it introduces a second runtime, security surface, and debugging complexity.

## References

- Security boundary policy (internal)
