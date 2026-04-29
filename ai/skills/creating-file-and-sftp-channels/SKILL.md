---
name: creating-file-and-sftp-channels
description: >
  Use when a user wants to create, configure, or modify a file or SFTP channel in Ghega.
  Use when someone asks about file channel, SFTP, file reader, file writer,
  directory polling, or file-based integration patterns.
  Use when the topic involves local file system, SFTP source/destination,
  file path configuration, polling intervals, or file transfer semantics.
license: Apache-2.0
---

# Creating File and SFTP Channels

This skill guides the creation and configuration of file and SFTP source and
destination channels in Ghega for file-based integrations and secure transfers.

## When to Use

- Building a new file reader (source) or file writer (destination) channel
- Configuring SFTP source or destination connectors
- Setting up directory polling and file watching
- Defining path templates, polling intervals, and file naming conventions
- Troubleshooting file transfer or permission issues

## Key Concepts

- **File Source**: Polls a directory and reads files matching a pattern
- **File Destination**: Writes output to a file path
- **SFTP Source**: Polls a remote directory over SFTP
- **SFTP Destination**: Writes files to a remote SFTP server
- **Polling Interval**: How often the source checks for new files
- **Idempotency via File Naming**: Using unique, stable filenames to prevent reprocessing

## Scaffold Guidelines

### File Source Channel

- Define the source directory and file filter pattern (e.g., `*.hl7`, `*.xml`)
- Set the polling interval based on latency requirements
- Configure move or archive behavior after successful read
- Set file encoding explicitly (UTF-8, ASCII, etc.)

### File Destination Channel

- Define the output directory and filename template
- Use timestamp or sequence components to avoid overwrites
- Configure temporary write paths and atomic move on completion
- Set permissions and ownership when relevant

### SFTP Source Channel

- Define the remote host, port, and base directory
- Configure the private key or credential reference (never embed secrets in YAML)
- Set the file filter pattern and polling interval
- Configure move or archive on the remote side after successful read

### SFTP Destination Channel

- Define the remote host, port, and target directory
- Use filename templates that ensure uniqueness
- Configure temporary remote filenames and atomic rename on completion
- Validate host key when possible

## Safe Defaults

| Setting | Recommended Default | Rationale |
|---------|---------------------|-----------|
| Polling interval | 30s | Balances latency with I/O overhead |
| Max file size | 100 MB | Prevents memory exhaustion |
| Read buffer | 64 KB | Reasonable default for most messages |
| Archive retention | 7 days | Allows replay without unbounded growth |
| SFTP timeout | 30s | Prevents hanging connections |
| SFTP retry count | 3 | Handles transient network issues |

## Idempotency via File Naming

Use stable, unique filenames to prevent duplicate processing:

- Include a correlation ID or message ID in the filename
- Use timestamps with sufficient granularity
- Avoid filenames that collide when multiple messages arrive concurrently
- Example safe pattern: `{correlation-id}-{timestamp}.hl7`

After successful processing:

- Move files to an archive directory, or
- Delete files if archive is not required, or
- Rename with a processed suffix (e.g., `.done`)

## Safety

Never include real patient data (PHI) in file examples or channel configurations.
Always use synthetic test data for examples and fixtures.
Never embed SFTP passwords or private keys in channel YAML; reference them through
the Ghega secret provider.

## References

- See [references/file-patterns.md](references/file-patterns.md) for local file
  system source and destination patterns.
- See [references/sftp-patterns.md](references/sftp-patterns.md) for SFTP
  connector configuration and secure transfer patterns.
