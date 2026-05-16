# Core Engineering Beliefs

These are the durable principles that guide Ghega's engineering decisions. They are unlikely to change and inform all ADRs and design choices.

## 1. PHI safety is non-negotiable

Protected Health Information must never leak into logs, error messages, or external systems. Safety is enforced at the type level, not by process discipline. See ADR-009 and [phi-logging-guidance.md](phi-logging-guidance.md).

## 2. The runtime carries no Java or JavaScript

The Go runtime must remain free of JVM and JS/Node dependencies. This reduces attack surface, simplifies deployment, and keeps the binary self-contained. The UI is the only TypeScript exception, and it runs in the browser, not the server. See ADR-003 and ADR-004.

## 3. Typed channels over scripted channels

Channel definitions are typed YAML with deterministic tests. There is no embedded scripting engine. Transformations use CEL expressions or Go-native mapping engines. This makes channels testable, diffable, and auditable.

## 4. Migration over emulation

Ghega helps teams move away from Mirth. It does not emulate Mirth's behavior or claim compatibility. Migration tooling translates Mirth structure into Ghega-native typed channels. See ADR-001.

## 5. AI assists, AI does not execute

AI skills guide channel creation, testing, migration, and debugging. Skills are documentation artifacts that help agents produce correct output. They never execute in the message processing path.

## 6. Tests are deterministic and fast

Channel tests run without network, database, or external service dependencies. Test fixtures use synthetic data only. The full test suite must complete in seconds, not minutes.

## 7. The repository is the system of record

All knowledge — architecture decisions, plans, beliefs, specifications — lives in the repository. If it is not checked in, versioned, and discoverable, it effectively does not exist.
