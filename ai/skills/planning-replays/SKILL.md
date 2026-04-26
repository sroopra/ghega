---
name: planning-replays
description: >
  Use when a user wants to plan, design, or execute a message replay in Ghega.
  Use when someone asks about replaying messages, message recovery,
  durable message processing, idempotency, or replay safety.
  Use when the topic involves replay decisions, deduplication,
  or ensuring replays do not cause duplicate side effects.
license: Apache-2.0
---

# Planning Replays

This skill guides the safe planning and execution of message replays in Ghega.

## When to Use

- A message failed and needs to be reprocessed
- A batch of messages needs replay after a system outage
- Designing replay behavior for a new channel
- Ensuring idempotency so replays are safe

## Key Concepts

- **Replay**: Reprocessing a previously received message
- **Idempotency**: Reprocessing the same message produces the same result
- **Deduplication**: Detecting and skipping duplicate messages
- **PayloadRef**: A reference to stored payload data, enabling safe replay without exposing PHI

## Replay Decision Tree

1. **Can the message be safely replayed?**
   - Yes if the destination is idempotent (e.g., upsert to database)
   - No if the destination has side effects that cannot be undone (e.g., sent notification)

2. **Is the original payload still available?**
   - Check the `PayloadRef` storage location

3. **Has this message already been replayed?**
   - Check replay audit log for duplicate detection

## Safety

- Never replay messages containing PHI into non-production environments
- Always verify destination idempotency before replaying
- Log replay decisions as metadata only, never payload content

## References

- See [references/replay-decision-tree.md](references/replay-decision-tree.md) for a detailed decision framework.
