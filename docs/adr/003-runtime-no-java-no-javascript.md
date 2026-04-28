# ADR-003: Runtime Prohibition on Java and JavaScript Execution

## Status

Accepted

## Context

To keep the engine lightweight, secure, and simple to deploy, the core runtime must avoid embedding a JVM or a JS engine. This boundary protects the project from runtime bloat, complex sandboxing, and supply-chain risks associated with multi-language runtimes.

## Decision

The Ghega engine runtime prohibits Java, the JVM, and JavaScript execution engines. No Java runtime, JRE, JDK, Node.js, npm, yarn, or pnpm may be present in the runtime image. Go packages may not import JavaScript execution engines (e.g., goja, otto, yaegi, v8go). This is enforced by CI and runtime boundary checks.

## Consequences

- Channel logic must be expressed in Go or declarative YAML, not JavaScript.
- The runtime image remains small (distroless static) and free of JVM/Node supply-chain risks.
- Users cannot bring existing Mirth JavaScript transformations without rewriting them as typed Go mappings.
- Java-based tools may still be used for external conformance testing, but never as required runtime dependencies.

## Alternatives considered

- Allow embedded JavaScript for user scripting (e.g., via Goja or Otto). Rejected (for now) because it introduces a second runtime, security surface, and debugging complexity.

## References

- Security boundary policy (internal)
