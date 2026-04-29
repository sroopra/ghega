---
name: creating-db-channels
description: >
  Use when a user wants to create, configure, or modify a database channel in
  Ghega. Use when someone asks about database channel, DB reader, DB writer,
  SQL query, or JDBC configuration.
  Use when the topic involves query configuration, connection pooling,
  PHI-safe query patterns, or YAML definitions for database source and
  destination channels.
license: Apache-2.0
---

# Creating Database Channels

This skill guides the creation and configuration of database channels in Ghega.
Database channels read from or write to relational databases using SQL queries
and standard database drivers.

## When to Use

- Building a database reader (source) that polls a table or runs a query
- Building a database writer (destination) that inserts or updates records
- Configuring connection pools and query timeouts
- Designing PHI-safe queries that avoid exposing sensitive data in logs
- Setting up polling intervals for database source channels

## Key Concepts

- **DB Source**: Executes a query at a polling interval and emits each row as a
  message
- **DB Destination**: Executes an insert, update, or upsert for each incoming
  message
- **Connection Pool**: Reuses database connections to reduce overhead
- **Query Parameterization**: Uses placeholders instead of string concatenation
  to prevent SQL injection

## Scaffolding a Database Channel

### DB Source Channel

A database source channel runs a SQL query on a schedule and routes results
into the channel pipeline. Typical configuration includes:

- `connectionString`: reference to a DSN configured in the environment
- `query`: the SQL select statement
- `pollInterval`: duration between query executions
- `queryTimeout`: maximum time to wait for query results
- `columnMapping`: how result columns map to message fields

### DB Destination Channel

A database destination channel writes incoming messages to a database.
Typical configuration includes:

- `connectionString`: reference to a DSN configured in the environment
- `statement`: the insert, update, or upsert SQL statement
- `parameterMapping`: how message fields map to query parameters
- `batchSize`: number of rows to batch before executing (if supported)

## Connection Pool Recommendations

| Pool Setting | Recommended Value | Notes |
|--------------|-------------------|-------|
| Max open connections | 10 | Increase only if query latency is high |
| Max idle connections | 5 | Reduces connection churn |
| Connection max lifetime | 30 minutes | Prevents stale connection issues |
| Connection max idle time | 10 minutes | Closes unused connections gently |

## Query Timeout Recommendations

| Query Type | Timeout | Notes |
|------------|---------|-------|
| Simple lookup | 5s | Indexed single-row queries |
| Batch read | 30s | May scan many rows |
| Write/Update | 10s | Should be fast with proper indexing |
| Complex report | 120s | Run during off-peak hours |

## PHI-Safe Query Patterns

- **Never select `*` in production**: Explicitly list only the columns needed
- **Avoid logging query parameters that contain identifiers**: Redact MRNs,
  account numbers, and names before logging
- **Use parameterized queries**: Never concatenate user input or message fields
  directly into SQL strings
- **Limit result sets**: Use `TOP`, `LIMIT`, or `ROWNUM` to bound query results
- **Audit access**: Ensure database audit logs capture channel queries

## Polling Interval Recommendations

| Use Case | Poll Interval | Notes |
|----------|---------------|-------|
| Event table polling | 10s | Detect new rows quickly |
| Status table polling | 1m | Balance freshness and load |
| Nightly summary | 15m or cron | Scheduled batch extraction |
| Archive table | 1h | Low-priority historical data |

## Safety

Never include real database passwords, connection strings, or credentials in
channel configurations or examples. Always use placeholder values and inject
credentials through environment variables or a secrets manager.

Never include real patient data (PHI) in SQL queries, query results, or examples.
Always use synthetic test data for examples and fixtures.

## References

- See [references/db-patterns.md](references/db-patterns.md) for common database
  source and destination patterns.
- See [references/query-safety.md](references/query-safety.md) for detailed
  guidance on writing secure, PHI-safe SQL queries.
