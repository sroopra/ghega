# File Channel Patterns

## Local File Source

Polls a local directory for new files and reads them into the channel.

```yaml
source:
  type: file
  directory: /data/inbound
  pattern: "*.hl7"
  pollIntervalSeconds: 60
  deleteAfterRead: false
  moveAfterRead: /data/archive
```

## Local File Destination

Writes channel output to files in a local directory.

```yaml
destination:
  type: file
  directory: /data/outbound
  fileNamePattern: "output-${timestamp}.txt"
  append: false
```

## Directory Polling

Polling interval determines how responsive the channel is. Shorter intervals
increase CPU and I/O load.

| Scenario | Recommended Interval |
|----------|----------------------|
| High volume real-time | 5 - 15 seconds |
| Standard batch | 60 seconds |
| Low volume nightly | 300 - 900 seconds |

## File Name Patterns

Use safe, deterministic patterns to avoid collisions.

- `${timestamp}` — Unix timestamp in milliseconds
- `${messageId}` — Channel-assigned message identifier
- `${date}` — Current date in YYYY-MM-DD format

## Atomic Write

To prevent consumers from reading partially written files, write to a temporary
name and rename on completion.

```yaml
destination:
  type: file
  directory: /data/outbound
  tempPrefix: ".tmp-"
  fileNamePattern: "output-${timestamp}.txt"
```

## Idempotency Checklist

- [ ] Processed files are moved or renamed after completion
- [ ] Duplicate file names trigger a warning rather than overwrite
- [ ] File checksum is logged for audit trails
