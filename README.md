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

### Authentication

Ghega uses OIDC with session cookies in a BFF pattern. The Go backend handles all token exchange, and the SPA never stores tokens in localStorage. The UI handles login and logout automatically.

Configure OIDC with the following environment variables:

| Variable | Description |
|----------|-------------|
| `GHEGA_AUTH_ENABLED` | Set to `true` to enable authentication. Defaults to `false` (dev mode). |
| `GHEGA_OIDC_ISSUER` | OIDC provider issuer URL (required when auth is enabled). |
| `GHEGA_OIDC_CLIENT_ID` | OIDC client ID (required when auth is enabled). |
| `GHEGA_OIDC_CLIENT_SECRET` | OIDC client secret (required when auth is enabled). |
| `GHEGA_OIDC_REDIRECT_URL` | OIDC redirect URL. Defaults to `http://localhost:8080/auth/callback`. |
| `GHEGA_SESSION_SECRET` | Secret used to sign session cookies (required when auth is enabled). |

When `GHEGA_AUTH_ENABLED=false`, the server runs in dev mode and injects a fake developer user so the Console can be used without an identity provider.

### FHIR Support

Ghega supports FHIR R4 as a source, destination, and mapping target. Channels can declare `source.type: fhir` to accept FHIR REST interactions, or `destination.type: fhir` to send FHIR JSON to external servers.

Generate an HL7v2-to-FHIR channel scaffold:

```bash
ghega generate channel hl7v2-to-fhir --name adt-to-fhir --out ./channels/adt-to-fhir
```

The mapping engine converts HL7v2 segments to FHIR resources:

| HL7v2 Segment | FHIR Resource |
|---------------|---------------|
| PID | Patient |
| PV1 | Encounter |
| OBX | Observation |
| OBR | DiagnosticReport |
| MSH | MessageHeader |

Supported Bundle types: `batch`, `transaction`, `searchset`, `history`, `collection`.

See `docs/adr/011-fhir-support.md` for architecture details.

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
