# Environment Variable Reference

All `GHEGA_*` environment variables recognized by Ghega.

## Server

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_PORT` | `8080` | HTTP server listen port |
| `GHEGA_DATABASE_URL` | `ghega.db` | Path to SQLite database file |
| `GHEGA_MIGRATIONS_DIR` | — | Directory containing migration reports to serve via API |

## MLLP

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_MLLP_HOST` | `0.0.0.0` | MLLP listener bind address |
| `GHEGA_MLLP_PORT` | `2575` | MLLP listener port |

## Destination

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_DESTINATION_URL` | — | Default HTTP URL for message delivery. Without this, message delivery will fail (though MLLP ACKs still return). |

## Authentication (Optional)

Auth is **disabled by default**. Set `GHEGA_AUTH_ENABLED=true` to require OIDC
login. When enabled, all of the following must be set:

| Variable | Default | Purpose |
|---|---|---|
| `GHEGA_AUTH_ENABLED` | `false` | Enable OIDC authentication |
| `GHEGA_OIDC_ISSUER` | — | OIDC provider issuer URL |
| `GHEGA_OIDC_CLIENT_ID` | — | OIDC client ID |
| `GHEGA_OIDC_CLIENT_SECRET` | — | OIDC client secret |
| `GHEGA_OIDC_REDIRECT_URL` | — | OAuth redirect URL |
| `GHEGA_SESSION_SECRET` | — | Secret for signing session cookies |

## Typical Local Development Setup

For local development, no env vars are required. Just run:

```bash
./ghega serve
```

This uses all defaults: HTTP on `:8080`, MLLP on `:2575`, SQLite at `./ghega.db`,
auth disabled.

To test end-to-end message delivery, start a simple HTTP receiver and set:

```bash
GHEGA_DESTINATION_URL=http://localhost:9999/webhook ./ghega serve
```

## Docker

The `Dockerfile` supports the same env vars. Example:

```bash
docker build -t ghega .
docker run -p 8080:8080 -p 2575:2575 \
    -e GHEGA_DATABASE_URL=/data/ghega.db \
    -v ./data:/data \
    ghega
```
