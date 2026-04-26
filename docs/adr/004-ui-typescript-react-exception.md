# ADR-004: UI Exception — TypeScript and React Permitted for the Console

## Status

Proposed

## Context

The Ghega Console is the user-facing web interface for managing channels, viewing logs, and debugging flows. A modern, interactive UI is essential for adoption. While the engine runtime avoids JavaScript (ADR-003), the Console is a separate deployable unit and benefits from the mature React/TypeScript ecosystem.

## Decision

TBD

## Consequences

TBD

## Alternatives considered

- Build the Console entirely in Go using server-rendered HTML templates (e.g., HTMX). Rejected (for now) because it would significantly slow down feature delivery for a rich interactive UI.

## References

- ADR-003 Runtime Prohibition on Java and JavaScript Execution
