---
name: creating-mllp-channels
description: >
  Use when a user wants to create, configure, or modify an MLLP channel in Ghega.
  Use when someone asks about MLLP listeners, MLLP senders, MLLP endpoints,
  HL7v2 over TCP, or bidirectional MLLP channels.
  Use when the topic involves port configuration, LLP framing, ACK behavior,
  or MLLP channel YAML definitions.
license: Apache-2.0
---

# Creating MLLP Channels

This skill guides the creation and configuration of MLLP (Minimal Lower Layer Protocol)
channels in Ghega. MLLP is the standard transport for HL7v2 messages over TCP.

## When to Use

- Building a new MLLP listener or sender channel
- Configuring ACK (acknowledgment) behavior
- Setting up bidirectional MLLP communication
- Troubleshooting MLLP connection or framing issues

## Key Concepts

- **MLLP Framing**: Messages are wrapped with `0x0B` (start) and `0x1C 0x0D` (end) bytes
- **ACK Modes**: `AA` (application accept), `AR` (application reject), `AE` (application error)
- **Channel YAML**: Defines source, destination, and mapping for the channel

## Safety

Never include real patient data (PHI) in channel configurations or examples.
Always use synthetic test data for examples and fixtures.

## References

- See [references/safety.md](references/safety.md) for PHI-safe example guidelines.
