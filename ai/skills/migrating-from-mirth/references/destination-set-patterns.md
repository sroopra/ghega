# Destination Set Patterns

Mirth channels use `destinationSet`, `router.routeMessage`, and `responseMap`
to control which destinations receive a message and how responses are handled.
Ghega channels currently support a single destination, so these patterns must
be rewritten into routing configuration or multiple channels.

## Common Patterns

### 1. Single Destination with Conditional Routing

**Mirth:**
```
if (msg['MSH']['MSH.9']['MSH.9.1'] == 'ADT') {
  destinationSet['ADT Destination'] = true;
} else {
  destinationSet['Generic Destination'] = true;
}
```

**Ghega Equivalent:**
Create two Ghega channels, one for each destination, and filter at the source
using a channel-level filter or a CEL condition. Alternatively, use a single
channel with a conditional mapping and route downstream via an orchestrator.

### 2. Broadcast to All Destinations

**Mirth:**
```
for (var key in destinationSet) {
  destinationSet[key] = true;
}
```

**Ghega Equivalent:**
Broadcasting requires a fan-out pattern. In Ghega, implement this as:
- A source channel that publishes to a message bus topic, or
- A custom Go transformation that emits multiple messages, or
- Multiple channels that share the same source configuration

### 3. Response Map Handling

**Mirth:**
```
var response = responseMap.get('HTTP Destination');
logger.info('Status: ' + response.getStatus());
```

**Ghega Equivalent:**
Ghega destinations return responses through the channel execution context.
Access response metadata via the destination configuration or a custom Go
plugin that inspects the response after the send operation.

### 4. Router Message Dispatch

**Mirth:**
```
router.routeMessage('channel-id', msg);
```

**Ghega Equivalent:**
Cross-channel dispatch is not a built-in single-channel feature. Use:
- An explicit channel-to-channel trigger in an orchestration layer, or
- A message bus that allows one channel to publish and another to subscribe

## Decision Table

| Mirth Pattern | Ghega Approach |
|---------------|----------------|
| Conditional destinationSet | Multiple channels with filters, or orchestrated routing |
| Broadcast destinationSet | Fan-out via bus, or multiple channels |
| responseMap read | Destination response config or custom Go plugin |
| router.routeMessage | Orchestration layer or pub/sub bus |

## Migration Notes

- The migration report flags every `destinationSet` and `router` call as a
  `destination_dispatch` pattern with severity `medium`.
- When rewriting, document the intended routing semantics so the Ghega
  deployment can be validated.
- If a Mirth channel has multiple destinations, the migration report already
  warns that only the primary destination was converted.

## Safety

Do not embed destination URLs, credentials, or endpoint names from production
systems in migration documentation. Use placeholder values such as
`https://example.com/endpoint` and `GHEGA-TEST-DEST`.
