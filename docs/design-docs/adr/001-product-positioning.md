# ADR-001: Product Positioning — Ghega Is a Modern Engine, Not a Mirth Clone

## Status

Accepted

## Context

Ghega targets healthcare integration teams that have outgrown Mirth Connect (NextGen Connect). The product must signal that it is a fresh, opinionated engine rather than a drop-in replacement. Positioning as a clone would limit architectural freedom, set incorrect user expectations, and invite direct feature parity debates that would constrain design decisions.

## Decision

Ghega is positioned as a modern healthcare integration engine, not a Mirth clone. It provides typed channel definitions, deterministic tests, durable message processing, replay safety, observability, migration tooling, and AI-assisted authoring. It does not claim Mirth runtime compatibility, feature parity, or drop-in replacement status.

## Consequences

- Architectural decisions are not constrained by Mirth compatibility.
- Migration tooling will translate Mirth structure into Ghega-native typed channels, not emulate Mirth behavior.
- Users must expect a learning curve and rethinking of channel design.
- Marketing and documentation must consistently reinforce the "beyond Mirth" positioning.

## Alternatives considered

- Position Ghega as a "Mirth-compatible migration target." Rejected (for now) because compatibility would force early commitment to Mirth-specific abstractions and reduce design space.

## References

- Marketing positioning brief (internal)
