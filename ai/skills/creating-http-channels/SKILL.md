---
name: creating-http-channels
description: >
  Use when a user wants to create, configure, or modify an HTTP channel in Ghega.
  Use when someone asks about HTTP channel, REST API, webhook, HTTP listener,
  HTTP sender, or HTTP source/destination connectors.
  Use when the topic involves HTTP timeouts, retry policies, webhook receivers,
  REST endpoints, or HTTP-based integration patterns.
license: Apache-2.0
---

# Creating HTTP Channels

This skill guides the creation and configuration of HTTP source and destination
channels in Ghega for REST APIs, webhooks, and HTTP-based integrations.

## When to Use

- Building a new HTTP listener (source) or HTTP sender (destination) channel
- Configuring webhook receivers
- Setting up REST API integrations
- Defining timeout, retry, and backpressure policies for HTTP connectors
- Troubleshooting HTTP connection or response handling issues

## Key Concepts

- **HTTP Source**: Listens for incoming HTTP requests (webhooks, REST pushes)
- **HTTP Destination**: Sends outbound HTTP requests (REST APIs, callbacks)
- **Timeout Policy**: Maximum wait time for request/response round-trips
- **Retry Policy**: Backoff strategy for transient failures (5xx, timeouts)
- **Idempotency Key**: Header or payload field used to deduplicate requests

## Scaffold Guidelines

### HTTP Source Channel

- Define the listen path and permitted HTTP methods
- Configure authentication/authorization if required
- Set request size limits to prevent abuse
- Specify expected content-type and parser settings

### HTTP Destination Channel

- Define the base URL and endpoint path
- Configure headers, including content-type and any auth headers
- Set connection and read timeouts explicitly
- Configure retry count, backoff interval, and circuit breaker thresholds

## Safe Defaults

| Setting | Recommended Default | Rationale |
|---------|---------------------|-----------|
| Connection timeout | 10s | Prevents hanging connections |
| Read timeout | 30s | Allows for slow endpoints without indefinite wait |
| Retry count | 3 | Balances reliability with load |
| Backoff base | 1s exponential | Reduces thundering herd |
| Max request size | 10 MB | Prevents memory exhaustion |
| Idempotency TTL | 24h | Matches typical business window |

## Safety

Never include real patient data (PHI) in HTTP payload examples or channel configurations.
Always use synthetic test data for examples, fixtures, and webhook payloads.
Avoid embedding API keys, tokens, or passwords in channel YAML; reference secrets
through the Ghega secret provider instead.

## References

- See [references/http-patterns.md](references/http-patterns.md) for common HTTP
  source and destination patterns, including webhook verification and callback handling.
- See [references/testing-checklist.md](references/testing-checklist.md) for
  HTTP channel test generation guidelines.
