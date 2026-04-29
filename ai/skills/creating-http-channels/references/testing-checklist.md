# HTTP Channel Testing Checklist

Use this checklist before deploying an HTTP channel to production.

## Functional Validation

- [ ] Source channel accepts requests on the configured port and path
- [ ] Destination channel successfully delivers messages to the target URL
- [ ] Correct HTTP method is used for every destination call
- [ ] Request headers are set as specified in the channel configuration
- [ ] Response status codes are handled according to the error-handling policy

## Timeout and Retry Validation

- [ ] Timeout values are configured and match the target SLA
- [ ] Retry policy is documented and tested with simulated failures
- [ ] Exponential backoff does not exceed acceptable total latency
- [ ] Circuit breaker or rate-limiting behavior is verified if applicable

## Security Validation

- [ ] TLS is enabled for all external endpoints
- [ ] Authentication headers or tokens are injected from environment variables
- [ ] No secrets are hard-coded in channel YAML or test fixtures
- [ ] Request and response logs do not contain PHI or credentials

## Performance Validation

- [ ] Channel handles expected peak throughput without connection exhaustion
- [ ] Connection pooling is configured for high-volume destinations
- [ ] Payload sizes are within limits enforced by the destination API

## Synthetic Test Data

- [ ] All test messages use synthetic data (no real patient names or identifiers)
- [ ] Edge cases are covered (empty body, oversized body, invalid JSON or XML)
