---
name: creating-db-channels
description: >
  Use when a user wants to create, configure, or modify a database channel in Ghega.
  Use when someone asks about database channel, DB reader, DB writer, SQL query,
  or JDBC configuration.
  Use when the topic involves query setup, connection pooling, or PHI-safe query patterns.
license: Apache-2.0
---

# Creating Database Channels

This skill guides the creation and configuration of database source and destination
channels in Ghega. Database channels execute SQL queries to read or write data.

## When to Use

- Building a new database reader or writer channel
- Configuring SQL queries for inbound or outbound data flow
- Setting up connection pooling for database channels
- Ensuring queries are PHI-safe and do not leak sensitive data

## Key Concepts

- **DB Source**: Executes a query and emits each row as a message
- **DB Destination**: Executes an insert or update for each incoming message
- **Connection Pool**: Reuses database connections to reduce overhead
- **Parameterized Queries**: Use placeholders to prevent injection

## DB Source Configuration

```yaml
source:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    SELECT id, status, created_at
    FROM events
    WHERE processed = false
    LIMIT 100
  pollIntervalSeconds: 60
```

## DB Destination Configuration

```yaml
destination:
  type: database
  connectionRef: DB_CONNECTION_STRING
  query: >
    INSERT INTO audit_log (message_id, status, processed_at)
    VALUES (?, ?, ?)
```

## Connection Pooling

Configure pool size based on expected throughput and database capacity.

```yaml
database:
  connectionRef: DB_CONNECTION_STRING
  pool:
    maxOpen: 10
    maxIdle: 5
    maxLifetimeSeconds: 3600
```

## PHI-Safe Query Patterns

- Use column-level selection instead of `SELECT *`
- Avoid logging query results that contain identifiers
- Parameterize all user-supplied values
- Exclude PHI columns from reader queries unless explicitly required

## Safety

Never include real database passwords or connection strings in channel configurations.
Use environment-variable references for all connection parameters.
Always use synthetic test data for examples and fixtures.

## References

- See [references/db-patterns.md](references/db-patterns.md) for common database channel patterns.
- See [references/query-safety.md](references/query-safety.md) for PHI-safe query guidelines.
