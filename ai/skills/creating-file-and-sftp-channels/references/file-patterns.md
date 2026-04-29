# File Patterns

This reference covers common patterns for file source and destination channels.

## File Source Patterns

### Directory Polling Reader

The most common file source pattern. The channel polls a directory at a fixed
interval and picks up files matching a pattern.

Key settings:
- `directory`: must exist and be readable
- `pattern`: use a restrictive glob to avoid picking up partial or temporary files
- `pollInterval`: balance latency against filesystem I/O load
- `archiveDirectory`: must exist and be writable; prevents duplicate processing

### Filtered Extension Reader

Use when the inbound directory contains mixed file types. Restrict the pattern
to the expected extension (e.g., `*.hl7`, `*.json`, `*.xml`).

### Ordered Batch Reader

Use when files must be processed in a specific order. Name files with a
sortable prefix (timestamp or sequence number) and configure the channel to
select the oldest matching file first.

## File Destination Patterns

### Timestamped Writer

Generates output files with a timestamp prefix to ensure uniqueness and
sortability.

Example naming pattern:
- `output-{{timestamp}}.hl7`

### Partitioned Writer

Splits output into subdirectories based on a message attribute (e.g., date,
source system, message type). This simplifies downstream archival and auditing.

Example directory pattern:
- `outbound/{{messageType}}/{{date}}/`

### Append Writer

Appends multiple messages to a single file. Use only when the downstream
consumer expects concatenated messages and handles boundary detection.

## Error Handling

- Permission errors on read: move file to error directory and alert
- Permission errors on write: retry once, then queue for manual review
- Partial files: exclude files with temporary extensions (`.tmp`, `.part`)
- Empty files: log and archive; do not treat as a processing failure

## Security

- Run the channel process with the least-privilege user that can read the
  source directory and write the destination directories
- Restrict directory permissions so only the channel owner and authorized
  administrators have access
- Scan inbound files for unexpected content types before processing
- Do not execute or interpret file names as commands
