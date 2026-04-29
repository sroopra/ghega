---
name: creating-db-channels
description: >
  Use when a user wants to create, configure, or modify a database channel in Ghega.
  Use when someone asks about database channel, DB reader, DB writer, SQL query,
  JDBC, or database source/destination connectors.
  Use when the topic involves SQL queries, connection pooling, query configuration,
  or database-driven integration patterns.
license: Apache-2.0
---

# Creating Database Channels

This skill guides the creation and configuration of database source and destination
channels in Ghega for SQL-based integrations and data exchange.

## When to Use

- Building a new database reader (source) or database writer (destination) channel
- Configuring SQL queries for polling or event-driven data extraction
- Setting up connection pooling and database credentials
- Defining query parameters, result mapping, and batch sizes
- Troubleshooting query performance or connection issues

## Key Concepts

- **Database Source**: Polls a table/view or listens for changes via query
- **Database Destination**: Executes inserts, updates, or stored procedures
- **Query Configuration**: SQL statement, parameters, and result mapping
- **Connection Pooling**: Reuses database connections for efficiency
- **PHI-Safe Queries**: Avoid selecting or logging sensitive fields unnecessarily

## Scaffold Guidelines

### Database Source Channel

- Define the query type: polling (timestamp/cursor based) or trigger-based
- For polling: define the table, timestamp column, and poll interval
- For query-based: define the SQL statement and parameter bindings
- Set the batch size to control memory usage
- Map query results to the channel message format

### Database Destination Channel

- Define the target table or stored procedure
- Define the insert/update statement with parameter placeholders
- Set batch insert size when writing multiple rows
- Configure conflict resolution (insert vs. upsert)
- Handle foreign key and constraint violations gracefully

## Safe Defaults

| Setting | Recommended Default | Rationale |
|---------|---------------------|-----------|
| Connection pool min | 2 | Maintains warm connections |
| Connection pool max | 10 | Prevents overwhelming the database |
| Connection timeout | 10s | Prevents hanging connections |
| Query timeout | 30s | Prevents runaway queries |
| Poll interval | 60s | Balances latency with database load |
| Batch size | 100 | Reasonable memory/throughput tradeoff |
| Max row size | 1 MB | Prevents memory issues with large fields |

## PHI-Safe Query Patterns

- Select only the columns required for the integration
- Avoid `SELECT *` in production queries
- Do not log query results or row contents at INFO level
- Use parameterized queries to prevent injection
- Sanitize any dynamic schema or table names through an allowlist

## Safety

Never include real patient data (PHI) in SQL examples or channel configurations.
Always use synthetic test data for examples and fixtures.
Never embed database passwords in channel YAML; reference them through the Ghega
secret provider.

## References

- See [references/db-patterns.md](references/db-patterns.md) for common database
  source and destination patterns, including polling strategies and batch operations.
- See [references/query-safety.md](references/query-safety.md) for PHI-safe query
  guidelines, injection prevention, and audit rules.
