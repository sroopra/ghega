---
name: creating-http-channels
description: >
  Use when a user wants to create, configure, or modify an HTTP channel in Ghega.
  Use when someone asks about HTTP channel, REST API, webhook, HTTP listener,
  or HTTP sender configuration.
  Use when the topic involves timeout policies, retry logic, or HTTP endpoint setup.
license: Apache-2.0
---

# Creating HTTP Channels

This skill guides the creation and configuration of HTTP source and destination
channels in Ghega. HTTP channels are used for REST APIs, webhooks, and general
HTTP-based integration.

## When to Use

- Building a new HTTP listener or sender channel
- Configuring webhook receivers or REST API clients
- Setting timeout and retry policies for HTTP calls
- Troubleshooting HTTP connection or response issues

## Key Concepts

- **HTTP Source**: Listens on a configured path and method for incoming requests
- **HTTP Destination**: Sends outbound requests to a remote endpoint
- **Timeout Policy**: Maximum wait time for a response before failing
- **Retry Policy**: Rules for re-attempting failed requests

## Scaffold an HTTP Channel

1. Choose source or destination type (`http`)
2. Configure the endpoint URL or listen path
3. Set method (GET, POST, PUT, DELETE, PATCH)
4. Define headers and content-type
5. Apply timeout and retry policies

## Recommended Timeout and Retry Policies

- **Timeout**: Start with 30 seconds for typical APIs; use 10 seconds for health checks
- **Retries**: 3 retries with exponential backoff (1s, 2s, 4s)
- **Circuit breaker**: Enable after 5 consecutive failures with a 60s reset window

## Safety

Never include real credentials, API keys, or bearer tokens in channel configurations.
Use environment-variable references or a secret manager for authentication.
Always use synthetic test data for examples and fixtures.

## References

- See [references/http-patterns.md](references/http-patterns.md) for common HTTP channel patterns.
- See [references/testing-checklist.md](references/testing-checklist.md) for validation steps.
