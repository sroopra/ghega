#!/usr/bin/env bash
set -euo pipefail

OLD_NAME="${OLD:-}"
NEW_NAME="${NEW:-}"

if [ -z "${OLD_NAME}" ] || [ -z "${NEW_NAME}" ]; then
    echo "Usage: make rename-product OLD=<old-slug> NEW=<new-slug>"
    echo "Error: both OLD and NEW must be set."
    exit 1
fi

if [ "${OLD_NAME}" = "${NEW_NAME}" ]; then
    echo "Error: OLD and NEW are the same. Nothing to do."
    exit 0
fi

# Build a list of files to modify
files=(
    branding/product.yaml
    README.md
    Makefile
    internal/cli/root.go
    internal/server/server.go
)

# Add ui/src/ files that contain the old name
while IFS= read -r -d '' f; do
    files+=("$f")
done < <(find ui/src -type f \( -name '*.tsx' -o -name '*.ts' -o -name '*.jsx' -o -name '*.js' -o -name '*.html' -o -name '*.css' -o -name '*.md' \) -print0 2>/dev/null || true)

changed=()

for f in "${files[@]}"; do
    if [ ! -f "${f}" ]; then
        continue
    fi

    # Use sed with word boundaries to avoid partial matches.
    # \b works for GNU sed. On macOS BSD sed uses [[:<:]] and [[:>:]].
    # We assume GNU sed in this environment (Linux).
    if sed -i "s/\\b${OLD_NAME}\\b/${NEW_NAME}/g" "${f}"; then
        # Check if the file was actually modified
        if git diff --quiet -- "${f}" 2>/dev/null; then
            : # no change
        else
            changed+=("${f}")
        fi
    fi
done

# Revert Go module import paths and package references — product rename
# should not break module names or package-level identifiers.
for f in "${changed[@]}"; do
    if [[ "$f" == *.go ]]; then
        old_escaped=$(printf '%s' "${OLD_NAME}" | sed 's/[]\/.^$*]/\\&/g')
        new_escaped=$(printf '%s' "${NEW_NAME}" | sed 's/[]\/.^$*]/\\&/g')
        # Revert import paths
        sed -i "s|github\.com/\([^/]*\)/${new_escaped}|github.com/\1/${old_escaped}|g" "$f"
        # Revert package references (e.g., testname.UIFS -> ghega.UIFS)
        sed -i "s/${new_escaped}\./${old_escaped}./g" "$f"
    fi
done

echo "=== Product Rename Summary ==="
echo "OLD: ${OLD_NAME}"
echo "NEW: ${NEW_NAME}"
echo ""

if [ ${#changed[@]} -eq 0 ]; then
    echo "No files were modified."
else
    echo "Files changed (${#changed[@]}):"
    for f in "${changed[@]}"; do
        echo "  - ${f}"
    done
    echo ""
    echo "Review the diff with: git diff"
fi

exit 0
