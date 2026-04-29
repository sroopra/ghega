---
name: creating-file-and-sftp-channels
description: >
  Use when a user wants to create, configure, or modify a file or SFTP channel in Ghega.
  Use when someone asks about file channel, SFTP, file reader, file writer,
  or directory polling configuration.
  Use when the topic involves file path setup, polling intervals, or idempotency
  via file naming conventions.
license: Apache-2.0
---

# Creating File and SFTP Channels

This skill guides the creation and configuration of file and SFTP source and
destination channels in Ghega. These channels are used for batch file processing,
directory polling, and secure file transfer.

## When to Use

- Building a new file reader or writer channel
- Configuring SFTP inbound or outbound transfer
- Setting up directory polling with intervals
- Ensuring idempotency via file naming strategies

## Key Concepts

- **File Source**: Polls a directory and reads files matching a pattern
- **File Destination**: Writes messages to files in a target directory
- **SFTP Source**: Polls a remote directory over SFTP
- **SFTP Destination**: Writes files to a remote directory over SFTP
- **Polling Interval**: How often the directory is scanned

## File Source Configuration

```yaml
source:
  type: file
  directory: /data/inbound
  pattern: "*.hl7"
  pollIntervalSeconds: 60
```

## SFTP Destination Configuration

```yaml
destination:
  type: sftp
  host: sftp.example.com
  port: 22
  remoteDirectory: /outbound
  fileNamePattern: "export-${timestamp}.json"
```

## Idempotency via File Naming

Use deterministic file names or move processed files to an archive directory to
avoid duplicate processing.

Strategies:
- Include a message identifier or hash in the file name
- Move files to `.done/` or `.archive/` after processing
- Use atomic rename operations where supported

## Safety

Never include real credentials, private keys, or host passwords in channel configurations.
Use environment-variable references or a secret manager for SFTP authentication.
Ensure file permissions restrict read and write access to the runtime user.

## References

- See [references/file-patterns.md](references/file-patterns.md) for local file channel patterns.
- See [references/sftp-patterns.md](references/sftp-patterns.md) for SFTP-specific configuration.
