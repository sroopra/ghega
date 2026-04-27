#!/usr/bin/env bash
# Ghega UI Dependency Audit
# Runs npm audit, generates an SBOM/dependency list, and fails on critical vulnerabilities.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UI_DIR="${SCRIPT_DIR}/../ui"

cd "${UI_DIR}"

echo "=== Ghega UI Dependency Audit ==="
echo ""

echo "--- Generating dependency list (SBOM) ---"
npm list --json > sbom.json 2>/dev/null || true
cat sbom.json | jq -r '.dependencies | keys[]' > dependencies.txt 2>/dev/null || npm list --depth=0 > dependencies.txt
echo "Dependency list written to: ${UI_DIR}/dependencies.txt"
echo ""

echo "--- Running npm audit ---"
# Run audit and capture output; fail if critical vulnerabilities are found
AUDIT_OUTPUT=$(npm audit --audit-level=moderate --json 2>/dev/null || true)

# Check for critical vulnerabilities in the audit output
CRITICAL_COUNT=$(echo "${AUDIT_OUTPUT}" | jq '.metadata.vulnerabilities.critical // 0' 2>/dev/null || echo "0")

if [ "${CRITICAL_COUNT}" -gt 0 ]; then
  echo "CRITICAL vulnerabilities found: ${CRITICAL_COUNT}"
  echo "${AUDIT_OUTPUT}" | jq '.vulnerabilities | to_entries[] | select(.value.severity == "critical") | {package: .key, severity: .value.severity}' 2>/dev/null || true
  echo ""
  echo "Audit failed due to critical vulnerabilities."
  exit 1
fi

echo "No critical vulnerabilities found."
echo "Audit complete."
