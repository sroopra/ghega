# Code Template Migration

Mirth code templates are reusable JavaScript functions shared across channels.
In Ghega, equivalent reuse is achieved through mappings, CEL expressions, or
custom Go packages.

## What Is a Mirth Code Template?

A code template is a library of JavaScript functions stored in Mirth Connect.
Channels reference these functions in transformer steps, filters, or
preprocessor scripts. Common uses include:

- Normalizing identifiers
- Formatting dates
- Looking up reference values
- Logging helpers

## Migration Strategy

### 1. Identify Usage

For each code template referenced by a migrated channel, list every call site
in the transformer steps. The migration report flags `external_call` patterns
that may originate from code templates.

### 2. Classify the Logic

| Template Type | Ghega Equivalent |
|---------------|------------------|
| Simple string manipulation | Mapping with `transform: cel` or built-in transform |
| Date formatting | Mapping with `transform: date` or CEL date functions |
| Lookup table | External reference data loaded via channel config or custom Go |
| Logging utility | Structured logging configured at the channel or system level |
| Complex business rule | Custom Go function imported as a channel plugin |

### 3. Replace with Mappings

If the code template only performs field-to-field logic, replace it with a
Ghega mapping in `channel.yaml`:

```yaml
mappings:
  - source: PID-3.1
    target: patient_mrn_normalized
    transform: cel
    expression: "source.trim().toUpperCase()"
```

### 4. Replace with Custom Go

If the template contains logic that cannot be expressed as a mapping or CEL
expression, implement it as a custom Go function. Register the function as a
transform plugin and reference it in the channel:

```yaml
mappings:
  - source: PID-7
    target: patient_dob_formatted
    transform: custom
    plugin: datefmt-v1
```

### 5. Remove Unused Templates

If a code template is referenced but never called in the transformed channel,
remove it from the rewrite task list. The migration report only flags patterns
that actually appear in executable transformer steps.

## Mapping Checklist

- [ ] Every code template call site is identified
- [ ] The replacement is expressed as a mapping, CEL expression, or custom Go
- [ ] No JavaScript remains in the final channel configuration
- [ ] Test data for the replacement is synthetic

## Safety

Do not include the original JavaScript source of code templates in Ghega
channel files or documentation unless it is fully sanitized. Avoid copying
production helper functions that may embed credentials, URLs, or PHI.
