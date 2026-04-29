# Custom Go Plugin Guide

When a Mirth pattern cannot be expressed as a Ghega mapping or CEL
expression, implement it as a custom Go transformation plugin.

## When to Use a Custom Go Plugin

- Loop constructs that iterate over repeating segments
- Complex conditional logic with many branches
- External lookups or database queries
- Domain-specific calculations
- E4X features that require manual XML traversal

## Plugin Interface

A custom transform plugin is a Go package that implements the transform
interface expected by Ghega. The plugin receives the message payload and
channel context, performs the transformation, and returns the modified
payload.

### Minimal Plugin Structure

```go
package myplugin

// Transform applies the plugin logic to the input payload.
// It must be safe for concurrent use.
func Transform(payload []byte, ctx Context) ([]byte, error) {
    // Implementation goes here.
    return payload, nil
}
```

Do not include the above Go code snippet in any skill markdown file as an
executable script block. It is shown here for conceptual reference only.

## Registration

Register the plugin in the channel YAML:

```yaml
mappings:
  - source: OBX
    target: OBX
    transform: custom
    plugin: myplugin-v1
```

The plugin name (`myplugin-v1`) must match the registered plugin key in the
Ghega runtime.

## Development Workflow

1. Identify the Mirth pattern from the migration report.
2. Write the equivalent logic in Go.
3. Unit-test the plugin with synthetic HL7v2 messages.
4. Register the plugin in the Ghega runtime configuration.
5. Reference the plugin in the channel YAML mapping.
6. Validate the channel with `ghega validate`.

## Safety

- Never embed credentials or secrets in plugin code.
- Use only synthetic test data in unit tests.
- Ensure the plugin does not log message contents at INFO level or higher.
- Validate that the plugin handles missing fields gracefully.

## Mapping Multiple Patterns

If a single Mirth transformer step contains several patterns (e.g. a loop
with a conditional inside), it is often cleaner to combine them into one
custom plugin rather than chaining multiple mappings.

## Example Scenario

**Mirth step:**
Iterate over every `OBX` segment and uppercase the `OBX.5` field only if
`OBX.2` equals `TX`.

**Custom plugin approach:**
Implement a Go function that parses the message, loops over `OBX` segments,
applies the conditional uppercase, and serializes the result.

**Channel mapping:**
```yaml
mappings:
  - source: ALL_SEGMENTS
    target: ALL_SEGMENTS
    transform: custom
    plugin: obx-conditional-uppercase-v1
```

## Testing

Before deploying a custom plugin:

- [ ] Plugin compiles without errors
- [ ] Unit tests pass with synthetic data
- [ ] Plugin handles empty or malformed input gracefully
- [ ] No PHI appears in test fixtures
- [ ] Plugin is registered in the runtime before the channel is loaded
