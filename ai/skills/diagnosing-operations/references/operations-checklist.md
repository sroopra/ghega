# Operations Checklist

## Initial Response (First 5 Minutes)

- [ ] Confirm the affected channel ID and environment
- [ ] Check channel status (running, stopped, error)
- [ ] Review active alerts for the channel and correlated infrastructure
- [ ] Identify whether the issue is isolated or widespread

## Message Inspection (Metadata Only)

- [ ] Query recent failed messages by error classification
- [ ] Check message retry counts and dead-letter status
- [ ] Verify idempotency key collisions or deduplication rejections
- [ ] Confirm message timestamps align with reported incident window
- [ ] Do not access or log payload bytes

## Channel Health

- [ ] Verify source connector is listening and accepting connections
- [ ] Verify destination connectors are reachable and returning expected status codes
- [ ] Review channel start, stop, and restart events
- [ ] Check for recent deployments or configuration changes

## Infrastructure and Resources

- [ ] Check CPU, memory, and disk utilization for channel runners
- [ ] Verify network policies allow required ingress and egress
- [ ] Review queue depth and consumer lag
- [ ] Check database connection pool health if applicable

## Mapping and Transformations

- [ ] Review recent changes to CEL expressions or field mappings
- [ ] Verify referenced fields exist in the expected schema
- [ ] Check for type mismatches or null-handling gaps

## Rollback or Mitigation

- [ ] If a recent deployment correlates with the incident, consider rollback
- [ ] Document the decision and expected impact before acting
- [ ] Monitor metrics for recovery after mitigation

## Post-Incident

- [ ] Update incident timeline with metadata-only observations
- [ ] Identify root cause and preventive measures
- [ ] Schedule follow-up review if needed

## Example Status Query (Safe)

```bash
ghega channel status --channel-id adt-inbound --env production
```

Expected output includes channel state, source health, destination health, and recent error count. No payload content is returned.
