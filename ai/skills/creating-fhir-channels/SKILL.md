---
name: creating-fhir-channels
description: >
  Use when a user wants to create, configure, or modify a FHIR channel in Ghega.
  Use when someone asks about FHIR channels, FHIR APIs, FHIR servers,
  FHIR listeners, FHIR senders, R4, or DSTU2.
  Use when the topic involves FHIR resource types, bundle handling,
  content-type negotiation, or FHIR channel YAML definitions.
license: Apache-2.0
---

# Creating FHIR Channels

This skill guides the creation and configuration of FHIR (Fast Healthcare Interoperability Resources)
channels in Ghega. FHIR channels exchange structured healthcare data using RESTful APIs
and standardized resource formats.

## When to Use

- Building a new FHIR source or destination channel
- Configuring FHIR resource type filtering or transformation
- Handling FHIR Bundles (batch, transaction, search-set)
- Setting up content-type negotiation (JSON vs XML)
- Working with FHIR R4 or DSTU2 endpoints

## Key Concepts

- **FHIR Resource**: A single unit of healthcare data (e.g., Patient, Encounter, Observation)
- **Bundle**: A collection of resources sent or received together
- **Content-Type**: `application/fhir+json` or `application/fhir+xml`
- **Channel YAML**: Defines source, destination, and mapping for the channel

## Source vs Destination

### FHIR Source (Listener)

- Binds to a base URL path (e.g., `/fhir/R4`)
- Accepts `POST`, `PUT`, `GET`, and `DELETE` operations
- Validates incoming `Content-Type` header
- Optionally validates resource structure before routing

### FHIR Destination (Sender)

- Targets an external FHIR server base URL
- Sets outgoing `Content-Type` and `Accept` headers
- Handles HTTP status codes and retries
- Maps internal messages to FHIR resources before sending

## Bundle Handling

| Bundle Type | Purpose | Typical HTTP Method |
|-------------|---------|---------------------|
| `batch` | Process a set of resources independently | `POST` |
| `transaction` | Process atomically (all succeed or fail) | `POST` |
| `search-set` | Return search results | `GET` |
| `history` | Return resource history | `GET` |

When receiving a Bundle, the channel should:

1. Validate the `Bundle.type` field
2. Iterate over `entry[].resource` elements
3. Route each resource to the appropriate mapping or destination

## Content-Type Negotiation

Always request and respond with the correct media type:

- JSON: `application/fhir+json; fhirVersion=4.0`
- XML: `application/fhir+xml; fhirVersion=4.0`

For DSTU2, use `fhirVersion=1.0`.

## Safety

Never include real patient data (PHI) in channel configurations or examples.
Always use synthetic test data for examples and fixtures.

## References

- See [references/fhir-resource-types.md](references/fhir-resource-types.md) for common resource type reference.
- See [references/bundle-patterns.md](references/bundle-patterns.md) for bundle processing patterns.
