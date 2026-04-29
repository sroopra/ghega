# HTTP Patterns

This reference covers common patterns for HTTP source and destination channels.

## HTTP Source Patterns

### Simple REST Listener

A basic HTTP listener that accepts POST requests on a specific path.
Recommended for webhook receivers and simple REST APIs.

Key settings:
- Bind to a non-privileged port (e.g., 8080)
- Restrict methods to those required (usually POST or PUT)
- Set a request timeout to prevent resource exhaustion

### Path-Parameterized Listener

Use when the integration partner sends data in URL path segments.
Validate and sanitize all path parameters before use.

### Authenticated Listener

Use when the source must verify the caller. Ghega supports header-based
authentication. Configure the channel to expect a specific header and validate
it against a secret stored in the environment.

Never hard-code credentials in the channel YAML.

## HTTP Destination Patterns

### Fire-and-Forget Sender

Sends a message to an external endpoint without waiting for business-level
confirmation. Use for logging, notifications, and non-critical callbacks.

Recommended settings:
- Timeout: 10 seconds
- Retries: 0 to 2

### Request-Reply Sender

Sends a message and waits for a response that influences downstream processing.
Use for API calls that return data needed by the channel.

Recommended settings:
- Timeout: 30 seconds
- Retries: 3 with exponential backoff
- Validate HTTP status codes before parsing the body

### Idempotent Retry Sender

Use when the destination endpoint must receive the message exactly once.
Ensure the destination supports idempotency keys or deduplication logic.

Recommended settings:
- Generate an idempotency key from a stable message field
- Retries: 3 with exponential backoff
- Timeout: 30 seconds

## Error Handling

- Treat 4xx responses as client errors; do not retry without fixing the request
- Treat 5xx and network errors as transient; retry with backoff
- Treat 429 (Too Many Requests) as a rate-limit signal; retry after the
  `Retry-After` header value if present

## Security

- Use TLS for all external endpoints
- Validate TLS certificates in production; disable validation only in local tests
- Never log full request or response bodies that may contain PHI
- Redact sensitive headers (authorization, cookies) before logging
