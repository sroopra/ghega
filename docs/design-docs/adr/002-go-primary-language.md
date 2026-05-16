# ADR-002: Go as the Primary Implementation Language

## Status

Accepted

## Context

Ghega needs a single primary language for the core engine that offers strong concurrency, fast compile times, a small runtime footprint, and excellent cross-platform deployment characteristics. The language choice influences hiring, library availability, and the ability to ship static binaries.

## Decision

Go is the primary implementation language for the Ghega engine runtime. All engine code, connectors, parsers, stores, and message-processing logic are written in Go.

## Consequences

- Fast compile times and small static binaries enable rapid deployment.
- Goroutines and channels provide a natural concurrency model for I/O-bound integration work.
- The Go ecosystem in healthcare is smaller than Java's, so custom parsers and connectors must be built.
- Hiring pool is different from traditional Mirth/Java shops; training may be required.

## Alternatives considered

- Rust: strong performance and safety guarantees, but steeper learning curve and longer compile times for the team.
- Java: deep healthcare library ecosystem, but heavier runtime and contrary to the goal of a lightweight engine.

## References

- Language evaluation matrix (internal)
