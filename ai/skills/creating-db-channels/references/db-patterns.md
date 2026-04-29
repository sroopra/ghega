# Database Patterns

This reference covers common patterns for database source and destination
channels.

## DB Source Patterns

### Event Table Polling

Poll a dedicated event or queue table for new rows. After processing, mark the
row as complete or delete it.

Key settings:
- `query`: select unprocessed rows ordered by creation time
- `pollInterval`: frequent enough to meet latency requirements
- `updateAfterRead`: mark rows as processed to avoid duplicates

Example query structure:
- `SELECT id, payload, created_at FROM events WHERE processed = 0 ORDER BY created_at LIMIT 100`

### Change-Data-Capture (CDC) Simulation

When native CDC is unavailable, poll a timestamp or sequence column to detect
new or changed rows.

Key settings:
- `query`: select rows where `updated_at` is after the last poll timestamp
- `bookmarkColumn`: the column used to track progress
- `bookmarkStorage`: where the last-seen value is stored between polls

### Lookup Table Query

Use when the channel needs to enrich messages with reference data. Run a
parameterized lookup query for each incoming message.

Key settings:
- `query`: single-row select with a parameterized key
- `queryTimeout`: short timeout because the query should be fast
- `cacheResults`: optional in-memory cache for frequently looked-up values

## DB Destination Patterns

### Single-Row Insert

Insert one record per incoming message. Simple and reliable, but higher
overhead for high-volume channels.

Key settings:
- `statement`: parameterized INSERT with explicit column list
- `parameterMapping`: map message fields to statement parameters

### Batch Upsert

Accumulate multiple messages and execute a single batch upsert. Reduces
round-trips but increases memory use and complexity.

Key settings:
- `batchSize`: number of rows per batch (e.g., 100)
- `flushInterval`: maximum time to wait before flushing a partial batch
- `conflictResolution`: ON CONFLICT or MERGE logic for duplicate keys

### Audit Log Writer

Write an immutable audit record for every message processed. Useful for
compliance and debugging.

Key settings:
- `statement`: INSERT into an audit table with timestamp and channel name
- `idempotent`: use a deterministic UUID or hash to prevent duplicate audit rows

## Error Handling

- Connection errors: retry with exponential backoff; alert after repeated failure
- Query timeout: log the query hash (not full text if it may contain PHI);
  investigate slow query execution plans
- Constraint violations: route to a dead-letter table for manual review
- Deadlock: retry the transaction once before escalating

## Performance

- Create indexes on columns used in WHERE clauses, JOIN conditions, and ORDER BY
- Avoid SELECT *; fetch only the columns the channel needs
- Use prepared statements for repeated queries
- Monitor connection pool saturation and adjust pool size if needed
