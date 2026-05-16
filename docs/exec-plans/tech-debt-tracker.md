# Tech Debt Tracker

Items tracked here represent known technical debt that should be addressed. Each item includes context and priority.

## Active Items

### TD-001: Go module path mismatch

- **Priority:** High
- **Description:** `go.mod` declares module as `github.com/sroopra/ghega` but README, Dockerfile, and branding reference `github.com/ghega/ghega`. Align before public release.
- **Impact:** Import paths will break for external consumers.
- **ADR:** None

### TD-002: Go version drift

- **Priority:** Medium
- **Description:** `go.mod` specifies Go `1.25.0` but CI and Dockerfile use Go `1.26`. Align to a single version.
- **Impact:** Potential build inconsistencies across environments.

### TD-003: Runtime placeholder

- **Priority:** Medium
- **Description:** `internal/runtime/runtime.go` is an empty placeholder. The runtime orchestration layer needs implementation.
- **Related:** Architecture layer model requires a proper Runtime layer.

### TD-004: Stubbed CLI commands

- **Priority:** Low
- **Description:** `message redeliver`, `message replay`, and `message replay-preview` CLI commands exist but are not implemented.
- **Related:** Server API endpoints for replay/redeliver are also stubbed.

### TD-005: Channel list API returns empty

- **Priority:** Medium
- **Description:** `GET /api/v1/channels` returns an empty list. Needs to be wired to `pkg/channelstore`.

### TD-006: Generated artifacts in repo

- **Priority:** Low
- **Description:** `ghega.db`, `internal/cli/ghega.db`, and `ui/dist/` are committed. Consider adding to `.gitignore` or generating at build time only.

## Resolved Items

_None yet._
