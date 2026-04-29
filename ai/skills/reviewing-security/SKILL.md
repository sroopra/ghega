---
name: reviewing-security
description: >
  Use when performing a security review, auditing a channel, checking permissions,
  or when someone asks about HIPAA compliance, GDPR, or OWASP. Use when the topic
  involves network policy, payload retention, PHI logging, replay permissions, or
  secret handling in Ghega channels.
license: Apache-2.0
---

# Reviewing Security

This skill guides security reviews of Ghega channels and configurations to identify risks and ensure compliance.

## When to Use

- A security review is requested for a channel or set of channels
- Auditing channel permissions and access controls
- Verifying HIPAA technical safeguard compliance
- Checking GDPR Article 32 requirements
- Reviewing against OWASP ASVS Level 2

## Key Concepts

- **Network Policy**: Rules governing ingress, egress, and inter-service communication
- **Payload Retention**: How long message payload data is stored and where
- **PHI Logging**: Ensuring protected health information never appears in logs
- **Replay Permissions**: Who can trigger replays and under what conditions
- **Secret Handling**: How credentials and keys are stored and referenced

## Security Review Workflow

### 1. Review Network Policy

- Verify least-privilege network access for source and destination connectors
- Confirm TLS is required for external endpoints
- Check that internal services use mTLS where supported
- Validate firewall rules restrict traffic to known ports and hosts

### 2. Review Payload Retention

- Confirm retention periods match organizational policy and regulatory requirements
- Verify encrypted storage for payload data at rest
- Check that backup and restore procedures preserve encryption
- Ensure data deletion workflows are tested and documented

### 3. Review PHI Logging

- Confirm no payload fields are logged in plain text
- Verify log levels do not emit debug output containing message content
- Check that audit logs record metadata only (message IDs, timestamps, actions)
- Review log aggregation and retention for accidental PHI inclusion

### 4. Review Replay Permissions

- Verify replay capability is restricted to authorized roles
- Confirm replay actions are audited with metadata only
- Check that replays to non-production environments require additional approval
- Validate that replay rate limits prevent abuse

### 5. Review Secret Handling

- Confirm secrets are not hardcoded in channel YAML
- Verify secrets are referenced via secret stores or environment variables
- Check rotation policies and expiry alerts for TLS certificates and API keys
- Validate that secret access is scoped to the minimum required set

## Compliance Mappings

- See [references/security-checklist.md](references/security-checklist.md) for detailed mappings to OWASP ASVS L2, HIPAA technical safeguards, and GDPR Article 32.

## Safety

- Never include real credentials, tokens, or keys in review outputs
- Use synthetic examples when demonstrating configurations
- Report findings as metadata and recommendations, never as executable patches
- Escalate critical findings immediately rather than documenting them only
