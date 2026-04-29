# Database Channel Patterns

## Reader (Inbound)

A database source channel polls a table and emits rows as messages.

```yaml
source:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    SELECT event_id, payload, created_at
    FROM event_queue
    WHERE status = 'pending'
    ORDER BY created_at
    LIMIT 50
  pollIntervalSeconds: 30
  updateQuery: >
    UPDATE event_queue
    SET status = 'processing'
    WHERE event_id = ?
```

## Writer (Outbound)

A database destination channel writes each message to a target table.

```yaml
destination:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    INSERT INTO outbound_log (message_id, status, timestamp)
    VALUES (?, ?, ?)
```

## Polling with Cursor

For high-volume tables, use a cursor column instead of status flags to avoid
locking contention.

```yaml
source:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    SELECT event_id, payload, sequence_num
    FROM event_queue
    WHERE sequence_num > ?
    ORDER BY sequence_num
    LIMIT 100
  cursorColumn: sequence_num
```

## Batch Insert Destination

When throughput is high, batch multiple messages into a single insert.

```yaml
destination:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    INSERT INTO audit_log (message_id, status, timestamp)
    VALUES (?, ?, ?)
  batchSize: 25
  flushIntervalSeconds: 5
```

## Connection Pool Best Practices

| Parameter | Default | Recommended |
|-----------|---------|-------------|
| maxOpen | 10 | Match database connection limit |
| maxIdle | 5 | Half of maxOpen |
| maxLifetimeSeconds | 3600 | Align with database idle timeout |
| connTimeoutSeconds | 10 | Fail fast on network issues |

## Error Handling

- Transient errors (network, lock timeout) should trigger retry
- Constraint violations should route to a dead-letter channel
- Query timeouts should be shorter than the poll interval
