# Design Documents

## Architecture Decision Records (ADRs)

All ADRs are in `adr/`. Create new ADRs with `make adr NEW="My Decision Title"`.

| ADR | Title | Status |
|-----|-------|--------|
| [001](adr/001-product-positioning.md) | Product Positioning — Modern Engine, Not a Mirth Clone | Accepted |
| [002](adr/002-go-primary-language.md) | Go as Primary Language | Accepted |
| [003](adr/003-runtime-no-java-no-javascript.md) | No Java or JavaScript in Runtime | Accepted |
| [004](adr/004-ui-typescript-react-exception.md) | UI TypeScript/React Exception | Accepted |
| [005](adr/005-ui-dependency-hygiene.md) | UI Dependency Hygiene | Accepted |
| [006](adr/006-embedded-ui-production-deployment.md) | Embedded UI for Production Deployment | Accepted |
| [007](adr/007-spa-bff-authentication-pattern.md) | SPA + BFF Authentication Pattern | Accepted |
| [008](adr/008-java-conformance-tooling-exception.md) | Java Conformance Tooling Exception | Accepted |
| [009](adr/009-phi-safe-logger-contract.md) | PHI-Safe Logger Contract | Accepted |
| [010](adr/010-oidc-authentication.md) | OIDC Authentication | Accepted |
| [011](adr/011-fhir-support.md) | FHIR Support | Accepted |

## Engineering Policies

| Document | Purpose |
|----------|---------|
| [core-beliefs.md](core-beliefs.md) | Durable engineering principles |
| [phi-logging-guidance.md](phi-logging-guidance.md) | PHI-safe logging rules and patterns |

## See Also

- [ARCHITECTURE.md](../../ARCHITECTURE.md) — package/layer map
- [docs/DESIGN.md](../DESIGN.md) — design quality overview
