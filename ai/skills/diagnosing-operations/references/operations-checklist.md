# Operations Checklist

This checklist provides a structured sequence of steps for diagnosing operational issues.

## Initial Response

- [ ] Confirm the alert or symptom (channel down, queue full, high error rate)
- [ ] Identify the affected channel or channels
- [ ] Note the time the issue started
- [ ] Check for recent deployments or configuration changes in the same window
- [ ] Determine whether the issue is isolated or systemic

## Channel Health

- [ ] Check channel status (healthy, degraded, down, paused)
- [ ] Review restart count and crash loop events
- [ ] Verify resource utilization (CPU, memory, file descriptors, connections)
- [ ] Confirm the deployed configuration matches the intended version
- [ ] Check listener port accessibility and certificate validity

## Message Flow

- [ ] Identify a representative `messageId` or correlation window
- [ ] Trace the message through source, processing, and destination logs
- [ ] Note timestamps at each stage to find bottlenecks
- [ ] Review `errorCode` and `retryCount` for failed messages
- [ ] Check dead-letter queue depth and recent entries

## Queue and Throughput

- [ ] Measure current queue depth against maximum configured
- [ ] Compare producer rate to consumer rate
- [ ] Review batch size and polling interval settings
- [ ] Check for downstream destination latency or timeouts
- [ ] Verify autoscaling rules are configured and triggering correctly

## Correlation and Escalation

- [ ] Correlate the issue with any infrastructure alerts (network, storage, compute)
- [ ] Check related channels that share destinations or infrastructure
- [ ] Document findings in the incident record
- [ ] If root cause is unclear after the checklist, escalate with:
  - Affected channel names
  - Time window
  - Relevant log excerpts (metadata only, no payload)
  - Recent changes
  - Actions already taken
