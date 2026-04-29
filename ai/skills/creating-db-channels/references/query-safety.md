# Query Safety Guidelines

This document outlines PHI-safe query practices, injection prevention, and audit
rules for Ghega database channels.

## PHI-Safe Query Practices

### Minimize Data Selection

- Select only the columns required for the integration
- Never use `SELECT *` in production queries
- Avoid selecting PHI columns (e.g., SSN, full name, address) unless absolutely
  necessary for the channel purpose

### Parameterized Queries

Always use parameterized queries or prepared statements:

- Good: `WHERE id = :id`
- Bad: `WHERE id = '` + userInput + `'`

Parameterization prevents SQL injection and improves query plan caching.

### Dynamic Schema and Table Names

When table or schema names must be dynamic:

- Validate against an explicit allowlist
- Do not construct names from unvalidated user input
- Prefer configuration-driven mappings over runtime string concatenation

Example safe pattern:

```
Allowed tables: ["event_outbox", "message_audit", "patient_sync"]
Reject any table name not in the allowlist.
```

## Logging and Audit Rules

### What to Log

Log these at INFO level:

- Query execution time
- Row count returned or affected
- Error codes and messages (without row content)
- Connection metadata (host, database name)

### What Not to Log

Never log these:

- Row contents or column values
- Parameter values that may contain PHI
- Full SQL statements if they embed literal values
- Query results or result sets

### Audit Trail

For channels that process PHI:

- Log the channel name, query identifier, and timestamp
- Log the correlation ID associated with the message
- Do not log the message payload or query parameters

## Synthetic Data for Examples

Use only synthetic data in all SQL examples:

| Field | Example Value |
|-------|---------------|
| Patient Name | `TESTPATIENT,ONE` |
| MRN | `999999999` |
| Account Number | `TEST-ACCT-001` |
| Date of Birth | `1980-01-01` |
| Correlation ID | `ghega-test-00000000-0000-0000-0000-000000000001` |

## Example Safe Query

```sql
SELECT id, message_type, payload, updated_at
FROM event_outbox
WHERE updated_at > :last_poll_time
ORDER BY updated_at ASC
LIMIT :batch_size
```

This query:

- Uses explicit column selection
- Uses named parameters
- Does not select PHI columns
- Limits result set size

## Connection Security

- Use TLS for database connections when the database supports it
- Verify server certificates; do not disable TLS validation in production
- Store credentials in the Ghega secret provider, not in channel YAML
- Use least-privilege database users (read-only for sources, limited tables for destinations)
- Rotate credentials on a regular schedule

## Review Checklist

Before deploying a database channel:

- [ ] Query uses explicit column lists (no `SELECT *`)
- [ ] All parameters are bound, not concatenated
- [ ] Dynamic table/schema names are allowlisted
- [ ] No PHI columns are selected unnecessarily
- [ ] Logs do not contain row contents or parameter values
- [ ] Credentials are referenced through the secret provider
- [ ] Connection uses TLS in production
- [ ] Database user has minimal required privileges
- [ ] Query timeout is configured
- [ ] Batch size limits are configured
