# Ghega Project Assessment — May 2026

Full-project analysis covering: build/test status, local testability, agent/skills readiness, Mirth feature gaps, and plan gaps. Use this document to prioritize next steps.

---

## 1. What Works Today

**Build & Test**: `make build`, `make test`, `make ui-build` all pass. Zero test failures.

### Functional Components

| Component | Package | Status | Notes |
|---|---|---|---|
| HL7v2 parser + ACK | `pkg/hl7v2` | ✅ Solid | Full parse/serialize/ACK cycle |
| MLLP listener | `pkg/mllp` | ✅ Solid | TCP framing, multi-message, tested |
| Mapping engine | `pkg/mapping` | ✅ Solid | copy, case, static, CEL transforms + HL7→FHIR |
| Channel system | `pkg/channel` | ✅ Solid | YAML schema, validate, test, deploy, diff, rollback |
| Message store | `pkg/messagestore` | ✅ Solid | Memory + SQLite, metadata + payload |
| Channel store | `pkg/channelstore` | ✅ Solid | Revisions, audit trail, rollback |
| HTTP sender | `pkg/httpsender` | ✅ Solid | Retries, timeouts, dry-run |
| FHIR sender | `pkg/fhirsender` | ✅ Solid | FHIR JSON, retries, dry-run |
| FHIR source server | `internal/fhirserver` | ✅ Solid | CRUD + bundle ingestion |
| MLLP→HTTP engine | `internal/engine` | ✅ Solid | End-to-end ingest → persist → map → send → ACK |
| HTTP server + embedded UI | `internal/server` | ✅ Works | `/healthz`, messages, alerts, migrations APIs |
| OIDC auth + sessions | `internal/server`, `internal/session` | ✅ Works | BFF pattern, CSRF, dev-bypass mode |
| Mirth migration | `pkg/migration`, `pkg/mirthxml` | ✅ Strong | XML parse → channel gen + reports + rewrite tasks |
| CLI | `internal/cli` | ✅ Works | serve, channel *, generate, watch, migrate |
| React UI | `ui/` | ✅ Builds | Home, Channels, Messages, Alerts, Operations, Migrations, Login |
| AI skills (16) | `ai/skills/` | ✅ All valid | All pass `make validate-skills` |

### Stubbed / Not Implemented

- `message redeliver`, `replay`, `replay-preview` → "not yet implemented"
- `POST /messages/{id}/redeliver|replay` → returns 501
- `/api/v1/channels` → returns empty (not wired to store)
- UI Settings page → placeholder
- UI Operations page → mostly placeholder
- `internal/runtime/` → empty placeholder

---

## 2. Local Testability Guide

### Prerequisites

- Go 1.25+
- Node.js (for UI)

### Build & Run

```bash
# Build everything
make ui-build && make build

# Start server (HTTP :8080, MLLP :2575)
./ghega serve

# Health check
curl http://localhost:8080/healthz          # → {"status":"ok"}
curl http://localhost:8080/api/v1/me        # → dev user info
```

### Channel Workflow

```bash
# Generate + validate + test + deploy a channel
./ghega generate channel mllp-to-http --name demo --out ./demo
./ghega channel validate ./demo/channel.yaml
./ghega channel test ./demo/channel.yaml
GHEGA_DATABASE_URL=./ghega.db ./ghega channel deploy ./demo/channel.yaml
```

### Mirth Migration

```bash
./ghega migrate mirth ./internal/cli/testdata/mirth-export --out ./migrated
```

### UI Dev Server

```bash
cd ui && npm install && npm run dev         # → :3000, proxies to :8080
```

### Key Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_PORT` | `8080` | HTTP server port |
| `GHEGA_MLLP_HOST` | `0.0.0.0` | MLLP bind address |
| `GHEGA_MLLP_PORT` | `2575` | MLLP listener port |
| `GHEGA_DATABASE_URL` | `ghega.db` | SQLite database path |
| `GHEGA_DESTINATION_URL` | — | Default HTTP destination URL |
| `GHEGA_AUTH_ENABLED` | `false` | Enable OIDC auth |

### What Will Not Work

- Sending HL7v2 via MLLP works (ACK returned), but message delivery requires a running receiver at `GHEGA_DESTINATION_URL`
- Channels page in UI shows nothing (API not wired)
- Replay/redelivery commands return errors
- Settings/Operations pages are placeholders

---

## 3. Agent & Skills Readiness

### Agent Entrypoint

`AGENTS.md` is the sole agent instruction file. No `.github/copilot-instructions.md` or MCP server exists.

### Skills Inventory

All 16 skills in `ai/skills/` are complete with `SKILL.md` + `references/`:

| Skill | Can Guide Real Work? | Notes |
|---|---|---|
| `creating-mllp-channels` | ✅ Yes | |
| `creating-http-channels` | ✅ Yes | |
| `creating-file-and-sftp-channels` | ⚠️ Schema only | No file/SFTP runtime |
| `creating-db-channels` | ⚠️ Schema only | No DB connector runtime |
| `creating-fhir-channels` | ✅ Yes | |
| `generating-channel-tests` | ✅ Yes | |
| `debugging-hl7v2-messages` | ✅ Yes | |
| `reviewing-mappings` | ✅ Yes | |
| `planning-replays` | ⚠️ Partial | Replay not implemented |
| `diagnosing-failed-deployments` | ✅ Yes | |
| `diagnosing-operations` | ⚠️ Partial | References non-existent commands |
| `reviewing-security` | ✅ Yes | |
| `understanding-ghega-codebase` | ✅ Yes | |
| `migrating-from-mirth` | ✅ Strong | |
| `writing-typed-rewrite-tasks` | ✅ Strong | |
| `mapping-hl7v2-to-fhir` | ✅ Yes | |

### Working Agent Workflows

- **Mirth migration**: `migrating-from-mirth` → `writing-typed-rewrite-tasks`
- **New channel authoring**: `creating-*` → `generating-channel-tests` → CLI deploy
- **FHIR modernization**: `mapping-hl7v2-to-fhir` → `creating-fhir-channels`

### Skills Gaps

- No quickstart/setup skill for agent onboarding
- No MCP server (planned as "later" in skills plan v8)
- Skills are documentation-only guidance, no bundled scripts/templates
- `make eval-skills` is a placeholder (no real LLM evaluation)
- Some skills reference CLI commands that don't exist (`ghega channel status`, `ghega validate`)

---

## 4. Mirth Feature Gap Analysis

### Critical Gaps (must-have for Mirth parity)

| # | Gap | Details | Effort |
|---|---|---|---|
| MG-01 | **Channel pipeline model** | No filters, no multi-destination, no fan-out, no conditional routing. Ghega is 1 source → 1 destination only. | Large — architectural |
| MG-02 | **ACK/error semantics** | Engine returns AA (success) even on mapping/send failure. Serious production blocker. | Medium |
| MG-03 | **Connector runtimes** | Only MLLP source + HTTP/FHIR destinations run. File, SFTP, DB exist in schema/migration but have no runtime. No TCP/SMTP/SOAP/JMS/DICOM. | Large |
| MG-04 | **Message operations** | No search, replay, redelivery, reprocessing, dead-letter queue. All stubbed in CLI and API. | Medium–Large |
| MG-05 | **Observability** | No metrics, no Prometheus, no OpenTelemetry, no channel stats, no throughput dashboards. | Medium |
| MG-06 | **Admin UI wiring** | Channels page empty (API not wired). Settings/Operations pages are placeholders. Can't manage channels from UI. | Medium |
| MG-07 | **RBAC / Authorization** | Session has roles field but zero enforcement. No user admin. No channel-scoped permissions. | Medium |

### Important Gaps (needed for enterprise adoption)

| # | Gap | Details |
|---|---|---|
| MG-08 | Multi-tenancy | Not implemented; `--tenant` only in plan examples |
| MG-09 | Channel groups/tags | Not implemented |
| MG-10 | Environment promotion | No `--env` overlays despite plan examples |
| MG-11 | Alert notifications | Alerts model exists but no email/Slack/PagerDuty delivery |
| MG-12 | Import/export bundles | No channel bundle format for sharing/promotion |
| MG-13 | Shared code/plugin model | Intentionally no JS (ADR-003), but no shared Go plugin/library story yet |
| MG-14 | HA / clustering / DR | No planning or implementation |
| MG-15 | Audit subsystem | No real audit/event model beyond channel store audit trail |

---

## 5. Plan & Documentation Gaps

### Execution Plan Status

| Plan | Status | Executed |
|---|---|---|
| `ghega_branding_plan_v3.md` | Active | ~80% — branding files exist, module path still mismatched |
| `ghega_agent_harness_plan_update.md` | Active | ~70% — naming landed, some docs reference unimplemented features |
| `ghega_ai_skills_plan_v8.md` | Active | ~85% — all 16 skills exist, eval/MCP still TODO |
| `tech-debt-tracker.md` | Active | 0% — no items resolved |

**No completed plans exist** — `docs/exec-plans/completed/` is empty.

### Missing Plans

| Area | Planning State |
|---|---|
| Multi-destination channel routing | ❌ No plan |
| Connector runtime (file/SFTP/DB) | ❌ No plan (skills describe them, code doesn't) |
| Message search/replay/reprocess | ⚠️ Mentioned in stubs + tech debt, no execution plan |
| Observability / metrics | ❌ No plan |
| RBAC / user admin | ❌ No plan |
| Multi-tenancy | ❌ No plan |
| HA / clustering / DR | ❌ No plan |
| Product specifications | ❌ Explicitly missing (`docs/product-specs/index.md` says none written) |
| MCP server | ⚠️ Named in branding + skills plan as "later" |
| Channel API wiring | ⚠️ In tech debt tracker, no execution plan |

### Known Contradictions

- `go.mod` says `github.com/sroopra/ghega`, branding says `github.com/ghega/ghega`
- `go.mod` requires Go 1.25, Dockerfile uses Go 1.26
- Skills reference CLI commands that don't exist (`ghega channel status`, `ghega validate`)
- README/skills promise file/SFTP/DB connectors that have no runtime

---

## 6. Recommended Priority Order

Based on the gaps above, suggested execution order:

1. **MG-02: Fix ACK/error semantics** — small scope, production-critical
2. **MG-06: Wire channels API** — unblocks UI, small scope
3. **MG-01: Channel pipeline model** — architectural foundation for everything else
4. **MG-03: Connector runtimes (file/DB first)** — delivers on existing promises
5. **MG-04: Message operations** — replay/redelivery/search
6. **MG-05: Observability** — metrics/stats for production readiness
7. **MG-07: RBAC** — enterprise requirement
8. **TD-001: Module path alignment** — pre-release blocker
9. **Product specs** — define "done" for each feature area
10. **MCP server** — enables full agent autonomy story

---

_Assessment generated 2026-05-16 by 5-agent parallel analysis (GPT-5.4). Reference this document when creating execution plans for gap items._
