# Mapping Review Checklist

## Structural Review

- [ ] Every destination field has a defined source or default value
- [ ] No orphaned source fields that should be mapped but are not
- [ ] Mapping order respects dependencies (e.g., derived fields come after base fields)

## Data Integrity

- [ ] Date/time fields specify expected input and output formats
- [ ] Numeric fields validate range where applicable
- [ ] String fields specify max length or truncation behavior
- [ ] Code fields map to valid value sets or code systems

## Null and Missing Data

- [ ] Each mapping defines behavior for missing source field
- [ ] Default values are documented and appropriate
- [ ] Empty string vs. null vs. omitted field behavior is specified

## PHI and Security

- [ ] Mapping examples use synthetic data only
- [ ] No real patient identifiers in mapping comments or documentation
- [ ] Transformation error messages do not echo input values
- [ ] Audit logging records mapping name and result, not field contents

## Documentation

- [ ] Complex transformations have inline comments explaining intent
- [ ] Business rules referenced by the mapping are cited
- [ ] Known limitations or caveats are noted

## Example Safe Mapping Documentation

```yaml
# Synthetic example for Ghega mapping review.
# No real patient data.
mapping:
  - source: PID-3.1
    destination: patient_mrn
    transform: uppercase
    default: UNKNOWN
    notes: Maps synthetic test MRN to internal patient identifier.
```
