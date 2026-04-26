---
name: reviewing-mappings
description: >
  Use when a user wants to review, validate, or improve a data mapping in Ghega.
  Use when someone asks about field mappings, transformation logic,
  channel mappings, or mapping quality assurance.
  Use when the topic involves mapping correctness, data type conversions,
  null handling, or mapping documentation.
license: Apache-2.0
---

# Reviewing Mappings

This skill guides the review and validation of data mappings in Ghega channels.

## When to Use

- Reviewing a mapping configuration before deployment
- Validating that source fields correctly map to destination fields
- Checking for data type mismatches or conversion errors
- Reviewing null or missing value handling
- Ensuring mappings do not leak PHI into logs or error messages

## Key Concepts

- **Source Field**: The field in the incoming message
- **Destination Field**: The field in the outgoing message or data store
- **Transformation**: Logic applied during mapping (e.g., date formatting, string manipulation)
- **Default Value**: Value used when source field is missing or null

## Review Checklist

1. **Correctness**: Does each source field map to the intended destination?
2. **Data Types**: Are conversions safe (e.g., string to integer, date parsing)?
3. **Null Handling**: What happens when a source field is missing?
4. **PHI Safety**: Are any mapped fields logged or exposed in error messages?
5. **Documentation**: Is the mapping purpose documented?

## Safety

Mapping review must verify that:

- No PHI is included in mapping documentation or examples
- Error messages from failed mappings do not echo field values
- Transformation logs only record metadata, not payload content

## References

- See [references/mapping-review-checklist.md](references/mapping-review-checklist.md) for a detailed review checklist.
