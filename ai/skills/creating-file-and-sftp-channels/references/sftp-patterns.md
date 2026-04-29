# SFTP Patterns

This reference covers common patterns and security considerations for SFTP
source and destination channels.

## SFTP Source Patterns

### Remote Directory Polling

Connects to an SFTP server and polls a remote directory for new files. This is
the standard pattern for receiving files from external partners.

Key settings:
- `host` and `port`: SFTP server endpoint
- `username`: SFTP account (use a dedicated service account)
- `directory`: remote inbound directory
- `pattern`: restrict to expected file types
- `pollInterval`: be considerate of remote server resources
- `moveToDirectory`: remote archive directory; move files after successful pickup

### Secure Key Authentication

Prefer SSH private key authentication over password-based login. Store the
private key path in the channel configuration and protect the key file with
strict filesystem permissions (0400).

Never store the private key contents inline in the channel YAML.

## SFTP Destination Patterns

### Atomic Upload

Upload files using a temporary name and rename them to the final name on
completion. This prevents downstream consumers from reading partial files.

Key settings:
- `tempPrefix`: temporary file prefix (e.g., `.tmp-`)
- `fileNamePattern`: final file name after successful upload

### Directory Partitioning

Organize uploaded files into remote subdirectories by date or message type.
This simplifies partner-side archival and reconciliation.

Example remote path:
- `/inbound/adt/{{date}}/{{timestamp}}-adt.hl7`

## Connection and Retry Recommendations

| Scenario | Timeout | Retries | Notes |
|----------|---------|---------|-------|
| LAN SFTP | 10s | 3 | fast reconnect |
| WAN SFTP | 30s | 5 | longer backoff |
| Unreliable partner | 60s | 10 | alert after repeated failure |

## Security Considerations

- Use only SFTP (SSH File Transfer Protocol), not plain FTP
- Validate the remote host key and pin it in the channel configuration if
  possible
- Rotate SSH keys regularly and revoke old keys promptly
- Restrict the SFTP service account to the minimum required directories
- Enable connection logging on the SFTP server for audit purposes
- Do not enable shell access for the SFTP service account

## Common Pitfalls

- **Stale connections**: Some servers drop idle connections; configure a keepalive
  or accept the reconnection cost
- **Clock skew**: Timestamp-based file selection can fail if server clocks differ;
  use sequence numbers or server-side file metadata when possible
- **Large files**: For files larger than available memory, stream the transfer
  rather than loading the entire file into memory
