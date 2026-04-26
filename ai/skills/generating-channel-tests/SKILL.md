---
name: generating-channel-tests
description: >
  Use when a user wants to generate tests for a Ghega channel.
  Use when someone asks about deterministic testing, channel test fixtures,
  test-driven channel development, or validating channel behavior.
  Use when the topic involves test generation, golden-file testing,
  or assertions for channel input/output pairs.
license: Apache-2.0
---

# Generating Channel Tests

This skill guides the generation of deterministic, repeatable tests for Ghega channels.

## When to Use

- Creating tests for a new or existing channel
- Generating test fixtures from sample messages
- Setting up golden-file or snapshot testing for channel mappings
- Validating that a channel produces expected output for known input

## Key Concepts

- **Deterministic Tests**: Same input always produces same output
- **Fixtures**: Static input files (HL7v2, JSON, etc.) used as test data
- **Golden Files**: Expected output files compared against actual channel output
- **Synthetic Data**: All fixtures must use non-PHI synthetic data

## Test Structure

Generated tests for a channel typically include:

1. Input fixtures in `fixtures/`
2. Expected output in `tests/expected/`
3. A test runner that loads the channel YAML and asserts behavior

## Safety

Never generate tests that embed real patient data.
Always verify fixtures contain only synthetic data before committing.

## References

- See [references/testing-checklist.md](references/testing-checklist.md) for a complete test-generation checklist.
