# Security Review Checklist

This checklist maps channel security review items to OWASP ASVS Level 2,
HIPAA technical safeguards, and GDPR Article 32 requirements.

## Network and Transport

- [ ] TLS is enabled for all supported source and destination types
- [ ] Certificates are valid and not expiring within the rotation window
- [ ] Mutual TLS is configured where required by policy
- [ ] Firewall rules restrict source IPs to expected ranges
- [ ] Weak TLS versions and cipher suites are disabled
- [ ] Internal service traffic uses encrypted channels

**Maps to:** OWASP ASVS V6.1, V9.1; HIPAA 164.312(e); GDPR Article 32 (encryption)

## Payload Retention and Storage

- [ ] Retention period is defined, documented, and enforced
- [ ] Retention aligns with the shortest applicable regulation or business need
- [ ] Error and dead-letter queues follow the same retention policy
- [ ] Backups and replicas respect the retention schedule
- [ ] Secure deletion is used when retention expires

**Maps to:** HIPAA 164.312(c); GDPR Article 32 (pseudonymization, integrity)

## Logging and Audit

- [ ] Logs contain metadata and identifiers, not payload content
- [ ] Access to detailed payload logs requires explicit authorization
- [ ] Log destinations are tamper-evident and access-controlled
- [ ] Log retention meets regulatory requirements
- [ ] Failed access attempts are logged
- [ ] Configuration changes are logged with actor and timestamp

**Maps to:** OWASP ASVS V7.1; HIPAA 164.312(b); GDPR Article 32 (availability, monitoring)

## Access Control and Permissions

- [ ] Channel configuration access follows least-privilege
- [ ] Replay capability is restricted to authorized roles
- [ ] Replay actions are logged with initiator identity
- [ ] Secret access is scoped to the channel or namespace that needs it
- [ ] Role assignments are reviewed on a regular schedule

**Maps to:** OWASP ASVS V4.1; HIPAA 164.312(a), 164.312(d); GDPR Article 32 (confidentiality)

## Secret and Credential Management

- [ ] Secrets are stored in a dedicated secret store
- [ ] Channel YAML references secrets by name and key, not by value
- [ ] Secrets are rotated on a regular schedule
- [ ] Old secret versions are revoked promptly after rotation
- [ ] Secrets are not logged, printed, or returned in API responses
- [ ] No secrets are committed to version control

**Maps to:** OWASP ASVS V2.1, V6.1, V8.1; HIPAA 164.312(a), 164.312(e); GDPR Article 32 (confidentiality)

## Message Integrity

- [ ] Message integrity checks are enabled where supported by the protocol
- [ ] Tamper detection is configured for audit logs
- [ ] Replayed messages are tagged to distinguish them from live traffic
- [ ] Destinations that are not idempotent have replay safeguards

**Maps to:** HIPAA 164.312(c), 164.312(e); GDPR Article 32 (integrity)

## Code and Configuration

- [ ] No executable scripts are embedded in channel definitions
- [ ] Channel YAML is reviewed before deployment
- [ ] Dependencies and base images are scanned for known vulnerabilities
- [ ] Configuration drift is detected and alerted

**Maps to:** OWASP ASVS V1.1, V10.1; GDPR Article 32 (regular testing)

## Review Documentation

After completing the checklist:

- [ ] Record the review date and reviewer identity
- [ ] Note any findings and their severity
- [ ] Assign remediation tasks with due dates
- [ ] Schedule the next review
