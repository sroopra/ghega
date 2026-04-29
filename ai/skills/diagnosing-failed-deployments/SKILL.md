---
name: diagnosing-failed-deployments
description: >
  Use when a deployment failed, a channel deploy error occurred, or validation
  failed during deployment. Use when a deploy was rejected or rollback is needed.
  Use when someone asks about deployment errors, channel validation output,
  comparing revisions, or deciding between fix-and-retry versus rollback.
license: Apache-2.0
---

# Diagnosing Failed Deployments

This skill guides diagnosis and recovery when a Ghega channel deployment fails.

## When to Use

- A channel deployment returned an error or was rejected
- Validation output indicates a configuration problem
- A deployed channel behaves differently than expected
- A rollback to a previous revision is being considered

## Reading Deployment Errors

Deployment errors fall into several categories:

| Category | Typical Cause | First Action |
|----------|---------------|--------------|
| Validation error | Invalid YAML, missing required field, bad reference | Read the validation message and fix the source file |
| Dependency error | Referenced channel, secret, or config map missing | Verify the referenced resource exists in the target namespace |
| Runtime error | Listener port conflict, certificate issue, network policy | Check server logs and network policy rules |
| Timeout | Deployment took longer than the configured deadline | Review resource limits and retry |

When an error message is returned:

1. Note the error code and message text
2. Identify whether the failure happened during validation, scheduling, or runtime startup
3. Check the deployment logs for stack traces or detailed cause

## Checking Channel Validation Output

Channel validation runs before deployment. Common validation failures include:

- Missing `source` or `destination` configuration
- Invalid CEL expression in a mapping or filter
- Reference to a non-existent secret or config map key
- Port number outside the allowed range
- Duplicate channel name in the same namespace

Validation output typically includes:

- `field`: the YAML path that failed validation
- `rule`: the validation rule that was violated
- `message`: human-readable explanation

Fix the source YAML and re-run validation before attempting deployment again.

## Comparing Revisions

If a previously working channel now fails after an update, compare revisions:

1. Retrieve the last known good revision from the channel store or version control
2. Diff the current YAML against the last good revision
3. Pay attention to changes in:
   - Source or destination connector settings
   - Mapping or filter expressions
   - Timeout and retry configurations
   - Secret references or certificate names

If the diff shows no intentional changes, check for indirect changes such as:
- Updated base images or channel templates
- Rotated certificates or credentials
- Changed network policies or firewall rules

## Fix-and-Retry vs. Rollback

Choose fix-and-retry when:

- The root cause is identified and the fix is small and local
- Validation output points to a specific, correctable issue
- The deployment environment is stable and the failure is isolated

Choose rollback when:

- The root cause is unclear or requires significant investigation
- The failure affects multiple channels or the entire namespace
- A production incident is ongoing and recovery time matters more than root-cause analysis

Rollback procedure:

1. Identify the last known good revision
2. Mark the current revision as failed in the deployment record
3. Apply the last known good revision
4. Verify the channel returns to healthy status
5. Open a follow-up task to investigate the failed revision in a non-production environment

## References

- See [references/deployment-error-codes.md](references/deployment-error-codes.md)
  for a catalog of common deployment error codes and recovery steps.
