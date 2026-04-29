# HTTP Channel Patterns

This document describes common HTTP source and destination patterns for Ghega channels.

## HTTP Source Patterns

### Webhook Receiver

A webhook receiver accepts HTTP POST requests from an external system.

Key configuration:

- Listen on a dedicated path (e.g., `/webhooks/inbound`)
- Restrict to POST method unless the sender requires others
- Validate incoming signatures or tokens at the edge
- Return a 2xx response quickly to avoid sender retries
- Process the payload asynchronously when possible

### REST API Listener

A REST API listener exposes an endpoint for clients to push data.

Key configuration:

- Define clear path segments and versioning (e.g., `/v1/adt`)
- Accept JSON, XML, or HL7v2 based on use case
- Validate content-type before parsing
- Return structured error responses for malformed input

### HTTP Polling Source (Pull)

In pull mode, the channel periodically fetches from an external HTTP endpoint.

Key configuration:

- Set a polling interval (e.g., 30s, 5m)
- Use conditional requests (If-Modified-Since, ETag) when supported
- Track last-seen timestamps or cursors in channel state
- Handle empty responses gracefully

## HTTP Destination Patterns

### REST API Sender

Sends transformed messages to an external REST API.

Key configuration:

- Construct the URL from channel configuration, not message content
- Set content-type to match the API contract
- Include required headers (correlation IDs, api versions)
- Map response codes to channel outcomes (2xx = success, 4xx = fatal, 5xx = retry)

### Webhook Callback

Sends a callback to a URL provided by the original sender or a downstream system.

Key configuration:

- Validate the callback URL against an allowlist when possible
- Sign outbound payloads so receivers can verify authenticity
- Respect receiver rate limits
- Implement exponential backoff for retries

### HTTP File Upload

Sends a file or binary payload via HTTP multipart.

Key configuration:

- Set multipart boundaries correctly
- Do not include sensitive metadata in filename
- Use pre-signed URLs when available instead of embedding credentials

## Retry and Backoff Patterns

| Failure Type | Recommended Action |
|--------------|--------------------|
| Timeout | Retry with exponential backoff |
| 5xx Server Error | Retry up to configured limit |
| 4xx Client Error | Do not retry; treat as fatal |
| 429 Too Many Requests | Retry after respecting Retry-After header |
| DNS or connection failure | Retry with backoff; alert if persistent |

## Idempotency Patterns

For HTTP destinations, use one of the following idempotency strategies:

1. **Idempotency-Key Header**: Generate a unique key per message and send it in a header
2. **Payload Digest**: Hash the payload and deduplicate on digest match
3. **Correlation ID**: Reuse a stable identifier from the source message

Store idempotency records with a TTL that matches the retry window (default 24h).

## Security Patterns

- Use TLS for all external HTTP traffic
- Verify server certificates; do not disable TLS validation in production
- Rotate bearer tokens through the secret provider, not channel YAML
- Log request metadata (status, duration, correlation ID) without logging payload bytes
