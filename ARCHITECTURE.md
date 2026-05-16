# Ghega Architecture

## Overview

Ghega is an open-source healthcare integration engine. It processes healthcare messages (HL7v2, FHIR) through typed channel definitions with deterministic testing, durable storage, and AI-assisted authoring.

## Layered Architecture

Ghega follows a rigid layered architecture. Dependencies flow downward only. Cross-cutting concerns enter through explicit provider interfaces.

```text
┌─────────────────────────────────────────────────┐
│  UI Layer                                       │
│  ui/ (React SPA, embedded in Go binary)         │
├─────────────────────────────────────────────────┤
│  Runtime Layer                                  │
│  internal/server    HTTP API + BFF auth          │
│  internal/engine    MLLP processing pipeline     │
│  internal/fhirserver FHIR REST source            │
│  internal/cli       Command orchestration        │
├─────────────────────────────────────────────────┤
│  Service Layer                                  │
│  pkg/channel        Channel lifecycle            │
│  pkg/mapping        HL7v2/FHIR mapping engines   │
│  pkg/migration      Mirth conversion + reports   │
├─────────────────────────────────────────────────┤
│  Repo Layer (Storage)                           │
│  pkg/messagestore   Message metadata + payload   │
│  pkg/channelstore   Channel revisions + audit    │
│  internal/session   Session store                │
│  internal/alerts    Alert store                  │
├─────────────────────────────────────────────────┤
│  Config Layer                                   │
│  internal/config    Environment-driven config    │
│  branding/          Product metadata             │
├─────────────────────────────────────────────────┤
│  Types Layer                                    │
│  pkg/payloadref     Envelope, PayloadRef         │
│  pkg/hl7v2          HL7v2 message types          │
│  pkg/fhir           FHIR R4 resource types       │
│  pkg/mllp           MLLP framing types           │
└─────────────────────────────────────────────────┘
```

## Dependency Rules

| Layer | May depend on | Must NOT depend on |
|-------|--------------|-------------------|
| Types | standard library only | Config, Repo, Service, Runtime, UI |
| Config | Types | Repo, Service, Runtime, UI |
| Repo | Types, Config | Service, Runtime, UI |
| Service | Types, Config, Repo | Runtime, UI |
| Runtime | Types, Config, Repo, Service | UI |
| UI | Runtime (via HTTP API) | Go packages directly |

## Key Packages

### Types (`pkg/payloadref`, `pkg/hl7v2`, `pkg/fhir`, `pkg/mllp`)

Foundation types shared across the codebase. `PayloadRef` and `Envelope` enforce PHI-safe metadata-only handling at the type level.

### Config (`internal/config`)

Environment-driven configuration. All settings come from environment variables prefixed with `GHEGA_`.

### Repo (`pkg/messagestore`, `pkg/channelstore`)

Storage interfaces with dual implementations (in-memory for tests, SQLite for production). Each store follows the pattern: interface → memory impl → SQLite impl.

### Service (`pkg/channel`, `pkg/mapping`, `pkg/migration`)

Business logic. Channel validation, testing, deployment, diffing, and rollback. Mapping engines for HL7v2 flat mappings (with CEL expressions) and HL7v2-to-FHIR conversion. Mirth migration report generation.

### Runtime (`internal/server`, `internal/engine`, `internal/cli`)

Application orchestration. The HTTP server serves the API and embedded SPA. The engine wires the MLLP pipeline (receive → persist → map → send → update status → ACK). The CLI dispatches commands.

### UI (`ui/`)

React + TypeScript SPA built with Vite. Communicates exclusively through the Go HTTP API. Embedded into the Go binary at build time via `embed.go`.

## Cross-Cutting Concerns

### PHI Safety

Enforced at the type level through `PayloadRef` and `Envelope` contracts. The `internal/logging` package only accepts metadata fields. See `docs/design-docs/phi-logging-guidance.md`.

### Authentication

OIDC with session cookies in a BFF pattern. The Go backend handles all token exchange. The SPA never stores tokens. See ADR-007 and ADR-010.

### AI Skills

Skills live in `ai/skills/` and are validated by `internal/skills/validate`. They are documentation artifacts, not runtime code — they guide agents but never execute in the message path.

## Connector Model

```text
Source → Engine → Destination

Sources:   MLLP, FHIR REST, HTTP, File, SFTP, Database
Engine:    Parse → Validate → Map → Transform
Destinations: HTTP, FHIR, File, SFTP, Database
```

Each connector type is defined in `channel.yaml` with typed source/destination configuration.
