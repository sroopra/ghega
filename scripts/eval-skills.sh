#!/usr/bin/env bash
set -euo pipefail

SKILLS_DIR="ai/skills"

echo "=== Skill Evaluation Report ==="
echo "Date: $(date -Iseconds)"
echo "Skills Directory: ${SKILLS_DIR}"
echo ""

total=0
eval_count=0
warning_count=0

if [ ! -d "${SKILLS_DIR}" ]; then
    echo "ERROR: Skills directory '${SKILLS_DIR}' not found."
    exit 1
fi

for skill_path in "${SKILLS_DIR}"/*; do
    if [ ! -d "${skill_path}" ]; then
        continue
    fi

    skill_name=$(basename "${skill_path}")
    total=$((total + 1))

    skill_file="${skill_path}/SKILL.md"
    if [ ! -f "${skill_file}" ]; then
        echo "[WARN] ${skill_name}: missing SKILL.md"
        warning_count=$((warning_count + 1))
        continue
    fi

    # Validate YAML frontmatter presence
    if ! head -n1 "${skill_file}" | grep -q '^---$'; then
        echo "[WARN] ${skill_name}: missing YAML frontmatter"
        warning_count=$((warning_count + 1))
        continue
    fi

    # Check for eval/ subdirectory
    eval_dir="${skill_path}/eval"
    if [ -d "${eval_dir}" ]; then
        eval_count=$((eval_count + 1))
        eval_files=$(find "${eval_dir}" -maxdepth 1 -type f | wc -l)
        echo "[OK]   ${skill_name}: valid (${eval_files} eval prompt(s))"
    else
        echo "[OK]   ${skill_name}: valid (no eval/ prompts — evaluation requires an LLM provider)"
    fi
done

echo ""
echo "=== Summary ==="
echo "Total Skills:   ${total}"
echo "With Eval Dir:  ${eval_count}"
echo "Warnings:       ${warning_count}"
echo ""
echo "Note: Full skill evaluation requires an LLM provider and is non-deterministic."
echo "      This report is a deterministic placeholder."

exit 0
