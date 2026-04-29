# Rewrite Pattern Guide

This decision tree helps choose the right Ghega construct for each classified
Mirth pattern.

## Quick Decision Tree

```
Pattern Category
├── field_assignment
│   ├── RHS is static value → mapping with transform: static
│   ├── RHS is simple field copy → mapping with transform: copy
│   └── RHS is complex expression → mapping with transform: cel OR custom Go
├── conditional
│   ├── 2-3 branches, simple checks → mapping with transform: cel
│   └── many branches or complex logic → custom Go plugin
├── loop
│   ├── iterate to transform repeating segment → custom Go plugin
│   └── iterate to filter → custom Go plugin or channel filter
├── logger
│   └── remove inline call, use structured channel logging
├── destination_dispatch
│   ├── single destination, unconditional → configure in channel.yaml
│   ├── conditional routing → multiple channels or orchestration
│   └── broadcast → message bus fan-out
├── e4x_manipulation
│   ├── simple bracket access → mapping
│   ├── XML construction → Go template or struct mapping
│   ├── descendant query → explicit path mapping
│   └── namespace handling → custom Go parser
└── external_call
    ├── standard library function → CEL equivalent or built-in transform
    └── domain-specific library → custom Go plugin
```

## Detailed Guidance

### Field Assignment

Always prefer the simplest transform that satisfies the requirement:

1. **Static value** — `transform: static` with a `value` field.
2. **Copy** — `transform: copy` with `source` and `target`.
3. **CEL expression** — `transform: cel` with an `expression` field.
4. **Custom Go** — `transform: custom` with a `plugin` reference.

### Conditional

CEL supports ternary operators, string functions, and basic arithmetic.
If the conditional touches multiple segments or requires external data,
use custom Go.

### Loop

Ghega mappings are declarative and do not natively loop. For repeating
segments (e.g., multiple `OBX` segments), a custom Go plugin receives the
full message and can iterate over segment collections.

### Logger

Remove all `logger.*` and `print` calls from transformer logic. Ghega
channels log execution metadata automatically. If additional observability
is needed, configure it at the channel or system level rather than in a
mapping.

### Destination Dispatch

Ghega channels have one destination. If a Mirth channel dynamically selects
destinations, model each selection as a separate channel or introduce an
orchestration layer that routes messages after the initial channel.

### E4X Manipulation

Never preserve E4X syntax. Translate every E4X operation into an equivalent
mapping or custom Go transformation. See the `e4x-patterns.md` reference in
the `migrating-from-mirth` skill for a pattern catalog.

### External Call

External library calls are unsupported. Re-implement the logic in Go or
replace it with a Ghega-native transform. If the library is a standard
utility (date formatting, string padding), a CEL expression or built-in
transform usually suffices.

## Validation Checklist

After writing rewrite tasks:

- [ ] Every `needsRewrite` item in the migration report has a corresponding task
- [ ] No JavaScript remains in the channel configuration
- [ ] All examples use synthetic data
- [ ] Severity levels match the guidelines
- [ ] Custom plugins are registered in the channel YAML
