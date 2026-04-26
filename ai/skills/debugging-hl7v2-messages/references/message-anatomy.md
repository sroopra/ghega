# HL7v2 Message Anatomy Reference

## Common Segments

### MSH — Message Header

| Field | Name | Example |
|-------|------|---------|
| MSH-1 | Field Separator | `\|` |
| MSH-2 | Encoding Characters | `^~\\&` |
| MSH-3 | Sending Application | `GHEGA-TEST-SENDER` |
| MSH-5 | Receiving Application | `GHEGA-TEST-RECEIVER` |
| MSH-9 | Message Type | `ADT^A01` |
| MSH-10 | Message Control ID | `MSG-TEST-001` |
| MSH-11 | Processing ID | `P` |
| MSH-12 | Version ID | `2.5.1` |

### PID — Patient Identification

| Field | Name | Example (Synthetic) |
|-------|------|---------------------|
| PID-1 | Set ID | `1` |
| PID-3 | Patient Identifier List | `999999999^^^GHEGA-TEST` |
| PID-5 | Patient Name | `TESTPATIENT^ONE` |
| PID-7 | Date/Time of Birth | `19800101` |
| PID-8 | Administrative Sex | `M` |

### PV1 — Patient Visit

| Field | Name | Example (Synthetic) |
|-------|------|---------------------|
| PV1-1 | Set ID | `1` |
| PV1-2 | Patient Class | `I` |
| PV1-3 | Assigned Patient Location | `TESTWARD^01^01` |
| PV1-7 | Attending Doctor | `12345^TESTDOCTOR^ONE` |
| PV1-17 | Admitting Doctor | `12345^TESTDOCTOR^ONE` |
| PV1-19 | Visit Number | `VISIT-TEST-001` |

## Encoding Characters Explained

The standard encoding characters in `MSH-2` are:

- `^` — Component separator
- `~` — Repetition separator
- `\` — Escape character
- `&` — Subcomponent separator

## Example Synthetic Message (ADT_A01)

```
MSH|^~\&|GHEGA-TEST-SENDER|GHEGA-TEST-FACILITY|GHEGA-TEST-RECEIVER|GHEGA-TEST-FACILITY|20240101120000||ADT^A01^ADT_A01|MSG-TEST-001|P|2.5.1
EVN|A01|20240101120000
PID|1||999999999^^^GHEGA-TEST||TESTPATIENT^ONE||19800101|M|||123 TEST LANE^^TESTVILLE^TS^12345||||||||||||||||||||||||||
PV1|1|I|TESTWARD^01^01||||12345^TESTDOCTOR^ONE||||||||||||||||||||||||||||||||||||||
```

## Debugging Tips

1. Count fields by counting separators. MSH-1 is always `|`, so MSH-2 starts after the first `|`.
2. Verify `MSH-2` matches the actual delimiters used in the rest of the message.
3. Check that segment terminators are `\r` (carriage return), not `\n`.
4. Validate message type in `MSH-9.1` and trigger event in `MSH-9.2`.
