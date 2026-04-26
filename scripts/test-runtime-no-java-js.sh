#!/usr/bin/env bash
#
# test-runtime-no-java-js.sh
#
# Runtime boundary verification for Ghega.
# Ensures the runtime image and source tree contain no Java or JavaScript
# execution engines.
#
# Hard constraints:
#   - No java, javac, node, npm, yarn, pnpm in the Docker image.
#   - No JavaScript engine imports in Go source (goja, otto, etc.).
#   - No .java, .js, .ts, .tsx files under internal/, pkg/, cmd/.
#
# Usage:
#   bash scripts/test-runtime-no-java-js.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE="ghcr.io/ghega/ghega:local"

ERRORS=0

# ------------------------------------------------------------------------------
# Helper functions
# ------------------------------------------------------------------------------
error() {
    echo "ERROR: $1" >&2
    ERRORS=$((ERRORS + 1))
}

info() {
    echo "INFO: $1"
}

# ------------------------------------------------------------------------------
# 1. Verify the Docker image does not contain forbidden binaries
# ------------------------------------------------------------------------------
check_docker_image() {
    info "Checking Docker image for forbidden binaries..."

    if ! command -v docker >/dev/null 2>&1; then
        info "docker not available; skipping Docker image check"
        return 0
    fi

    if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
        info "Docker image $IMAGE not found; building it now..."
        make -C "$REPO_ROOT" docker
    fi

    local binaries=("java" "javac" "node" "npm" "yarn" "pnpm")
    for bin in "${binaries[@]}"; do
        # distroless images do not have a shell, so we use docker run with --entrypoint
        # and the absolute path heuristic. If the binary exists in PATH inside the
        # container, 'command -v' would work in a shell, but distroless has no shell.
        # Instead, we search the filesystem layers for known binary names.
        if docker run --rm --entrypoint /bin/sh "$IMAGE" -c "command -v $bin" >/dev/null 2>&1; then
            error "Forbidden binary found in image: $bin"
        else
            info "  $bin ... not found (OK)"
        fi
    done

    # For distroless images, also try a filesystem-level scan via a temporary busybox container
    # that mounts the image's rootfs. This is more thorough.
    info "Performing deep filesystem scan for forbidden binaries..."
    local cid
    cid=$(docker create "$IMAGE")
    trap 'docker rm -f "$cid" >/dev/null 2>&1 || true' EXIT

    for bin in "${binaries[@]}"; do
        if docker export "$cid" | tar -tf - | grep -qE "(^|/)$bin$"; then
            error "Forbidden binary found in image filesystem: $bin"
        else
            info "  filesystem $bin ... not found (OK)"
        fi
    done

    docker rm -f "$cid" >/dev/null 2>&1 || true
    trap - EXIT
}

# ------------------------------------------------------------------------------
# 2. Verify Go packages do not import JavaScript execution engines
# ------------------------------------------------------------------------------
check_go_imports() {
    info "Checking Go imports for JavaScript execution engines..."

    local forbidden_imports=(
        "github.com/dop251/goja"
        "github.com/robertkrimen/otto"
        "github.com/duke-git/lancet/v2/javascript"
        "github.com/traefik/yaegi"
        "github.com/containous/yaegi"
        "rogchap.com/v8go"
        "github.com/augustoroman/v8"
        "github.com/ry/v8worker"
        "github.com/andybalholm/gojs"
    )

    # Check go.mod first
    if [ -f "$REPO_ROOT/go.mod" ]; then
        for imp in "${forbidden_imports[@]}"; do
            if grep -qF "$imp" "$REPO_ROOT/go.mod"; then
                error "Forbidden import found in go.mod: $imp"
            else
                info "  go.mod $imp ... not found (OK)"
            fi
        done
    fi

    # Check all .go source files (skip test files, which may reference forbidden
    # imports for validation purposes)
    while IFS= read -r -d '' file; do
        for imp in "${forbidden_imports[@]}"; do
            if grep -qF "$imp" "$file"; then
                error "Forbidden import found in $file: $imp"
            fi
        done
    done < <(find "$REPO_ROOT" -type f -name '*.go' ! -name '*_test.go' -print0)
}

# ------------------------------------------------------------------------------
# 3. Verify no forbidden file extensions under internal/, pkg/, cmd/
# ------------------------------------------------------------------------------
check_file_extensions() {
    info "Checking for forbidden file extensions under internal/, pkg/, cmd/..."

    local dirs=("$REPO_ROOT/internal" "$REPO_ROOT/pkg" "$REPO_ROOT/cmd")
    local extensions=(".java" ".js" ".ts" ".tsx")

    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            continue
        fi
        for ext in "${extensions[@]}"; do
            # Allow .js files that are explicitly documented exceptions.
            # Currently there are no exceptions.
            local found
            found=$(find "$dir" -type f -name "*${ext}" || true)
            if [ -n "$found" ]; then
                error "Forbidden file extension found: $ext"
                echo "$found" >&2
            else
                info "  $dir ... no *$ext files (OK)"
            fi
        done
    done
}

# ------------------------------------------------------------------------------
# Main
# ------------------------------------------------------------------------------
main() {
    info "Starting runtime boundary verification for Ghega"

    check_go_imports
    check_file_extensions
    check_docker_image

    if [ "$ERRORS" -gt 0 ]; then
        echo "FAILED: $ERRORS error(s) found during runtime boundary verification." >&2
        exit 1
    fi

    info "All runtime boundary checks passed."
    exit 0
}

main "$@"
