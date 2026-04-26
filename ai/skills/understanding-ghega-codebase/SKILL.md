---
name: understanding-ghega-codebase
description: >
  Use when a user wants to understand, navigate, or contribute to the Ghega codebase.
  Use when someone asks about project structure, architecture, where to find code,
  or how a particular feature is implemented.
  Use when the topic involves Go packages, internal vs. pkg boundaries,
  CLI commands, configuration, logging, or channel runtime.
license: Apache-2.0
---

# Understanding the Ghega Codebase

This skill provides orientation for navigating and understanding the Ghega codebase.

## When to Use

- Onboarding to the Ghega project
- Finding where a specific feature is implemented
- Understanding the boundary between `internal/` and `pkg/`
- Contributing a new command, package, or channel type

## Directory Structure

```
cmd/ghega/         # Main CLI entrypoint
internal/cli/      # CLI command implementations
internal/config/   # Configuration loading and validation
internal/logging/  # PHI-safe logging wrappers
internal/runtime/  # Channel runtime and message processing
pkg/payloadref/    # PayloadRef and Envelope types (public API)
```

## Key Design Principles

1. **No Java or JavaScript in the runtime**: Ghega is written in Go. The only
   JavaScript exception is for the embedded UI (TypeScript/React).
2. **PHI-Safe Logging**: Never log payload bytes. Use `PayloadRef` and metadata-only
   logging via `internal/logging/`.
3. **Deterministic Channels**: Channel behavior must be testable and reproducible.
4. **Typed Configurations**: Channel definitions are YAML with validated schemas.

## Boundaries

- **`internal/`**: Packages private to the Ghega project. Cannot be imported by
  external modules.
- **`pkg/`**: Public API packages that external projects may import.
- **`cmd/`**: Application entrypoints.

## Safety

When modifying the codebase:

- Do not introduce JavaScript execution engines into Go packages
- Do not log payload content in any form
- Use synthetic data in all tests and examples
- Never commit secrets, API keys, or credentials

## References

- See [references/architecture-tour.md](references/architecture-tour.md) for a detailed walkthrough of the codebase architecture.
