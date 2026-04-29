---
name: creating-file-and-sftp-channels
description: >
  Use when a user wants to create, configure, or modify a file or SFTP channel
  in Ghega. Use when someone asks about file channel, SFTP, file reader,
  file writer, or directory polling configuration.
  Use when the topic involves path configuration, polling intervals,
  idempotency via file naming, or YAML definitions for file and SFTP channels.
license: Apache-2.0
---

# Creating File and SFTP Channels

This skill guides the creation and configuration of file and SFTP channels in
Ghega. These channels read from or write to local filesystems, network shares,
or remote SFTP servers.

## When to Use

- Building a file reader (source) that polls a directory for new files
- Building a file writer (destination) that drops output to a filesystem path
- Configuring SFTP source or destination for secure remote file transfer
- Setting up polling intervals and file naming conventions
- Ensuring idempotency when the same file might be seen more than once

## Key Concepts

- **Polling**: The source scans a directory at a regular interval for new files
- **File Pattern**: A glob or regex that limits which files are picked up
- **Move-on-Process**: After reading, the source moves the file to an archive
  or error directory to prevent reprocessing
- **Idempotency**: Using unique file names or tracking processed file hashes

## Scaffolding a File Channel

### File Source Channel

A file source channel polls a local or mounted directory for files matching a
pattern. Typical configuration includes:

- `directory`: absolute or relative path to monitor
- `pattern`: file name glob (e.g., `*.hl7`, `*.xml`)
- `pollInterval`: duration between scans (e.g., `30s`, `5m`)
- `archiveDirectory`: location to move files after successful processing
- `errorDirectory`: location to move files that fail processing

### File Destination Channel

A file destination channel writes transformed messages to a filesystem path.
Typical configuration includes:

- `directory`: target directory path
- `fileNamePattern`: pattern for generated file names
- `append`: whether to append to an existing file or create a new one

## Scaffolding an SFTP Channel

### SFTP Source Channel

An SFTP source channel connects to a remote SFTP server and polls a directory.
Typical configuration includes:

- `host`: SFTP server hostname or IP address
- `port`: SFTP port (default 22)
- `username`: SFTP account name
- `directory`: remote directory to monitor
- `pattern`: file name glob
- `pollInterval`: duration between remote directory scans
- `moveToDirectory`: remote directory for post-processing archival

### SFTP Destination Channel

An SFTP destination channel uploads files to a remote SFTP server.
Typical configuration includes:

- `host`, `port`, `username`: same as source
- `directory`: remote target directory
- `fileNamePattern`: pattern for uploaded file names
- `tempPrefix`: temporary prefix used during upload and renamed on success

## Polling Interval Recommendations

| Use Case | Poll Interval | Notes |
|----------|---------------|-------|
| High-volume realtime | 10s | Monitor CPU and I/O impact |
| Standard batch | 1m | Balanced latency and resource use |
| Nightly batch | 15m | Suitable for large files or slow networks |
| SFTP over WAN | 5m | Reduce load on remote server |

## Idempotency via File Naming

To avoid processing the same file twice:

1. **Archive after read**: Move the file out of the inbound directory immediately
2. **Unique names**: Use timestamp or UUID prefixes in generated file names
3. **Tracking table**: Maintain a record of processed file hashes if archival is
   not possible

## Safety

Never include real SFTP passwords or private key passphrases in channel
configurations or examples. Always use placeholder values and inject credentials
through environment variables or a secrets manager.

Never include real patient data (PHI) in file names, file contents, or examples.
Always use synthetic test data for examples and fixtures.

## References

- See [references/file-patterns.md](references/file-patterns.md) for common file
  source and destination patterns.
- See [references/sftp-patterns.md](references/sftp-patterns.md) for SFTP-specific
  configuration and security considerations.
