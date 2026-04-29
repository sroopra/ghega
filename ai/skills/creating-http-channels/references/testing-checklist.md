# HTTP Channel Testing Checklist

## Before Testing

- [ ] Channel YAML exists and validates
- [ ] All endpoints use synthetic (non-PHI) data
- [ ] Timeouts and retry policies are explicitly configured
- [ ] Mock server or test stub is available for the external HTTP endpoint

## Functional Tests

- [ ] HTTP source accepts valid requests and returns 2xx
- [ ] HTTP source rejects unsupported HTTP methods with 405
- [ ] HTTP source rejects oversized payloads with 413
- [ ] HTTP destination sends the expected request shape
- [ ] HTTP destination handles 2xx responses correctly
- [ ] HTTP destination retries on 5xx responses
- [ ] HTTP destination does not retry on 4xx responses
- [ ] HTTP destination respects timeout settings

## Idempotency Tests

- [ ] Duplicate requests with the same idempotency key are deduplicated
- [ ] Idempotency records expire after the configured TTL
- [ ] Replays produce the same observable outcome as the original

## Error Path Tests

- [ ] DNS resolution failure is handled gracefully
- [ ] Connection timeout triggers retry logic
- [ ] Malformed response body does not crash the channel
- [ ] Circuit breaker opens after consecutive failures

## Performance and Safety Tests

- [ ] Large payloads do not cause memory issues
- [ ] High request rates are throttled correctly
- [ ] Request and response metadata in logs contain no PHI
- [ ] Secret headers are redacted in test output

## Synthetic Data Rules for HTTP Channels

Use only synthetic data in HTTP payloads:

| Field | Example Value |
|-------|---------------|
| Patient Name | `TESTPATIENT,ONE` |
| MRN | `999999999` |
| Account Number | `TEST-ACCT-001` |
| Date of Birth | `1980-01-01` |
| Correlation ID | `ghega-test-00000000-0000-0000-0000-000000000001` |

## Example Safe JSON Payload

```json
{
  "patientId": "999999999",
  "name": "TESTPATIENT,ONE",
  "dob": "1980-01-01",
  "correlationId": "ghega-test-00000000-0000-0000-0000-000000000001"
}
```
