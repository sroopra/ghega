---
name: diagnosing-operations
description: >
  Use when a channel is down, a message is stuck, the queue is full, a listener is
  not responding, the error rate is high, or an alert has fired. Use when someone
  asks about message metadata, channel status, recent alerts, or tracing message flow.
license: Apache-2.0
---

# Diagnosing Operations

This skill guides operational diagnosis of Ghega channels and message flow without accessing payload bytes.

## When to Use

- A channel is reported down or unresponsive
- Messages appear stuck in a queue
- Queue depth exceeds normal thresholds
- An alert fired for high error rate or latency
- A listener is not accepting connections

## Key Concepts

- **Message Metadata**: Routing info, timestamps, status, and channel IDs without payload content
- **Channel Status**: Current state such as running, stopped, or error
- **Alert**: A threshold-based notification from the monitoring system
- **Message Flow**: The path a message takes from source through transformations to destinations

## Diagnostic Workflow

### 1. Inspect Message Metadata

- Query by message ID for status, timestamps, and routing decisions
- Check retry count and last error classification
- Verify idempotency key and deduplication result
- Do not access or log payload bytes

### 2. Check Channel Status

- Verify the channel is in the expected state (running, paused, stopped)
- Check source connector health and connection counts
- Check destination connector health and response times
- Review recent start, stop, or restart events

### 3. Review Recent Alerts

- List alerts for the channel in the last hour and last day
- Correlate alert timing with deployments or infrastructure changes
- Prioritize alerts by severity and business impact

### 4. Trace Message Flow

- Follow the message from source receipt through each route
- Note transformation outcomes and any filter decisions
- Check destination attempts, success/failure status, and retry schedule
- Identify the first point of failure or unexpected behavior

## Common Scenarios

### Queue Full

- Check if downstream destinations are slow or failing
- Verify channel backpressure settings
- Consider increasing concurrency limits or scaling consumers
- Do not purge queues without approval and audit logging

### Listener Not Responding

- Verify network policy allows inbound traffic on the listener port
- Check listener process health and restart if needed
- Review connection logs for rejected or dropped connections
- Validate TLS certificate expiry if TLS is enabled

### High Error Rate

- Identify the most frequent error classification
- Check for recent channel configuration changes
- Verify destination endpoints are reachable and healthy
- Review mapping or transformation changes that may be causing rejects

## Safety

- Never access payload bytes unless explicitly authorized and audited
- Log only metadata (message IDs, timestamps, status, channel IDs)
- Do not share message content in incident channels
- Follow the principle of least privilege when investigating

## References

- See [references/operations-checklist.md](references/operations-checklist.md) for a step-by-step operational checklist.
