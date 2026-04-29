# File Channel Patterns

This document describes common local file system source and destination patterns
for Ghega channels.

## File Source Patterns

### Directory Polling Reader

Polls a local directory for new files matching a pattern.

Key configuration:

- Source directory path (absolute or relative to channel working directory)
- File filter pattern (glob or regex)
- Polling interval
- Action after read: archive, delete, or rename

Example workflow:

1. Poll `/incoming/` every 30 seconds
2. Process files matching `*.hl7`
3. On success, move to `/archive/`
4. On failure, leave in place or move to `/error/`

### File Watcher (Event-Driven)

Uses filesystem events instead of polling where supported.

Key configuration:

- Watch directory path
- File filter pattern
- Debounce interval to handle rapid successive events

Note: Event-driven sources may still fall back to polling on platforms where
inotify or equivalent is unavailable.

## File Destination Patterns

### Single File Writer

Writes each message to a separate file.

Key configuration:

- Output directory
- Filename template with unique components
- Write mode: create, append, or overwrite
- Encoding: UTF-8 by default

### Batch File Writer

Accumulates multiple messages and writes them to a single batch file.

Key configuration:

- Batch size or time window
- Batch filename template
- Delimiter between records (newline, segment separator, etc.)
- Flush and rotate policy

### Atomic Write Pattern

To prevent partial files from being read by downstream consumers:

1. Write to a temporary filename (e.g., `.tmp-{uuid}`)
2. Flush and close the file
3. Atomically rename to the final filename

## Path Configuration Guidelines

- Use absolute paths in production to avoid working-directory ambiguity
- Ensure the channel runtime user has read/write permissions
- Separate incoming, processing, archive, and error directories
- Validate path components to prevent directory traversal

## File Permissions

- Create files with restrictive permissions (e.g., 0640) when they may contain
  message data
- Avoid world-readable directories for sensitive integrations
- Set umask appropriately in the runtime environment

## Error Handling

| Scenario | Recommended Action |
|----------|--------------------|
| Permission denied | Log error and alert; do not retry indefinitely |
| Disk full | Pause polling and alert |
| File locked by another process | Skip and retry on next poll |
| Malformed file content | Move to error directory and continue |
| Partial file detected | Skip files younger than a grace period |

## Idempotency and State Tracking

Track processed files using one of these strategies:

1. **File move/rename**: Move to an archive or processed directory
2. **Filename-based**: Use unique input filenames that are stable across restarts
3. **State file**: Maintain a small state file listing processed filenames

Prefer file move over state files when possible; state files add a failure point.
