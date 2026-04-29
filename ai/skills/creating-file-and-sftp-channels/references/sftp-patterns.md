# SFTP Channel Patterns

## SFTP Source

Polls a remote SFTP directory and downloads files for processing.

```yaml
source:
  type: sftp
  host: sftp.example.com
  port: 22
  remoteDirectory: /inbound
  pattern: "*.csv"
  pollIntervalSeconds: 120
  auth:
    type: key
    privateKeyRef: SFTP_PRIVATE_KEY
```

## SFTP Destination

Uploads files to a remote SFTP server.

```yaml
destination:
  type: sftp
  host: sftp.example.com
  port: 22
  remoteDirectory: /outbound
  fileNamePattern: "upload-${timestamp}.csv"
  auth:
    type: key
    privateKeyRef: SFTP_PRIVATE_KEY
```

## Authentication

Prefer key-based authentication. Store private keys in environment variables
or a secret manager.

```yaml
auth:
  type: key
  privateKeyRef: SFTP_PRIVATE_KEY
```

Password authentication should only be used in legacy environments.

```yaml
auth:
  type: password
  passwordRef: SFTP_PASSWORD
```

## Connection Settings

| Parameter | Default | Notes |
|-----------|---------|-------|
| port | 22 | Standard SFTP port |
| timeoutSeconds | 30 | Connection and read timeout |
| maxConnections | 5 | Connection pool size |
| retryAttempts | 3 | Retries on transient failures |

## Host Key Verification

Always verify the remote host key to prevent man-in-the-middle attacks.
Store the expected host key fingerprint in configuration.

```yaml
source:
  type: sftp
  host: sftp.example.com
  hostKeyFingerprint: "SHA256:abc..."
```

## Safety

- Never commit private keys or passwords to version control
- Restrict SFTP user permissions to the minimum required directories
- Enable logging of file transfers for audit purposes
