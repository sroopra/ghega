# FHIR Bundle Patterns

This document describes common patterns for handling FHIR Bundles in Ghega channels.

## Receiving a Bundle

When a FHIR source channel receives a Bundle:

1. Parse the `Bundle.type` field
2. Iterate over `Bundle.entry`
3. Validate each `entry.resource` type
4. Route or transform based on resource type

### Batch Bundle Processing

In a `batch` Bundle, each entry is processed independently.

- One entry failing does not affect others
- Responses are returned as a `Bundle` with matching `entry.response` elements
- Use when importing a set of unrelated resources

### Transaction Bundle Processing

In a `transaction` Bundle, all entries succeed or fail together.

- The server must process atomically
- If any entry fails, the entire Bundle is rejected
- Use when data consistency is required across multiple resources

## Sending a Bundle

When constructing a Bundle to send to a FHIR destination:

1. Set `Bundle.type` to `batch` or `transaction`
2. Populate each `entry.resource` with the transformed data
3. Optionally set `entry.request.method` and `entry.request.url` for RESTful semantics
4. Set the `Content-Type` header to `application/fhir+json`

### Example Bundle Structure (Synthetic Data)

```json
{
  "resourceType": "Bundle",
  "type": "batch",
  "entry": [
    {
      "resource": {
        "resourceType": "Patient",
        "identifier": [{ "value": "SYNTH-001" }],
        "name": [{ "family": "TESTPATIENT", "given": ["ONE"] }]
      },
      "request": {
        "method": "POST",
        "url": "Patient"
      }
    }
  ]
}
```

## Error Handling

If a Bundle entry fails validation:

- For `batch`: return `OperationOutcome` in the matching response entry
- For `transaction`: reject the entire Bundle with a single `OperationOutcome`

Log the error type and entry index, but do not log the full resource payload
if it may contain PHI.

## Performance Considerations

- Large Bundles (hundreds of entries) may require streaming or chunked processing
- Consider setting a maximum entry count per Bundle
- Validate resource types before attempting full schema validation
