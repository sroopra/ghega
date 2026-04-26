# PHI Safety Guidelines for MLLP Channels

## Synthetic Data Rules

All examples, fixtures, and test data used in MLLP channel documentation must use
synthetic (non-real) data.

## Approved Synthetic Identifiers

| Field | Example Value | Notes |
|-------|---------------|-------|
| Patient Name | `TESTPATIENT,ONE` | Clearly marked as test data |
| MRN | `999999999` | Fictional, does not match real format |
| Account Number | `TEST-ACCT-001` | Prefix identifies as synthetic |
| Date of Birth | `1980-01-01` | Standard test date |

## Prohibited Patterns

- Real names (e.g., `John Smith`, `Jane Doe` unless explicitly marked synthetic)
- Real SSNs or national identifiers
- Real phone numbers
- Real addresses
- Real medical record numbers from any production system

## Example Safe HL7v2 Segment

```
PID|1||999999999^^^GHEGA-TEST||TESTPATIENT^ONE||19800101|M|||123 TEST LANE^^TESTVILLE^TS^12345||||||||||||||||||||||||||
```

## Review Checklist

Before committing any MLLP channel example or fixture:

- [ ] All patient identifiers are synthetic
- [ ] No real organization names appear
- [ ] No provider names are real
- [ ] Dates do not correspond to real events
- [ ] File contains a comment header indicating synthetic data
