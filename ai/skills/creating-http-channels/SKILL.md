---
name: creating-http-channels
description: >
  Use when a user wants to create, configure, or modify an HTTP channel in Ghega.
  Use when someone asks about HTTP channel, REST API, webhook, HTTP listener,
  or HTTP sender configuration.
  Use when the topic involves timeout policies, retry logic, HTTP endpoints,
  or YAML definitions for HTTP source and destination channels.
license: Apache-2.0
---

# Creating HTTP Channels

This skill guides the creation and configuration of HTTP channels in Ghega.
HTTP channels are used to receive or send messages over REST APIs, webhooks,
and other HTTP-based integrations.

## When to Use

- Building a new HTTP listener (source) or sender (destination) channel
- Configuring timeout and retry policies for HTTP calls
- Setting up webhook receivers or REST API endpoints
- Troubleshooting HTTP connection or response handling issues

## Key Concepts

- **HTTP Source**: Listens on a configured path and port for incoming requests
- **HTTP Destination**: Sends messages to an external HTTP endpoint
- **Timeout Policies**: Prevent hung connections by enforcing read and write deadlines
- **Retry Policies**: Define how transient failures should be retried

## Scaffolding an HTTP Channel

### HTTP Source Channel

An HTTP source channel accepts incoming requests and routes them into the
channel pipeline. Typical configuration includes:

- `port`: the local port to bind
- `path`: the URL path to listen on
- `method`: allowed HTTP methods (GET, POST, PUT, etc.)
- `timeout`: maximum duration to wait for a complete request

### HTTP Destination Channel

An HTTP destination channel sends transformed messages to an external service.
Typical configuration includes:

- `url`: the target endpoint
- `method`: HTTP method to use
- `headers`: static or mapped request headers
- `timeout`: maximum duration to wait for a response

## Timeout and Retry Recommendations

| Scenario | Timeout | Retries | Backoff |
|----------|---------|---------|---------|
| Internal service | 5s | 3 | exponential |
| External API | 30s | 3 | exponential |
| Webhook receiver | 10s | 0 | none |
| Long-polling endpoint | 60s | 1 | fixed |

Default safe values when no specific requirement is known:

- **Timeout**: 30 seconds
- **Retries**: 3 attempts
- **Backoff**: exponential starting at 1 second

## Safety

Never include real API keys, bearer tokens, or basic auth credentials in channel
configurations or examples. Always use placeholder values like `Bearer <TOKEN>`
and inject secrets through environment variables or a secrets manager.

Never include real patient data (PHI) in channel configurations or examples.
Always use synthetic test data for examples and fixtures.

## References

- See [references/http-patterns.md](references/http-patterns.md) for common HTTP
  source and destination patterns.
- See [references/testing-checklist.md](references/testing-checklist.md) for
  validation steps before deploying an HTTP channel.
