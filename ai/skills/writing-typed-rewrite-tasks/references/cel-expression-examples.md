# CEL Expression Examples

Common Expression Language (CEL) is the preferred way to express simple
conditional and computational logic in Ghega mappings. This reference shows
patterns that frequently appear when rewriting Mirth transformer steps.

## Basic Comparisons

| Original JS | CEL Expression |
|-------------|----------------|
| `msg['PID']['PID.8'] == 'M'` | `source == 'M'` |
| `msg['PID']['PID.8'] != ''` | `source != ''` |
| `msg['PID']['PID.8'] == null` | `source == ''` (HL7 empty string semantics) |

## Ternary Operators

| Original JS | CEL Expression |
|-------------|----------------|
| `x == 'A' ? 'Alpha' : 'Beta'` | `source == 'A' ? 'Alpha' : 'Beta'` |
| `x == '' ? 'UNKNOWN' : x` | `source == '' ? 'UNKNOWN' : source` |

## String Functions

| Original JS | CEL Expression |
|-------------|----------------|
| `str.toUpperCase()` | `source.toUpperCase()` |
| `str.toLowerCase()` | `source.toLowerCase()` |
| `str.trim()` | `source.trim()` |
| `str.substring(0,3)` | `source.substring(0,3)` |
| `str.replace('old','new')` | `source.replace('old','new')` |
| `str.startsWith('PRE')` | `source.startsWith('PRE')` |

## Numeric Operations

| Original JS | CEL Expression |
|-------------|----------------|
| `parseInt(str)` | `int(source)` |
| `parseFloat(str)` | `double(source)` |
| `num + 1` | `source + 1` |
| `num.toString()` | `string(source)` |

## Date Patterns

| Original JS | CEL Expression |
|-------------|----------------|
| `date.getFullYear()` | Use `transform: date` or custom Go |
| `date.toISOString()` | Use `transform: date` or custom Go |

CEL has limited date parsing. For complex date manipulation, prefer a custom
Go plugin or the built-in `transform: date` mapping.

## Combining Conditions

```
(source == 'ADT' || source == 'A01') ? 'ADMISSION' : 'OTHER'
```

```
source.startsWith('GHEGA') && source != '' ? source : 'DEFAULT'
```

## Mapping Example

```yaml
mappings:
  - source: PID-8
    target: administrative_sex_expanded
    transform: cel
    expression: "source == 'M' ? 'Male' : (source == 'F' ? 'Female' : 'Unknown')"
```

## Limits

- CEL expressions must evaluate to a string, number, or boolean.
- CEL cannot iterate over collections; use custom Go for loops.
- CEL cannot access external systems; use custom Go for lookups.

## Safety

All values in examples are synthetic. Replace them with your own test data
when implementing real channels.
