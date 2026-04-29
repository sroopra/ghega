# Deployment Error Codes Reference

## Client Errors (Validation)

| Code | Meaning | Common Cause | Recovery |
|------|---------|--------------|----------|
| `VALIDATION_SCHEMA` | YAML does not match schema | Missing required field or wrong type | Review schema docs and fix YAML |
| `VALIDATION_MAPPING` | Mapping expression is invalid | CEL syntax error or unknown field path | Test expression locally before deploy |
| `VALIDATION_CONNECTOR` | Connector configuration is invalid | Missing endpoint or unsupported protocol | Verify connector settings against supported list |
| `VALIDATION_ROUTE` | Route rule is invalid | Missing condition or duplicate route name | Simplify route and re-validate |

## Server Errors (Runtime)

| Code | Meaning | Common Cause | Recovery |
|------|---------|--------------|----------|
| `DEPLOY_TIMEOUT` | Deployment timed out | Large channel or slow network | Retry during lower load or split deployment |
| `DEPLOY_CONFLICT` | Revision conflict | Another deployment occurred concurrently | Fetch latest revision and re-apply changes |
| `DEPLOY_RESOURCE_UNAVAILABLE` | Required resource not ready | Database or queue not reachable | Verify infrastructure health before retry |
| `DEPLOY_PERMISSION_DENIED` | Insufficient privileges | Service account lacks required role | Check RBAC policy and request elevated access |

## Rollback Errors

| Code | Meaning | Common Cause | Recovery |
|------|---------|--------------|----------|
| `ROLLBACK_REVISION_MISSING` | Target revision not found | Revision was pruned or never deployed | Choose an earlier known-good revision |
| `ROLLBACK_VALIDATION_FAILED` | Rolled-back revision fails validation | Schema changed since that revision | Manually reconstruct a compatible revision |

## General Recovery Steps

1. Capture the full error response and timestamp
2. Check the Ghega system health dashboard for correlated alerts
3. Review recent deployment audit logs for patterns
4. Apply the smallest fix that resolves the root cause
5. Validate the fix locally before redeploying
6. Monitor post-deployment metrics for regressions

## Audit Log Example (Safe)

```yaml
deployment_id: deploy-20240101-001
channel_id: adt-inbound
revision: rev-abc-123
timestamp: 2024-01-01T12:00:00Z
status: failed
error_code: VALIDATION_MAPPING
notes: CEL expression referenced unknown field. Fixed and redeployed as deploy-20240101-002.
# No secrets or payload content logged.
```
