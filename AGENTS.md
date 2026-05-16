# Ghega — Agent Guide

Ghega is an open-source healthcare integration engine. This file is the entrypoint for agents working in the codebase.

## Quick Links

| Document | Purpose |
|----------|---------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Package/layer map and dependency rules |
| [docs/DESIGN.md](docs/DESIGN.md) | Design quality overview |
| [docs/design-docs/index.md](docs/design-docs/index.md) | Design documents and ADRs |
| [docs/design-docs/core-beliefs.md](docs/design-docs/core-beliefs.md) | Durable engineering principles |
| [docs/exec-plans/](docs/exec-plans/) | Active and completed execution plans |
| [docs/product-specs/index.md](docs/product-specs/index.md) | Product specifications |
| [docs/references/](docs/references/) | External reference material |

## Repository Layout

```text
cmd/ghega/              CLI entrypoint
cmd/validate-skills/    Skill validation tool
internal/               Private application code
  cli/                  Command handlers (serve, channel, generate, migrate, watch)
  server/               HTTP API + embedded SPA + auth
  engine/               MLLP processing pipeline
  fhirserver/           FHIR REST source connector
  config/               Environment-driven configuration
  logging/              PHI-safe logging wrapper
  session/              Signed cookie session management
  alerts/               Alert model + store
  runtime/              Runtime boundary (placeholder)
  skills/validate/      AI skill validator
pkg/                    Public library packages
  channel/              Channel schema, validation, test runner, deploy/diff/rollback
  mapping/              HL7v2 mapping engine + CEL + FHIR mapping
  messagestore/         Message metadata/payload store (interface + memory/sqlite)
  channelstore/         Channel revision store (interface + memory/sqlite)
  payloadref/           PHI-safe payload references
  hl7v2/                HL7v2 parser + ACK generation
  mllp/                 MLLP TCP listener/framing
  httpsender/           HTTP sender with retries
  fhirsender/           FHIR-specific HTTP sender
  fhir/                 Minimal FHIR R4 types + bundle helpers
  mirthxml/             Mirth XML export parser
  migration/            Mirth-to-Ghega conversion + reports
ai/skills/              AI-assisted authoring skills (16 skills)
ui/                     React + TypeScript Console (embedded in Go binary)
branding/               Product metadata (product.yaml, visual-identity.md)
scripts/                CI scripts, boundary tests, branding checks
docs/                   Knowledge base (system of record)
```

## Code Conventions

- **Public API in `pkg/`**, private code in `internal/`
- **Interface-first storage**: every store has an interface + memory + SQLite impl
- **Dependency injection via constructors** with functional options pattern
- **PHI-safe logging**: never log payload bytes; use `internal/logging.Logger`
- **No Java/JS in the Go runtime** (enforced by CI + structural tests)
- **Synthetic data only** in examples, tests, and skills — never real PHI

## Skills

Ghega ships 16 local AI skills in `ai/skills/`. Each has a `SKILL.md` with YAML frontmatter and a `references/` directory. Validate with `make validate-skills`.

See [docs/exec-plans/active/ghega_ai_skills_plan_v8.md](docs/exec-plans/active/ghega_ai_skills_plan_v8.md) for the skills strategy.

## Build & Test

```bash
make build          # compile ghega binary
make test           # run all Go tests
make lint           # go vet + golangci-lint
make docker         # build container image
make validate-skills # validate ai/skills/
make ui-build       # build React UI
make adr NEW="title" # create new ADR
```

## Key Constraints

1. All architectural decisions are recorded in `docs/design-docs/adr/`
2. PHI safety is enforced at the type level — see [docs/design-docs/phi-logging-guidance.md](docs/design-docs/phi-logging-guidance.md)
3. The UI is the only TypeScript/React exception — see ADR-004
4. Branding is centralized in `branding/product.yaml`
