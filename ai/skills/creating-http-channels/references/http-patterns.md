# HTTP Channel Patterns

## Listener (Inbound)

An HTTP source channel listens on a local path and dispatches incoming requests.

```yaml
source:
  type: http
  path: /api/v1/events
  method: POST
  headers:
    - Content-Type: application/json
```

## Sender (Outbound)

An HTTP destination channel forwards messages to an external endpoint.

```yaml
destination:
  type: http
  url: https://api.example.com/v1/events
  method: POST
  headers:
    - Content-Type: application/json
    - Accept: application/json
```

## Webhook Receiver

Webhooks are HTTP sources with minimal validation. Always verify the payload
signature or token when available.

```yaml
source:
  type: http
  path: /webhooks/inbound
  method: POST
```

## Retry Configuration

Use exponential backoff to avoid overwhelming the target system.

```yaml
destination:
  type: http
  url: https://api.example.com/v1/events
  retry:
    maxAttempts: 3
    backoff: exponential
    initialDelayMs: 1000
```

## Timeout Configuration

```yaml
destination:
  type: http
  url: https://api.example.com/v1/events
  timeoutSeconds: 30
```

## Safe Defaults Summary

| Parameter | Default | Recommended Range |
|-----------|---------|-------------------|
| timeoutSeconds | 30 | 10 - 120 |
| maxAttempts | 3 | 1 - 5 |
| initialDelayMs | 1000 | 500 - 5000 |
| circuitBreakerThreshold | 5 | 3 - 10 |
