# Channel Test Generation Checklist

## Before Generating Tests

- [ ] Channel YAML exists and is valid
- [ ] Input fixtures use only synthetic (non-PHI) data
- [ ] Expected outputs have been reviewed for PHI leakage

## Test Coverage Requirements

- [ ] At least one test for each message type supported by the channel
- [ ] At least one test for error paths (malformed input, missing fields)
- [ ] At least one test for ACK behavior (if MLLP)
- [ ] Tests verify deterministic output (same input = same output)

## Fixture Guidelines

- Use `.hl7` extension for HL7v2 fixtures
- Use `.json` extension for FHIR or JSON fixtures
- Include a header comment indicating synthetic data origin
- Keep fixtures concise; test one scenario per fixture

## Example Fixture Header

```
# Synthetic test fixture for Ghega channel testing.
# No real patient data. Generated for deterministic test coverage.
```

## Output Validation

After running generated tests:

- [ ] All tests pass
- [ ] No PHI appears in test output or logs
- [ ] Diff output (if any) does not expose sensitive fields
