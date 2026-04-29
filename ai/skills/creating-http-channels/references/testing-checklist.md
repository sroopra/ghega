# HTTP Channel Testing Checklist

## Pre-Deployment

- [ ] Endpoint URL is reachable from the Ghega runtime network
- [ ] TLS certificate is valid and trusted (for HTTPS endpoints)
- [ ] Required headers are documented and configured
- [ ] Timeout value matches the slowest expected response time
- [ ] Retry policy does not exceed the downstream rate limit

## Functional Tests

- [ ] Send a synthetic request and verify correct routing
- [ ] Verify HTTP method restrictions (e.g. POST-only paths reject GET)
- [ ] Verify response status codes map correctly to success or error handling
- [ ] Test with malformed payload to confirm error responses

## Resilience Tests

- [ ] Simulate downstream timeout and confirm retry behavior
- [ ] Simulate downstream failure and confirm circuit breaker opens
- [ ] Verify that retry exhaustion produces a clear error log
- [ ] Confirm idempotency key is present for retry-safe destinations

## Security Tests

- [ ] No credentials are hard-coded in the channel YAML
- [ ] Authentication uses environment variables or secret references
- [ ] Webhook endpoints validate incoming signatures or tokens
- [ ] TLS is enforced for all external communication

## Performance Baseline

- [ ] Measure latency under expected load
- [ ] Confirm memory usage stays within allocated limits
- [ ] Verify that concurrent connections do not exhaust connection pools
