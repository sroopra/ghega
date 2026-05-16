# Design Quality Overview

## Current State

Ghega is in active early development. The core architecture is established with accepted ADRs covering language choice, runtime boundaries, authentication, PHI safety, and FHIR support.

## Design Principles

See [design-docs/core-beliefs.md](design-docs/core-beliefs.md) for durable engineering principles.

## Architecture

See [ARCHITECTURE.md](../ARCHITECTURE.md) for the layered package map and dependency rules.

## Quality Gates

| Gate | Mechanism | Status |
|------|-----------|--------|
| PHI safety | Type-level enforcement + logger contract | Active |
| No Java/JS in runtime | CI structural tests + scripts | Active |
| Branding consistency | `scripts/check-branding.sh` | Active |
| Channel validation | `ghega channel validate` + `ghega channel test` | Active |
| Skill validation | `make validate-skills` | Active |
| UI dependency hygiene | `scripts/ui-dependency-audit.sh` | Active |
| Architecture layer enforcement | Planned — custom linter | Planned |

## Known Gaps

- `internal/runtime` is a placeholder — runtime orchestration not yet implemented
- Message replay/redelivery CLI and API are stubbed
- Channel list API returns empty until wired to `channelstore`
- UI Operations and Settings pages are placeholders
- Architecture layer dependency enforcement is not yet automated

## Design Documents

See [design-docs/index.md](design-docs/index.md) for the full index of ADRs and engineering policies.
