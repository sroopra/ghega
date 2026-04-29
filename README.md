# Ghega

Ghega is an open-source healthcare integration engine for teams that want to move beyond Mirth.

It provides typed channel definitions, deterministic tests, durable message processing, replay safety, observability, migration tooling, and AI-assisted authoring for HL7v2, FHIR, MLLP, HTTP, SFTP, files, and database integrations.

## Getting Started

Clone the repository and build the binary:

```bash
git clone https://github.com/ghega/ghega.git
cd ghega
```

## Building

Build the `ghega` binary to the repository root:

```bash
make build
```

This compiles `cmd/ghega` into a `./ghega` binary.

## Running

Start the Ghega HTTP server:

```bash
./ghega serve
```

The server listens on port `8080` by default. Override it with the `--port` flag or the `GHEGA_PORT` environment variable.

### Migration from Mirth

Import Mirth Connect channel exports:

```bash
ghega migrate mirth ./mirth-export --out ./migrated
```

The command produces per-channel reports in `./migrated/` with auto-converted mappings and typed rewrite tasks for patterns that could not be migrated automatically.

## Testing

Run the full test suite:

```bash
make test
```

This executes all Go tests (`go test ./...`) and validation checks, including runtime boundary verification.

## Architecture

- `cmd/ghega/` — CLI entrypoint.
- `internal/cli/` — Command handlers (`serve`, `channel`, `generate`, `version`).
- `internal/config/` — Configuration loading.
- `internal/logging/` — PHI-safe logging wrapper.
- `internal/runtime/` — Runtime boundary validation.
- `pkg/payloadref/` — Core types (`Envelope`, `PayloadRef`) for metadata-only message handling.
- `ai/skills/` — AI-assisted authoring skills.
- `docs/` — Architecture Decision Records (ADRs) and guidance.
- `branding/` — Product metadata and visual identity placeholders.

See `docs/adr/` for architectural decisions and `docs/phi-logging-guidance.md` for logging conventions.
