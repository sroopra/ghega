---
name: reviewing-security
description: >
  Use when performing a security review, auditing a channel, checking permissions,
  or evaluating compliance. Use when the topic involves HIPAA compliance, GDPR,
  OWASP, security risks, network policy, payload retention, PHI logging,
  replay permissions, or secret handling.
license: Apache-2.0
---

# Reviewing Security

This skill guides security reviews of Ghega channels and their surrounding
infrastructure.

## When to Use

- Conducting a security audit of one or more channels
- Reviewing channel permissions and access controls
- Assessing HIPAA or GDPR compliance posture
- Evaluating channels against OWASP ASVS requirements
- Investigating secret handling or network policy concerns

## Security Review Overview

A channel security review examines five areas:

1. Network policy and transport security
2. Payload retention and storage
3. PHI logging and audit trails
4. Replay permissions and message reprocessing
5. Secret handling and credential management

## Network Policy

Network policy controls which systems can communicate with a channel.

Review checklist:

- [ ] Source and destination endpoints use TLS where supported
- [ ] Certificate chains are valid and not near expiry
- [ ] Mutual TLS is enabled for sensitive endpoints if required by policy
- [ ] Firewall rules restrict source IPs to expected ranges
- [ ] Internal service-to-service traffic uses encrypted channels
- [ ] Ports exposed by listeners are documented and minimized

Common risks:

| Risk | Indicators | Mitigation |
|------|------------|------------|
| Unencrypted transport | `tls: disabled` or plain HTTP/MLLP | Enable TLS and enforce certificate validation |
| Overly broad IP allow-list | `0.0.0.0/0` or large CIDR blocks | Restrict to known source ranges |
| Expired certificate | Certificate `notAfter` in the past | Rotate certificates before expiry |
| Weak cipher suites | Legacy TLS versions enabled | Enforce TLS 1.2 or higher with strong ciphers |

## Payload Retention

Payload retention policies determine how long message content is kept.

Review checklist:

- [ ] Retention period is defined and documented for each channel
- [ ] Retention aligns with the shortest applicable regulation or business need
- [ ] Messages in error or dead-letter queues have the same retention as successful messages
- [ ] Backups and replicas respect the same retention schedule
- [ ] Secure deletion is used when retention expires

For healthcare channels, default to the minimum retention necessary.

## PHI Logging

Audit logs must record access and actions without exposing protected health information.

Review checklist:

- [ ] Logs contain `messageId` and metadata, not payload content
- [ ] Access to payload logs requires explicit authorization
- [ ] Log destinations are tamper-evident and access-controlled
- [ ] Log retention meets regulatory requirements
- [ ] Failed access attempts are logged

Logging rules:

- Log channel start, stop, configuration changes, and errors
- Log authentication and authorization decisions
- Do not log patient identifiers, diagnosis codes, or clinical values
- If payload size is logged, do so without content inspection

## Replay Permissions

Message replay allows reprocessing of previously received messages.

Review checklist:

- [ ] Replay capability is restricted to authorized roles
- [ ] Replay actions are logged with initiator identity and time
- [ ] Replayed messages are tagged to distinguish them from live traffic
- [ ] Replay rate limits prevent accidental overload
- [ ] Destinations that are not idempotent have replay safeguards

Replay is powerful for recovery but can violate data-consent boundaries if
misused. Treat replay permissions as highly sensitive.

## Secret Handling

Secrets include API keys, certificates, database credentials, and tokens.

Review checklist:

- [ ] Secrets are stored in a dedicated secret store, not in channel YAML
- [ ] Channel YAML references secrets by name and key, not by value
- [ ] Secrets are rotated on a regular schedule
- [ ] Old secret versions are revoked promptly after rotation
- [ ] Secret access is scoped to the channel or namespace that needs it
- [ ] Secrets are not logged, printed, or returned in API responses

Never commit secret values to version control or paste them into documentation.

## Compliance Mapping

### OWASP ASVS Level 2

| ASVS Requirement | Channel Review Action |
|------------------|-----------------------|
| V1.1 Secure software development lifecycle | Verify security review is part of channel deployment process |
| V2.1 Password security | Verify secrets are not hardcoded; use secret store |
| V3.1 Session management | Verify token and session handling in HTTP channels |
| V4.1 Access control | Verify channel permissions follow least-privilege |
| V6.1 Cryptography | Verify TLS and certificate usage |
| V7.1 Log content | Verify logs do not contain PHI or secrets |
| V8.1 Data protection | Verify encryption at rest and in transit |
| V9.1 Communications security | Verify secure transport for all connectors |
| V10.1 Malicious code | Verify no executable scripts in channel definitions |

### HIPAA Technical Safeguards

| Safeguard | Channel Review Action |
|-----------|-----------------------|
| Access control (164.312(a)) | Verify role-based access to channel configuration and replay |
| Audit controls (164.312(b)) | Verify logging covers access and configuration changes |
| Integrity (164.312(c)) | Verify message integrity checks and tamper detection |
| Person or entity authentication (164.312(d)) | Verify mutual TLS or credential-based authentication |
| Transmission security (164.312(e)) | Verify encryption in transit and integrity controls |

### GDPR Article 32

| Requirement | Channel Review Action |
|-------------|-----------------------|
| Pseudonymization and encryption | Verify payload encryption at rest and in transit |
| Ongoing confidentiality, integrity, availability | Verify access controls, monitoring, and backup procedures |
| Ability to restore availability and access | Verify replay, backup, and recovery capabilities |
| Regular testing and evaluation | Verify security reviews are scheduled and documented |

## References

- See [references/security-checklist.md](references/security-checklist.md)
  for a printable security review checklist mapped to the frameworks above.
