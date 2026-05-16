# CLI Command Reference

Complete reference for all `ghega` CLI commands.

## Top-Level

```
ghega <command> [subcommand] [flags]
ghega version          Print version
ghega help             Show help
ghega -h | --help      Show help
```

---

## `ghega serve`

Start the HTTP API server and MLLP listener.

```bash
ghega serve [--port PORT] [--migrations-dir DIR]
```

| Flag | Default | Env Var | Purpose |
|---|---|---|---|
| `--port` | `8080` | `GHEGA_PORT` | HTTP server port |
| `--migrations-dir` | — | `GHEGA_MIGRATIONS_DIR` | Directory for migration reports |

Also reads: `GHEGA_DATABASE_URL`, `GHEGA_MLLP_HOST`, `GHEGA_MLLP_PORT`, `GHEGA_DESTINATION_URL`, and all `GHEGA_AUTH_*`/`GHEGA_OIDC_*` vars.

Starts:
- HTTP server with embedded UI on the configured port
- MLLP listener on `GHEGA_MLLP_HOST:GHEGA_MLLP_PORT`
- SQLite store at `GHEGA_DATABASE_URL` (falls back to in-memory on failure)

### API Endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | Health check |
| GET | `/api/v1/me` | Current user info |
| GET | `/api/v1/messages` | List messages |
| GET | `/api/v1/messages/{id}` | Get message by ID |
| POST | `/api/v1/messages/{id}/redeliver` | ⚠️ Not implemented (501) |
| POST | `/api/v1/messages/{id}/replay` | ⚠️ Not implemented (501) |
| GET | `/api/v1/channels` | ⚠️ Returns empty list |
| GET | `/api/v1/alerts` | List alerts |
| GET | `/api/v1/migrations` | List migrations |
| GET | `/api/v1/migrations/{id}` | Get migration by ID |
| GET | `/` | Embedded UI (Ghega Console) |

---

## `ghega channel`

Manage channel definitions. All subcommands work with channel YAML files.

### `channel validate <path>`

Validate a channel YAML file against the schema and policy rules.

```bash
ghega channel validate ./my-channel/channel.yaml
```

Output: `channel is valid` on success, error details on failure.

### `channel test <path> [--junit PATH]`

Run test fixtures defined in the channel YAML.

```bash
ghega channel test ./my-channel/channel.yaml
ghega channel test ./my-channel/channel.yaml --junit ./results.xml
```

| Flag | Purpose |
|---|---|
| `--junit` | Write JUnit XML report to this path |

### `channel deploy <path>`

Deploy a channel revision to the local SQLite store.

```bash
ghega channel deploy ./my-channel/channel.yaml
# or with explicit DB:
GHEGA_DATABASE_URL=./ghega.db ghega channel deploy ./my-channel/channel.yaml
```

**Note:** Deploy persists the channel revision. It does **not** make the running server execute this channel's routing rules.

### `channel diff <path>`

Compare local channel YAML against the currently deployed revision.

```bash
ghega channel diff ./my-channel/channel.yaml
```

### `channel rollback <name> --to <hash>`

Roll back a channel to a previous revision.

```bash
ghega channel rollback my-adt --to abc123
```

| Flag | Required | Purpose |
|---|---|---|
| `--to` | Yes | Revision hash to roll back to |

---

## `ghega generate`

Scaffold new artifacts.

### `generate channel mllp-to-http`

```bash
ghega generate channel mllp-to-http --name my-adt --out ./my-adt
```

| Flag | Required | Default | Purpose |
|---|---|---|---|
| `--name` | Yes | — | Channel name |
| `--message-type` | No | `ADT_A01` | HL7v2 message type |
| `--out` | Yes | — | Output directory |

Creates: `channel.yaml`, `tests/fixture.yaml`, `fixtures/sample.hl7`, `fixtures/minimal.hl7`

### `generate channel hl7v2-to-fhir`

```bash
ghega generate channel hl7v2-to-fhir --name my-fhir --out ./my-fhir
```

Same flags as `mllp-to-http`. Creates: `channel.yaml`, `tests/fixture.yaml`, `fixtures/sample.hl7`, `fixtures/expected.json`

### `generate migration-task`

```bash
ghega generate migration-task --channel my-channel --out ./tasks \
    --category script-rewrite --severity high \
    --description "Convert E4X date formatting"
```

| Flag | Required | Default | Purpose |
|---|---|---|---|
| `--channel` | Yes | — | Channel name |
| `--out` | Yes | — | Output directory |
| `--category` | No | — | Task category |
| `--severity` | No | `medium` | Task severity |
| `--description` | No | — | Task description |

---

## `ghega migrate`

### `migrate mirth <path> --out <dir>`

Convert a Mirth Connect export into Ghega channel artifacts.

```bash
ghega migrate mirth ./mirth-export/ --out ./migrated
ghega migrate mirth ./single-channel.xml --out ./migrated
```

| Flag | Required | Purpose |
|---|---|---|
| `--out` | Yes | Output directory for generated artifacts |
| `--samples` | No | Sample messages directory |
| `--expected` | No | Expected output directory |

Accepts either a directory of XML files or a single XML export file.

---

## `ghega watch <directory>`

Watch a directory for `channel.yaml` changes and re-run validation/tests.

```bash
ghega watch ./channels/
```

Polls every 1 second. Stops on `Ctrl+C` (SIGINT/SIGTERM). This is a
**validate-and-test loop only** — it does not deploy or affect the server.

---

## `ghega message` (stubs)

These commands exist but are **not yet implemented**:

```bash
ghega message redeliver <id> --destination <dest>   # → "not yet implemented"
ghega message replay <id> --as-new                  # → "not yet implemented"
ghega message replay-preview <id>                   # → "not yet implemented"
```
