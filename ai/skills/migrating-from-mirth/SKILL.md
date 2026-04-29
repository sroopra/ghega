---
name: migrating-from-mirth
description: >
  Use when a user wants to migrate from Mirth, perform a Mirth import,
  convert Mirth XML, or move from Mirth to Ghega.
  Use when someone asks about Mirth channel conversion, Mirth Connect export,
  or translating Mirth connectors into Ghega channels.
  Use when the topic involves migration reports, auto-converted channels,
  rewrite tasks, or transformer classification from a Mirth source.
license: Apache-2.0
---

# Migrating from Mirth

This skill guides the migration of Mirth Connect channels to Ghega using the
`ghega migrate mirth` command and the resulting migration reports.

## When to Use

- Running `ghega migrate mirth <export-dir> --out <output-dir>`
- Interpreting migration reports for a set of exported channels
- Deciding whether a channel is ready to use or needs manual rewrite work
- Classifying JavaScript and E4X patterns found in Mirth transformers
- Generating typed rewrite tasks for patterns that cannot be auto-converted

## Migration Overview

The migration CLI reads Mirth channel XML files from an export directory,
converts structural elements (connectors, metadata) to Ghega channel YAML,
and classifies every transformer step into one of three dispositions:

1. **Auto-convertible** — mapped directly to a Ghega mapping
2. **Needs rewrite** — requires a typed rewrite task (human or AI)
3. **Unsupported** — not yet supported by Ghega

## Interpreting Migration Reports

Migration produces two levels of reports:

- **Summary report** (`migration-report.yaml` in the output root):
  - `generatedAt`: timestamp in RFC3339
  - `channels`: list of `ChannelSummary` with `name`, `status`,
    `rewriteTasksCount`, and `warningsCount`
  - `totalChannels`, `totalAutoConverted`, `totalNeedsRewrite`,
    `totalUnsupported`

- **Per-channel report** (`<channel-name>/migration-report.yaml`):
  - `channelName`: sanitized Ghega-compatible name
  - `originalName`: name from the Mirth XML
  - `status`: one of `auto-converted`, `needs-rewrite`, `unsupported`, `mixed`
  - `autoConverted`: list of successfully migrated elements
  - `needsRewrite`: list of typed rewrite task items
  - `unsupported`: list of features that cannot be migrated yet
  - `warnings`: human-readable warnings generated during conversion

A channel with status `auto-converted` can typically be deployed as-is.
Status `needs-rewrite` or `mixed` means the `rewrite-tasks.yaml` file in the
channel directory must be addressed before the channel is production-ready.

## Auto-Converted vs. Typed Rewrite

### Auto-Converted Elements

- Source connectors mapped to Ghega source types (`mllp`, `http`, `file`, `db`)
- Destination connectors mapped to Ghega destination types (`http`, `file`, `db`)
- Simple field assignments (`msg['PID']['PID.5']['PID.5.1'] = 'VALUE'`)
- Simple copy mappings (`msg['PID']['PID.3'] = msg['PV1']['PV1.19']`)
- Static value assignments

### Elements That Need Typed Rewrite

- Conditional logic (`if`, `switch`)
- Loop constructs (`for`, `for each`, `while`)
- Logger and debug statements
- Destination dispatch calls (`destinationSet`, `router.routeMessage`)
- Complex right-hand-side expressions in field assignments
- E4X/XML manipulation

### Unsupported Elements

- External library or function calls not native to Mirth or Ghega
- Advanced E4X features (descendants, namespace queries)
- Certain connector types not yet mapped

## Classifying JavaScript and E4X Patterns

The migration tool performs static regex-based analysis. It does **not**
execute JavaScript. Patterns are classified by category:

| Category | Typical Pattern | Disposition |
|----------|-----------------|-------------|
| `field_assignment` | `msg['SEG']['SEG.1'] = ...` | Auto-convertible if RHS is simple; NeedsRewrite if complex |
| `conditional` | `if (...) { ... }` | NeedsRewrite |
| `loop` | `for (...) { ... }` | NeedsRewrite |
| `logger` | `logger.info(...)` | NeedsRewrite |
| `destination_dispatch` | `destinationSet(...)` | NeedsRewrite |
| `e4x_manipulation` | `new XML(...)`, `..*`, `.@` | Unsupported |
| `external_call` | `myLib.doWork(...)` | Unsupported |

When reviewing a migration report, examine each `needsRewrite` item:
- Read the `description` and `category` fields
- Determine whether the pattern can be expressed as a Ghega mapping,
  a CEL expression, or requires custom Go code
- Produce a typed rewrite task that replaces the original JavaScript

## Generating Typed Rewrite Tasks

For every `needsRewrite` or `unsupported` item, create a clear task in the
`rewrite-tasks.yaml` file for that channel. Each task must include:

- `severity`: `low`, `medium`, or `high`
- `description`: concise explanation of what the original code did and what
  the replacement should do
- `category`: the pattern category (optional but recommended)

High severity is reserved for loops, conditionals, and E4X manipulation
that affect core message routing or field population.
Medium severity applies to destination dispatch and complex assignments.
Low severity applies to logging and debug statements.

## Safety

- Never include real patient data (PHI) in migration examples or rewrite tasks
- Always use synthetic test data when illustrating transformer patterns
- Do not paste production Mirth scripts into chat or documentation verbatim
- Sanitize channel names, field values, and identifiers before sharing

## References

- See [references/migration-report-format.md](references/migration-report-format.md)
  for the full schema of reports generated by `ghega migrate mirth`.
- See [references/e4x-patterns.md](references/e4x-patterns.md) for common E4X
  patterns and their Ghega equivalents.
- See [references/code-template-migration.md](references/code-template-migration.md)
  for how Mirth code templates map to Ghega constructs.
- See [references/destination-set-patterns.md](references/destination-set-patterns.md)
  for common destination dispatch patterns and routing alternatives.
