---
name: getting-started
description: >
  Bootstrap and operate Ghega on a local machine. Use when you need to build
  Ghega from source, start the server, generate and validate channels, run
  Mirth migrations, or verify the system is working. This is the first skill
  to load when starting any Ghega task.
license: Apache-2.0
---

# Getting Started with Ghega

This skill enables an agent to build, run, and operate Ghega end-to-end on a
local machine. It is an orchestrator skill — for deep domain work, follow the
links to specialized skills.

## Prerequisites

- **Go 1.25+** (check with `go version`)
- **Node.js 18+** (for UI build only; check with `node --version`)
- **Make** (standard on macOS/Linux)

## Build

```bash
make ui-build   # build React UI (embedded into Go binary)
make build       # compile the ghega binary → ./ghega
```

Verify: `./ghega version` prints the version string.

To run tests: `make test` (all Go tests). To lint: `make lint`.

## Start the Server

```bash
./ghega serve
```

This starts:
- **HTTP API + UI** on `:8080`
- **MLLP listener** on `:2575`
- **SQLite** database at `./ghega.db` (created automatically)

Auth is **disabled by default** (dev mode). No OIDC setup needed for local work.

### Verify It Works

```bash
curl http://localhost:8080/healthz           # → {"status":"ok"}
curl http://localhost:8080/api/v1/me          # → dev user object
curl http://localhost:8080/api/v1/messages    # → message list (empty initially)
```

Open `http://localhost:8080` in a browser for the Ghega Console UI.

## Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_PORT` | `8080` | HTTP server port |
| `GHEGA_MLLP_HOST` | `0.0.0.0` | MLLP bind address |
| `GHEGA_MLLP_PORT` | `2575` | MLLP listener port |
| `GHEGA_DATABASE_URL` | `ghega.db` | SQLite database path |
| `GHEGA_DESTINATION_URL` | — | Default HTTP destination for messages |
| `GHEGA_AUTH_ENABLED` | `false` | Set `true` to enable OIDC auth |

For OIDC auth (optional), see [references/environment-reference.md](references/environment-reference.md).

## Core Workflows

### 1. Generate and Test a Channel

```bash
# Generate an MLLP-to-HTTP channel scaffold
./ghega generate channel mllp-to-http --name my-adt --out ./my-adt

# Validate the channel definition
./ghega channel validate ./my-adt/channel.yaml

# Run the channel's test fixtures
./ghega channel test ./my-adt/channel.yaml

# Deploy the channel revision to the local store
./ghega channel deploy ./my-adt/channel.yaml
```

There are two generator templates:
- `mllp-to-http` — HL7v2 MLLP source → HTTP destination
- `hl7v2-to-fhir` — HL7v2 MLLP source → FHIR destination

For channel authoring guidance, use the specialized skills:
- [creating-mllp-channels](../creating-mllp-channels/SKILL.md)
- [creating-http-channels](../creating-http-channels/SKILL.md)
- [creating-fhir-channels](../creating-fhir-channels/SKILL.md)

### 2. Channel Lifecycle (Deploy / Diff / Rollback)

```bash
# Compare local changes against deployed revision
./ghega channel diff ./my-adt/channel.yaml

# Rollback to a previous revision
./ghega channel rollback my-adt --to <hash>
```

Deploy, diff, and rollback manage **persisted channel revisions** in the SQLite
store. See [references/command-reference.md](references/command-reference.md) for all flags.

### 3. Mirth Migration

```bash
# Migrate a Mirth export directory
./ghega migrate mirth ./mirth-export --out ./migrated

# Generate a typed rewrite task for unconverted scripts
./ghega generate migration-task --channel my-channel --out ./tasks
```

This produces channel YAML, migration reports, and rewrite tasks. For full
migration workflows, use:
- [migrating-from-mirth](../migrating-from-mirth/SKILL.md)
- [writing-typed-rewrite-tasks](../writing-typed-rewrite-tasks/SKILL.md)

### 4. Watch Mode (Continuous Validation)

```bash
./ghega watch ./channels/
```

Watches a directory tree for `channel.yaml` files. On every change, it
re-runs validation and tests automatically. This is a **local dev loop** — it
validates and tests, but does not deploy or affect the running server.

### 5. UI Development

```bash
cd ui && npm install && npm run dev
```

Starts Vite on `:3000`, proxying `/api` requests to `:8080`. The server must
be running separately via `./ghega serve`.

## Current Limitations

Be aware of these when planning work:

| Area | Status |
|---|---|
| Channel runtime execution | `serve` runs a default MLLP→HTTP pipeline. It does **not** dynamically load deployed channel YAML. Deployed channels are persisted revisions, not live routing configs. |
| `/api/v1/channels` | Returns empty — not yet wired to the channel store. |
| Message replay/redelivery | CLI commands `message redeliver\|replay\|replay-preview` are **stubs**. API returns 501. |
| File/SFTP/DB connectors | Valid in channel schema, but no runtime implementation. |
| UI Settings/Operations | Placeholder pages. |
| Message delivery | Requires a running HTTP receiver at `GHEGA_DESTINATION_URL`. Without one, messages are ACK'd but delivery status is `failed`. |

## MLLP Message Testing

To send HL7v2 messages to the running MLLP listener:

```bash
# Using printf and netcat (install via brew install netcat)
printf '\x0bMSH|^~\\&|SendApp|SendFac|GhegaApp|GhegaFac|20240101120000||ADT^A01|MSG001|P|2.5\rPID|1||MRN12345^^^Hospital||TESTA^SYNTHETICA\r\x1c\r' | nc localhost 2575
```

The `\x0b` and `\x1c\r` are MLLP start/end framing bytes. You should receive
an ACK message back.

## Troubleshooting

| Problem | Solution |
|---|---|
| `make build` fails | Ensure Go 1.25+. Run `go mod download` first. |
| Port already in use | Set `GHEGA_PORT=9090` or `GHEGA_MLLP_PORT=2576`. |
| UI build fails | Ensure Node.js 18+. Run `cd ui && npm install`. |
| MLLP no response | Verify `./ghega serve` is running. Check MLLP port matches. |
| Messages show `failed` | Set `GHEGA_DESTINATION_URL` to a running HTTP receiver. |

## What To Do Next

After getting Ghega running, use these skills for domain work:

| Task | Skill |
|---|---|
| Create MLLP channels | `creating-mllp-channels` |
| Create HTTP channels | `creating-http-channels` |
| Create FHIR channels | `creating-fhir-channels` |
| Generate channel tests | `generating-channel-tests` |
| Debug HL7v2 messages | `debugging-hl7v2-messages` |
| Review mappings | `reviewing-mappings` |
| Migrate from Mirth | `migrating-from-mirth` |
| Map HL7v2 to FHIR | `mapping-hl7v2-to-fhir` |
| Understand the codebase | `understanding-ghega-codebase` |

## References

- [references/command-reference.md](references/command-reference.md) — CLI commands, flags, and examples
- [references/environment-reference.md](references/environment-reference.md) — all environment variables
- [references/workflow-recipes.md](references/workflow-recipes.md) — step-by-step recipes for common tasks
