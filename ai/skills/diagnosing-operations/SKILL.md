---
name: diagnosing-operations
description: >
  Use when a channel is down, a message appears stuck, a queue is full, or a
  listener is not responding. Use when there is a high error rate, an alert fired,
  or someone asks about message flow, channel status, or operations diagnostics.
  Use when the topic involves inspecting message metadata, reviewing recent alerts,
  or tracing message flow without accessing payload bytes.
license: Apache-2.0
---

# Diagnosing Operations Issues

This skill guides operational diagnosis of running Ghega channels without accessing
sensitive payload bytes.

## When to Use

- A channel health check reports down or degraded
- Messages are accumulating in a queue without being processed
- An alert fired for high error rate or latency
- A listener port is not accepting connections
- Message flow needs to be traced for debugging

## Inspecting Message Metadata

Message metadata contains operational signals without exposing payload content:

| Field | What It Reveals |
|-------|-----------------|
| `messageId` | Unique identifier for correlation across logs |
| `timestamp` | When the message entered the system |
| `channelId` | Which channel processed the message |
| `sourceType` | Originating connector type (MLLP, HTTP, file, DB) |
| `destinationType` | Target connector type |
| `status` | Current state: received, processing, sent, error, dead-letter |
| `retryCount` | How many times delivery has been attempted |
| `errorCode` | Classification of the last failure, if any |
| `durationMs` | Time spent in the channel so far |

Use these fields to identify patterns:

- All errors from the same `channelId` point to a channel-specific issue
- High `retryCount` across many messages points to a destination problem
- Sudden spikes in `durationMs` point to resource contention or downstream slowness

## Checking Channel Status

Channel status provides a high-level health signal:

- **Healthy**: channel is running and processing messages within SLO
- **Degraded**: channel is running but latency or error rate is elevated
- **Down**: channel is not running or failing health checks
- **Paused**: channel was intentionally stopped

When checking status:

1. Look at the channel process or pod status
2. Review recent restart count or crash loop events
3. Check resource utilization (CPU, memory, open connections)
4. Verify the channel configuration has not drifted from the deployed version

## Reviewing Recent Alerts

Alerts are triggered by threshold rules. When an alert fires:

1. Note the alert name, severity, and threshold that was breached
2. Check the time window — is this a spike or a sustained trend?
3. Correlate with deployments or configuration changes in the same window
4. Check related channels that share the same destination or infrastructure

Common alert patterns:

| Alert | Likely Cause | First Action |
|-------|--------------|--------------|
| High error rate | Destination unreachable or validation failing | Check destination health and recent changes |
| Queue depth high | Consumer slower than producer | Scale consumers or review batch sizes |
| Latency p99 high | Downstream slowdown or resource limit | Review resource metrics and downstream SLOs |
| Listener connect failures | Port conflict, certificate expiry, network policy | Verify listener config and certificates |

## Tracing Message Flow

To trace a message through the system without reading its payload:

1. Start with the `messageId` from the source system or log line
2. Follow the message through each stage using its `messageId`:
   - Source receive log
   - Channel processing log
   - Destination send or error log
3. Note timestamps at each stage to identify where time is spent
4. Look for `errorCode` transitions that indicate where the failure occurred

If the messageId is not available, use correlation fields such as:

- Source connector name + timestamp range
- Destination connector name + batch identifier
- Alert trigger metadata that includes channel and time window

## Queue Diagnostics

When a queue is full or messages are stuck:

1. Check the current queue depth and the configured maximum
2. Review the consumer throughput over the last hour
3. Identify whether the slowdown is in the channel logic or the destination
4. Check for dead-letter queue growth, which indicates repeated failures
5. Consider temporarily pausing the source to prevent overflow while investigating

## Safety

- Never access or log message payload bytes for operational diagnosis
- Use metadata fields and correlation identifiers instead
- If payload inspection is required, route the request through the proper
  authorization and audit workflow
- Always use synthetic data when creating reproduction scenarios

## References

- See [references/operations-checklist.md](references/operations-checklist.md)
  for a step-by-step checklist when responding to operational incidents.
