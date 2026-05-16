# ADR-004: UI Exception — TypeScript and React Permitted for the Console

## Status

Accepted

## Context

The Ghega Console is the user-facing web interface for managing channels, viewing logs, and debugging flows. A modern, interactive UI is essential for adoption. While the engine runtime avoids JavaScript (ADR-003), the Console is a separate deployable unit and benefits from the mature React/TypeScript ecosystem.

## Decision

TypeScript and React are permitted exclusively for the Ghega Console UI. The engine runtime remains free of JavaScript execution. UI code is confined to the `ui/` directory and is built into static assets. No TypeScript or JavaScript may exist in `internal/`, `pkg/`, or `cmd/`.

## Consequences

- The UI can leverage the full React/Vite/TypeScript ecosystem for rapid feature delivery.
- The runtime boundary (no JS in engine) remains intact.
- UI build is separate from engine build; CI must verify both.
- UI dependencies must be audited for supply-chain risks (see ADR-005).

## Alternatives considered

- Build the Console entirely in Go using server-rendered HTML templates (e.g., HTMX). Rejected (for now) because it would significantly slow down feature delivery for a rich interactive UI.

## References

- ADR-003 Runtime Prohibition on Java and JavaScript Execution
