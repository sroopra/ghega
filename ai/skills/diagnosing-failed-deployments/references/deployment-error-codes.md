# Deployment Error Codes

This reference catalogs common deployment error codes and recommended recovery steps.

## Validation Errors

| Code | Meaning | Recovery |
|------|---------|----------|
| `VAL-001` | Missing required field in channel YAML | Add the missing field and re-validate |
| `VAL-002` | Invalid CEL expression | Check CEL syntax against the specification; test the expression in isolation |
| `VAL-003` | Reference to unknown secret or config map | Create the missing resource or correct the reference name |
| `VAL-004` | Port number out of allowed range | Use a port within the configured range |
| `VAL-005` | Duplicate channel name in namespace | Rename the channel or remove the duplicate |
| `VAL-006` | Invalid certificate reference | Verify the certificate exists and is not expired |

## Dependency Errors

| Code | Meaning | Recovery |
|------|---------|----------|
| `DEP-001` | Referenced channel does not exist | Deploy the dependency first or correct the reference |
| `DEP-002` | Referenced config map not found | Create the config map or correct the key name |
| `DEP-003` | Referenced secret not found | Create the secret or verify the namespace |
| `DEP-004` | Network policy blocks traffic | Update the network policy to allow required traffic |

## Runtime Errors

| Code | Meaning | Recovery |
|------|---------|----------|
| `RUN-001` | Port already in use | Choose a different port or stop the conflicting service |
| `RUN-002` | Certificate load failure | Verify certificate path, format, and permissions |
| `RUN-003` | Listener bind failure | Check firewall rules and interface configuration |
| `RUN-004` | Startup timeout | Increase timeout or review resource constraints |
| `RUN-005` | Health check failure | Review logs for initialization errors |

## General Guidance

When an error code is not listed above:

1. Capture the full error message and stack trace
2. Check the server logs for correlated events
3. Verify that the deployment environment has not changed independently
4. If the issue persists, capture the channel YAML and logs for escalation
