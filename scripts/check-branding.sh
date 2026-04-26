#!/usr/bin/env bash
set -euo pipefail

# This script verifies that no "CareMeld" or "caremeld" strings remain in the
# repository (case-insensitive). Historical planning documents (*plan*.md) and
# the script itself are excluded because they may reference the old name by design.

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Find any matches, excluding .git, this script, test files that check for
# the string, and historical plan files.
MATCHES=$(grep -ri "caremeld" "${REPO_ROOT}" \
    --exclude-dir=.git \
    --exclude="$(basename "$0")" \
    --exclude='*_test.go' \
    --exclude='*plan*.md' \
    || true)

if [ -n "${MATCHES}" ]; then
    echo "ERROR: Found remaining 'CareMeld' / 'caremeld' references in the repository:"
    echo "${MATCHES}"
    exit 1
fi

echo "OK: No 'CareMeld' / 'caremeld' references found in source files."
