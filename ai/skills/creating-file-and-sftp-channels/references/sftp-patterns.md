# SFTP Channel Patterns

This document describes SFTP source and destination patterns for Ghega channels.

## SFTP Source Patterns

### Remote Directory Polling

Polls a remote SFTP directory for new files.

Key configuration:

- Remote host and port (default 22)
- Remote base directory
- Authentication: private key or credential reference
- File filter pattern
- Polling interval
- Post-read action on remote side: move, rename, or delete

Example workflow:

1. Connect to SFTP server every 30 seconds
2. List files in `/remote/incoming/` matching `*.hl7`
3. Download each file to a local staging area
4. Validate and process
5. On success, move remote file to `/remote/archive/`
6. On failure, leave remote file in place or move to `/remote/error/`

### Secure Download with Integrity Check

For sensitive payloads, verify file integrity after download:

- Download the file
- Download a companion checksum file (e.g., `.sha256`) if provided
- Verify checksum before processing
- Reject files that fail integrity verification

## SFTP Destination Patterns

### Remote File Upload

Uploads each message as a separate file to an SFTP server.

Key configuration:

- Remote host and port
- Remote target directory
- Filename template with unique components
- Authentication method
- Temporary upload filename and atomic rename

### Atomic Remote Write

To prevent partial files on the remote server:

1. Upload to a temporary filename (e.g., `.uploading-{uuid}`)
2. Verify upload completion
3. Rename to the final filename on the remote server

This ensures downstream consumers only see complete files.

## Authentication Patterns

| Method | When to Use | Notes |
|--------|-------------|-------|
| Private key | Production | Most secure; reference key path through secret provider |
| Password | Legacy systems only | Rotate frequently; never embed in YAML |
| Host key verification | Always | Prevents man-in-the-middle attacks |

## Connection Management

- Reuse SFTP connections across polling cycles when possible
- Set connection and read timeouts explicitly
- Close connections gracefully on channel shutdown
- Handle stale connections by reconnecting on timeout or error

## Security Guidelines

- Use only SFTP (SSH File Transfer Protocol), not FTP or FTPS unless required
- Verify remote host keys; do not disable host key checking in production
- Restrict SFTP user to a chroot jail or specific directories
- Log transfer metadata (filename, size, duration) without logging payload content
- Encrypt data at rest on both local staging and remote destination when required

## Retry and Error Handling

| Scenario | Recommended Action |
|----------|--------------------|
| Connection timeout | Retry with exponential backoff |
| Authentication failure | Alert immediately; do not retry indefinitely |
| Permission denied on remote | Alert and skip file |
| Remote directory not found | Alert and pause polling |
| Partial upload detected | Delete remote temporary file and retry |
| Host key mismatch | Fail closed and alert |

## Directory Layout Convention

A well-organized remote layout:

```
/remote/
  incoming/     # Files to be consumed by Ghega
  processing/   # Files currently being downloaded
  archive/      # Successfully processed files
  error/        # Files that failed processing
```

Coordinate this layout with the remote system owner to avoid conflicts.
