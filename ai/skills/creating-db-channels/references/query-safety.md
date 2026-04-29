# Query Safety

This reference provides detailed guidance on writing secure, PHI-safe SQL
queries for database channels.

## Parameterized Queries

Always use parameterized queries or prepared statements. Never concatenate
message fields, file names, or external input directly into SQL strings.

Unsafe example (do not use):
- `INSERT INTO patients (name) VALUES (' + msgName + ')`

Safe example:
- `INSERT INTO patients (name) VALUES (?)` with the value bound as a parameter

## Minimizing Data Exposure

### Column Selection

Select only the columns required by the channel. Avoid `SELECT *` because it
increases memory use, network traffic, and the risk of exposing unexpected
sensitive columns.

### Row Limits

Always cap query results unless the channel explicitly requires a full table
scan. Use database-specific limit clauses:

- PostgreSQL / MySQL: `LIMIT`
- SQL Server: `TOP`
- Oracle: `FETCH FIRST` or `ROWNUM`

### Filtering in the Database

Push filters into the query rather than fetching all rows and filtering in the
channel. This reduces data movement and exposure.

## Logging and Monitoring

### Query Logging

When logging queries for debugging:

- Log the query template (with parameter placeholders) rather than the executed
  query with values substituted
- If parameter values must be logged, redact PHI fields before writing to logs
- Never log connection strings or credentials

### Audit Requirements

Ensure database access audit logs are enabled. The audit trail should capture:

- The service account or application name running the query
- Timestamp and source IP address
- The query template or hash
- Number of rows accessed

## Access Control

### Least Privilege

The database user for a channel should have only the permissions it needs:

- Read-only sources: grant SELECT only on required tables or views
- Write-only destinations: grant INSERT (and UPDATE for upserts) only on target
  tables
- Deny access to unrelated schemas and tables

### View-Based Access

When possible, expose data through views rather than base tables. Views can:

- Filter rows to only those relevant to the channel
- Exclude sensitive columns
- Apply row-level security policies

## Example Safe Query Patterns

### Reading with Limits

```
SELECT patient_id, admission_date, department
FROM admissions
WHERE processed = 0
ORDER BY admission_date
LIMIT 100;
```

### Writing with Explicit Columns

```
INSERT INTO audit_log (channel_name, message_id, processed_at, status)
VALUES (?, ?, ?, ?);
```

### Updating with a Bookmark

```
UPDATE events
SET processed = 1, processed_at = ?
WHERE id IN (...);
```

## Sanitizing Test Data

All example queries in documentation and test fixtures must use synthetic data.
Replace real identifiers with placeholders such as:

- Patient IDs: `SYNTH-0001`, `SYNTH-0002`
- Names: `TestPatient`, `SyntheticName`
- Dates: `2024-01-01`, `2024-12-31`

Never use real names, addresses, phone numbers, or medical record numbers in
examples.
