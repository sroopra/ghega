# PHI-Safe Logging Guidance

Ghega processes healthcare messages that may contain Protected Health Information (PHI).
Logs are a common source of accidental PHI exposure. The following rules ensure that
payload bytes never enter log streams.

## Core Rules

1. **Never use `fmt.Printf` with `Envelope` or `PayloadRef` directly.**
   - `fmt.Printf("%v", envelope)` is safe because `Envelope.String()` is implemented to
     emit metadata only, but relying on default formatting (`%v` on a struct without a
     `String()` method) will print every field and is **not safe**.
   - Always call `.String()` explicitly or use the structured logger.

2. **Always log metadata fields explicitly.**
   - Good: `logger.LogMessageReceived(channelID, messageID, receivedAt, ref)`
   - Bad: `log.Printf("received: %s", rawPayloadBytes)`
   - Bad: `log.Printf("envelope: %+v", envelope)` without verifying the type implements
     a safe `String()` method.

3. **Prefer structured logging with explicit fields.**
   - Use the `internal/logging.Logger` wrapper which only accepts discrete metadata fields.
   - It is impossible to accidentally pass payload bytes because no method on the wrapper
     accepts `[]byte` or an unconstrained `string` that could be payload content.

## Safe Patterns

```go
ref := payloadref.PayloadRef{
    StorageID: "store-001",
    Location:  "s3://ghega/inbound/msg-123",
}

env := payloadref.Envelope{
    ChannelID:  "adt-inbound",
    MessageID:  "msg-123",
    ReceivedAt: time.Now(),
    Ref:        ref,
}

// Explicit safe string representation
log.Println(env.String())

// Structured logger (preferred)
logger := logging.New(os.Stdout, slog.LevelInfo)
logger.LogMessageReceived(env.ChannelID, env.MessageID, env.ReceivedAt, env.Ref)
```

## Unsafe Patterns

```go
// NEVER do this — raw payload bytes leak into logs
log.Printf("payload: %s", rawHL7Message)

// NEVER do this — default struct formatting may reveal future fields
log.Printf("envelope: %+v", envelope)

// NEVER do this — fmt.Sprintf with %v on an arbitrary struct is unsafe
fmt.Printf("%v", someStructThatMightContainPayloadBytes)
```

## Verification

The test suites in `pkg/payloadref/` and `internal/logging/` assert that:

- `PayloadRef` and `Envelope` never contain payload bytes in any string representation.
- The logger wrapper never emits payload bytes.

When adding new fields to `Envelope` or `PayloadRef`, update `String()` and `GoString()`
to ensure the new fields do not introduce PHI leakage, and extend the tests with the
new field names.
