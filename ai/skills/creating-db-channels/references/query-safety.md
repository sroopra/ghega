# Query Safety Guidelines

## Parameterized Queries

Always use parameterized queries or prepared statements. Never concatenate
values directly into SQL strings.

Safe:
```sql
SELECT id, status FROM events WHERE type = ?
```

Unsafe:
```sql
SELECT id, status FROM events WHERE type = 'INBOUND'
-- Never interpolate user input here
```

## Column Selection

Select only the columns required for the integration. Avoid `SELECT *` to
reduce data exposure and improve performance.

Preferred:
```sql
SELECT event_id, status, created_at FROM events
```

Avoid:
```sql
SELECT * FROM events
```

## PHI Column Handling

If a table contains PHI, exclude those columns from reader queries unless the
channel is explicitly authorized to access them.

| Column Type | Reader Query | Writer Query |
|-------------|--------------|--------------|
| Identifier (event_id) | Include | Include |
| Status flag | Include | Include |
| Timestamp | Include | Include |
| Patient name | Exclude unless required | Exclude unless required |
| Medical record number | Exclude unless required | Exclude unless required |
| Clinical notes | Exclude unless required | Exclude unless required |

## Query Logging

- Log the query shape (without parameter values) for debugging
- Never log query results that may contain identifiers
- Use redaction for any logged columns that contain sensitive data

## Audit Checklist

- [ ] All queries use parameterized placeholders
- [ ] No `SELECT *` in production channels
- [ ] PHI columns are excluded or explicitly approved
- [ ] Query logs do not contain sensitive values
- [ ] Connection strings use environment-variable references
