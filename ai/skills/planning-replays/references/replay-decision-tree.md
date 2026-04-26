# Replay Decision Tree

## Step 1: Identify the Message

- Locate the message by message ID in the Ghega audit log
- Verify the `PayloadRef` points to existing stored data
- Confirm the message type and channel configuration

## Step 2: Assess Destination Idempotency

| Destination Type | Idempotent? | Replay Safe? |
|------------------|-------------|--------------|
| Database upsert  | Yes         | Yes          |
| File write       | No          | No (overwrite risk) |
| HTTP POST        | Depends     | Check endpoint contract |
| MLLP send        | No          | No (duplicate send) |
| Notification/email | No        | No (duplicate alert) |

## Step 3: Check for Prior Replays

- Query replay audit log by original message ID
- If replayed before, evaluate whether another replay is justified
- Document the reason for each replay attempt

## Step 4: Plan the Replay

- Select the target environment (must match original unless approved)
- Verify channel configuration has not changed since original receipt
- Confirm any mapping or transformation changes are compatible

## Step 5: Execute and Verify

- Trigger replay via Ghega replay API or CLI
- Monitor destination for expected outcome
- Verify no duplicate records or side effects occurred
- Log replay result as metadata (message ID, timestamp, outcome)

## Abort Conditions

Abort replay planning if any of the following are true:

- [ ] Payload data is no longer available
- [ ] Destination is known to be non-idempotent and no mitigation exists
- [ ] Message has already been replayed more than twice without resolution
- [ ] Target environment does not match source environment (without explicit approval)
- [ ] PHI would be exposed to an unauthorized environment

## Example Replay Log Entry (Safe)

```yaml
replay_id: replay-20240101-001
original_message_id: msg-abc-123
channel_id: adt-inbound
timestamp: 2024-01-01T12:00:00Z
outcome: success
notes: Replay executed after destination system recovery.
# No payload content logged. Metadata only.
```
