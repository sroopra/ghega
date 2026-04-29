# Database Channel Patterns

This document describes common database source and destination patterns for Ghega channels.

## Database Source Patterns

### Timestamp-Based Polling

Polls a table for rows where a timestamp column is newer than the last poll.

Key configuration:

- Table name and timestamp column (e.g., `updated_at`)
- Poll interval
- Ordering to ensure consistent paging
- Batch size to limit rows per poll

Example query shape:

```sql
SELECT id, message_type, payload, updated_at
FROM event_outbox
WHERE updated_at > :last_poll_time
ORDER BY updated_at ASC
LIMIT :batch_size
```

Important: Use a strictly monotonic timestamp or a composite cursor (timestamp + id)
to avoid missing rows that share the same timestamp.

### Auto-Increment ID Polling

Polls a table for rows with an ID greater than the last seen ID.

Key configuration:

- Table name and ID column
- Last seen ID cursor (stored in channel state)
- Batch size

Example query shape:

```sql
SELECT id, message_type, payload
FROM event_outbox
WHERE id > :last_id
ORDER BY id ASC
LIMIT :batch_size
```

This pattern is simpler than timestamp polling but requires an auto-incrementing
primary key.

### Change Data Capture (CDC) Simulation

Some databases support CDC or logical replication. Where native CDC is unavailable,
a polling query can simulate it.

Key configuration:

- Dedicated outbox or changelog table
- Clear status column (e.g., `pending`, `processed`)
- Atomic update to mark rows as processed

Example workflow:

1. Select rows with status `pending`
2. Process each row
3. Update status to `processed` in a transaction

### Query-Based Source

Executes a parameterized query on each poll cycle.

Key configuration:

- SQL statement with named or positional parameters
- Parameter bindings from channel state or configuration
- Result column mapping to message fields

## Database Destination Patterns

### Single Row Insert

Inserts one row per message.

Key configuration:

- Target table and column list
- Parameter placeholders matching message fields
- Error handling for duplicate keys or constraint violations

Example query shape:

```sql
INSERT INTO message_audit (correlation_id, message_type, received_at)
VALUES (:correlation_id, :message_type, :received_at)
```

### Batch Insert

Accumulates multiple messages and inserts them in a single batch.

Key configuration:

- Batch size or time window
- Multi-row VALUES clause or bulk load API
- Error handling for partial batch failures

Example query shape:

```sql
INSERT INTO message_audit (correlation_id, message_type, received_at)
VALUES
  (:correlation_id_1, :message_type_1, :received_at_1),
  (:correlation_id_2, :message_type_2, :received_at_2)
```

### Upsert (Insert or Update)

Inserts a row if it does not exist, otherwise updates it.

Key configuration:

- Unique key columns for conflict detection
- Update columns and their new values

Example query shape (PostgreSQL-style):

```sql
INSERT INTO patient_sync (mrn, name, last_updated)
VALUES (:mrn, :name, :last_updated)
ON CONFLICT (mrn) DO UPDATE SET
  name = EXCLUDED.name,
  last_updated = EXCLUDED.last_updated
```

### Stored Procedure Call

Invokes a stored procedure for complex logic.

Key configuration:

- Procedure name and parameter list
- Input parameter bindings
- Output parameter or result set handling

Prefer stored procedures only when necessary; plain SQL is easier to audit and test.

## Connection Pooling Patterns

| Pool Setting | Typical Value | Notes |
|--------------|---------------|-------|
| Min idle | 2 | Keeps connections warm |
| Max active | 10 | Limits concurrent load on DB |
| Max wait | 10s | Fails fast when pool exhausted |
| Validation query | `SELECT 1` | Lightweight health check |
| Connection TTL | 30m | Recycles connections periodically |

## Paging and Batch Processing

- Always use `LIMIT`/`OFFSET` or cursor-based paging for large result sets
- Avoid loading entire tables into memory
- Process and acknowledge each batch before fetching the next
- Store the last cursor value in channel state for crash recovery

## Error Handling

| Scenario | Recommended Action |
|----------|--------------------|
| Connection timeout | Retry with backoff; alert if persistent |
| Query timeout | Kill query, log, and alert |
| Duplicate key violation | Treat as fatal or upsert based on design |
| Constraint violation | Log details and route to error handler |
| Null value in required column | Log and route to error handler |
| Database unavailable | Pause polling and retry with backoff |
