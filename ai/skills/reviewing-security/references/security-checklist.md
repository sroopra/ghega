# Security Checklist

This checklist maps Ghega channel security controls to common compliance frameworks.

## OWASP ASVS Level 2 Mapping

| ASVS ID | Requirement | Ghega Control | Verification |
|---------|-------------|---------------|--------------|
| V1.1.1 | Secure software development lifecycle | Channel YAML is version-controlled and reviewed | Verify pull request reviews include security checks |
| V1.2.1 | Authentication mechanisms | API and CLI use authenticated sessions | ✅ OIDC session cookie auth implemented (ADR-010) |
| V1.4.1 | Access control design | RBAC restricts channel create, deploy, and replay | Review role bindings for least privilege |
| V2.10.1 | Service authentication | mTLS or token auth for internal services | Verify connector configs require TLS |
| V4.1.1 | HTTP request security | HTTP connectors use TLS and validate certificates | Check `tls.enabled` and `tls.verify` settings |
| V6.1.1 | Data classification | Payload retention policies are defined and enforced | Confirm retention matches policy |
| V6.2.1 | Algorithms and key lengths | Encryption uses approved algorithms | Verify AES-256-GCM or equivalent for payload storage |
| V7.1.1 | Log content security | Logs contain no PHI or secrets | Review sample log entries for metadata-only content |
| V8.2.1 | Client-side data protection | Payload is not cached in UI or CLI in plain text | Verify UI and CLI do not store payload locally |
| V9.1.1 | Communications security | All external traffic uses TLS | Confirm no plaintext HTTP or MLLP without wrapper |

## HIPAA Technical Safeguards Mapping

| Safeguard | Regulation Reference | Ghega Control | Verification |
|-----------|---------------------|---------------|--------------|
| Access Control | 164.312(a) | RBAC and authentication for all channel operations | Verify user and service account permissions |
| Audit Controls | 164.312(b) | Audit logs record all channel deploy, replay, and access events | Review audit log coverage and integrity |
| Integrity | 164.312(c) | Channel configurations are versioned and signed | Verify commit signing and immutable history |
| Person or Entity Authentication | 164.312(d) | API and CLI require valid credentials | ✅ OIDC session cookie auth implemented (ADR-010) |
| Transmission Security | 164.312(e) | TLS for data in transit | Validate certificate chains and cipher suites |
| Automatic Logoff | 164.312(a)(2)(iii) | Sessions expire after inactivity | ✅ Session cookie lifecycle managed server-side (ADR-010) |
| Encryption and Decryption | 164.312(a)(2)(iv) | Payload encrypted at rest and in transit | Verify storage encryption and TLS settings |

## GDPR Article 32 Mapping

| Article 32 Requirement | Ghega Control | Verification |
|------------------------|---------------|--------------|
| Pseudonymization | Payloads are referenced by opaque IDs in audit logs | Confirm PayloadRef usage in logs |
| Encryption | Payload encrypted at rest and in transit | Verify encryption configuration |
| Ongoing confidentiality | RBAC and network policies restrict access | Review access controls regularly |
| Ongoing integrity | Channel configs versioned and change-reviewed | Verify code review and CI gates |
| Ongoing availability | Redundancy and backup for channel state and payloads | Test restore procedures |
| Ability to restore | Backup and disaster recovery tested | Confirm recovery time objectives are met |

## Ghega-Specific Checks

### Network Policy

- [ ] Source listeners bind only to required interfaces
- [ ] Destination connectors use allow-listed endpoints
- [ ] Internal service communication uses mTLS where available
- [ ] Firewall rules are documented and reviewed quarterly

### Payload Retention

- [ ] Retention period is defined per channel and environment
- [ ] Automated deletion workflows are enabled and monitored
- [ ] Backups are encrypted and retention-limited
- [ ] Cross-border transfer policies are documented

### PHI Logging

- [ ] No payload content in application logs
- [ ] No payload content in error traces
- [ ] Audit logs contain metadata only
- [ ] Log aggregation pipelines filter out accidental PHI

### Replay Permissions

- [ ] Replay role is separate from deploy and read roles
- [ ] Replay to production requires additional approval
- [ ] Replay rate limits are configured per channel
- [ ] All replay actions are audited

### Secret Handling

- [ ] Secrets are externalized to a secret store
- [ ] Channel YAML references secrets by name, not value
- [ ] TLS certificates have expiry monitoring
- [ ] API keys are rotated on a defined schedule

## Example Safe Configuration Snippet

```yaml
connector:
  type: http
  endpoint: "https://api.example.com/hl7"
  tls:
    enabled: true
    verify: true
  auth:
    type: bearer
    tokenRef: "secrets.api-token"  # Referenced, not embedded
retention:
  payload_days: 30
  audit_days: 365
logging:
  level: info
  include_payload: false
```
