# E4X Patterns and Ghega Equivalents

Mirth channels frequently use ECMAScript for XML (E4X) to manipulate HL7v2
messages or XML payloads. Ghega does not support E4X directly, so every E4X
pattern must be rewritten.

## Common E4X Patterns

### 1. Simple Field Assignment

**Mirth E4X:**
```
msg['PID']['PID.5']['PID.5.1'] = 'TESTPATIENT';
```

**Ghega Equivalent:**
Add a mapping in `channel.yaml`:
```yaml
mappings:
  - source: static
    target: PID-5.1
    transform: static
    value: TESTPATIENT
```

### 2. Field Copy

**Mirth E4X:**
```
msg['PID']['PID.3']['PID.3.1'] = msg['PV1']['PV1.19']['PV1.19.1'];
```

**Ghega Equivalent:**
```yaml
mappings:
  - source: PV1-19.1
    target: PID-3.1
    transform: copy
```

### 3. Conditional Field Population

**Mirth E4X:**
```
if (msg['PID']['PID.8'] == 'M') {
  msg['PID']['PID.8'] = 'Male';
}
```

**Ghega Equivalent:**
Use a CEL expression or a custom mapping transform:
```yaml
mappings:
  - source: PID-8
    target: PID-8
    transform: cel
    expression: "source == 'M' ? 'Male' : source"
```

### 4. Loop Over Repeating Segments

**Mirth E4X:**
```
for each (var obx in msg..OBX) {
  obx['OBX.5'] = obx['OBX.5'].toString().toUpperCase();
}
```

**Ghega Equivalent:**
Loop constructs are not auto-converted. Rewrite as a custom Go transformation
or a Ghega mapping that operates on repeating fields with an appropriate
transform plugin.

### 5. XML Node Construction

**Mirth E4X:**
```
var doc = new XML('<result><status>OK</status></result>');
```

**Ghega Equivalent:**
Construct XML using Go templates or structured mapping output. Do not use
JavaScript XML APIs.

### 6. Descendant and Attribute Access

**Mirth E4X:**
```
var allNames = msg..['PID.5'];
var attr = node.@id;
```

**Ghega Equivalent:**
Descendant queries (`..`) and attribute accessors (`.@`) are unsupported.
Rewrite using explicit path mappings or custom Go logic that walks the
message structure.

### 7. Namespace Queries

**Mirth E4X:**
```
var ns = node.namespace();
```

**Ghega Equivalent:**
Namespace queries are unsupported. If the payload requires namespace handling,
use a custom Go transformation with an XML parser that exposes namespaces.

## Classification Summary

| E4X Feature | Disposition | Recommended Action |
|-------------|-------------|-------------------|
| Simple bracket access | Auto-convertible | Map to Ghega mapping |
| Field copy | Auto-convertible | Map to `copy` transform |
| Conditional | NeedsRewrite | CEL expression or custom Go |
| Loop | NeedsRewrite | Custom Go or mapping plugin |
| `new XML()` | Unsupported | Go template or struct mapping |
| Descendant `..` | Unsupported | Explicit path mapping |
| Attribute `.@` | Unsupported | Custom Go parser |
| Namespace | Unsupported | Custom Go parser |

## Safety

All examples use synthetic data. Replace `TESTPATIENT` with your own test
fixture values. Never paste production E4X snippets that contain real
patient identifiers.
