---
name: diagnosing-failed-deployments
description: >
  Use when a deployment failed, a channel deploy error occurred, validation failed,
  a deploy was rejected, or rollback is needed. Use when someone asks about deployment
  errors, channel validation output, revision comparison, or whether to fix-and-retry
  versus rollback.
license: Apache-2.0
---

# Diagnosing Failed Deployments

This skill guides the diagnosis and recovery of failed Ghega channel deployments.

## When to Use

- A channel deployment returned an error or was rejected
- Validation output shows configuration problems
- A deployed channel behaves differently than expected
- Rollback is being considered after a bad deployment

## Key Concepts

- **Deployment Error**: A failure returned by the deploy API or CLI
- **Validation Output**: Structured results from `ghega channel validate`
- **Revision**: A specific version of channel configuration
- **Rollback**: Reverting to a previous known-good revision

## Diagnostic Workflow

### 1. Read the Deployment Error

- Check the CLI or API response for the exact error message
- Note the error code and which resource failed
- Distinguish between client errors (validation) and server errors (runtime)

### 2. Check Channel Validation Output

- Run `ghega channel validate` against the channel YAML
- Review each validation result: error, warning, or info
- Fix validation errors before retrying deployment

### 3. Compare Revisions

- Retrieve the current deployed revision
- Diff it against the revision being deployed
- Focus on changes to connectors, mappings, and routes

### 4. Decide: Fix-and-Retry vs Rollback

Choose fix-and-retry when:
- The fix is small and isolated
- Validation errors are clearly identified
- The environment can tolerate another deploy attempt

Choose rollback when:
- The failure is systemic or unclear
- Multiple dependent channels are affected
- The previous revision is known to be stable

## Safety

- Never deploy untested changes to production during business hours without approval
- Always validate locally before deploying
- Log deployment decisions as metadata, never including secrets or payload content
- Verify rollback target revision was previously healthy

## References

- See [references/deployment-error-codes.md](references/deployment-error-codes.md) for a detailed error code reference.
