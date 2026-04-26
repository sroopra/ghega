---
name: debugging-hl7v2-messages
description: >
  Use when a user wants to debug, parse, or analyze an HL7v2 message.
  Use when someone asks about HL7v2 segment structure, field mapping errors,
  message validation, or why a message is being rejected.
  Use when the topic involves PID, PV1, OBR, OBX segments, separators,
  encoding characters, or HL7v2 message tracing.
license: Apache-2.0
---

# Debugging HL7v2 Messages

This skill guides the analysis and debugging of HL7v2 messages in Ghega.

## When to Use

- A message is being rejected or producing unexpected output
- Field mappings appear incorrect
- Segment structure needs verification
- Encoding characters or separators are suspected as the cause

## Key Concepts

- **Segments**: Logical groups (e.g., `MSH`, `PID`, `PV1`, `OBR`, `OBX`)
- **Fields**: Pipe-delimited within a segment (`|`)
- **Components**: Caret-delimited within a field (`^`)
- **Encoding Characters**: Defined in `MSH-2` (e.g., `^~\&`)
- **Separators**: Segment separator is `\r` (carriage return)

## Common Issues

1. **Wrong encoding characters**: Check `MSH-2` matches actual delimiters used
2. **Missing required segments**: Many message types require `MSH`, `PID`, `PV1`
3. **Field overflow**: Data too long for the target field width
4. **Incorrect segment order**: Some systems are strict about ordering

## Safety

When sharing HL7v2 messages for debugging:

- Replace all real patient identifiers with synthetic values
- Replace real provider names with fictional ones
- Replace real facility names with `GHEGA-TEST-FACILITY`
- Never paste production messages into logs or chat

## References

- See [references/message-anatomy.md](references/message-anatomy.md) for segment-by-segment reference.
