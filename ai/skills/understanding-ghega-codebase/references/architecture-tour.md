# Ghega Architecture Tour

## Layer Overview

Ghega is organized into three layers:

1. **CLI Layer** (`cmd/`, `internal/cli/`): User-facing commands and HTTP server.
2. **Runtime Layer** (`internal/runtime/`, `internal/config/`): Message processing,
   channel execution, and configuration management.
3. **Platform Layer** (`pkg/`, `internal/logging/`): Reusable types and utilities,
   including PHI-safe abstractions.

## Package Responsibilities

### `cmd/ghega/`

The main entrypoint. Initializes the CLI router and delegates to `internal/cli/`.

### `internal/cli/`

Implements commands such as:

- `version` — prints build version info
- `serve` — starts the HTTP server with health endpoints
- `channel validate` — validates a channel YAML file
- `generate` — generates channel scaffolding and tests

### `internal/config/`

Loads and validates Ghega configuration from files and environment variables.
Environment variables use the `GHEGA_` prefix.

### `internal/logging/`

Provides structured, PHI-safe logging. Key rules:

- Log message IDs, timestamps, and channel IDs
- Never log payload bytes or field values
- Use explicit methods like `LogMessageReceived(metadata)`

### `internal/runtime/`

The channel runtime. Responsibilities include:

- Reading messages from source connectors (MLLP, HTTP, file, etc.)
- Applying mappings and transformations
- Writing messages to destination connectors
- Managing `Envelope` and `PayloadRef` lifecycle

### `pkg/payloadref/`

Public API for payload references. Contains:

- `PayloadRef` — a storage reference to a payload (no bytes held)
- `Envelope` — metadata wrapper containing a `PayloadRef`

Both types implement `String()` to ensure payload bytes never appear in logs.

## Data Flow

```
Source Connector
       |
       v
  Envelope + PayloadRef
       |
       v
  Channel Runtime (mapping, transform)
       |
       v
  Destination Connector
       |
       v
  Audit Log (metadata only)
```

## Contributing Guidelines

- Add new CLI commands in `internal/cli/` and register them in `cmd/ghega/`
- Add reusable types in `pkg/` if they are part of the public API
- Keep runtime-specific code in `internal/runtime/`
- Write tests for every new package
- Run `make test` and `make lint` before committing
