# Ghega AI Skills Plan v8

## Purpose

Ghega is AI-assisted from the first developer interaction.

The goal is not to put an LLM into the runtime message path. The goal is to help users create, test, review, migrate, and debug channels faster and more safely.

Skills are a first-class part of the Ghega project.

## Strategic authoring experience

The target experience is:

```text
Describe desired integration
    ↓
Agent loads Ghega Skill
    ↓
Agent uses deterministic Ghega generators where useful
    ↓
Agent produces channel YAML + tests + fixtures
    ↓
Developer reviews diff
    ↓
ghega channel test passes
    ↓
ghega channel deploy
    ↓
edit-to-first-message under 5 seconds in dev
```

This is a core differentiator versus Mirth.

## Skill format

Ghega Skills are Agent Skills, following the standard skills format.

A skill directory must contain:

```text
SKILL.md
references/
scripts/
assets/
```

Only `SKILL.md` is mandatory.

`SKILL.md` must start with YAML frontmatter.

Required fields:

```yaml
---
name: creating-mllp-channels
description: Creates Ghega channel YAML, tests, and fixtures for HL7v2 MLLP integrations. Use when a user wants a new MLLP listener, MLLP sender, ADT feed, ORU feed, or HL7v2 channel.
license: Apache-2.0
---
```

Rules:

- `name` uses lowercase letters, numbers, and hyphens.
- `description` must clearly say what the skill does and when to use it.
- The description is the primary trigger mechanism.
- Skill bodies must be concise.
- Detailed content goes into `references/`.
- Scripts go into `scripts/`.
- Templates go into `assets/`.

## Skills vs generators

Do not confuse Skills and generators.

### Agent Skills

Located in:

```text
.claude/skills/
.agents/skills/
ai/skills/
```

Agent Skills:

- are loaded by an LLM-capable agent
- teach the agent how to perform a task
- follow the Agent Skills standard
- are not deterministic by themselves
- are not invoked like normal CLI commands

### Ghega generators

Located in:

```text
internal/cli/generate/
```

Generators:

- are deterministic Go code
- work without an LLM
- create channel skeletons, tests, fixtures, and examples
- are invoked through the Ghega CLI

Example:

```bash
ghega generate channel mllp-to-http   --name adt-a01-to-ris   --message-type ADT_A01   --out ./channels/adt-a01-to-ris
```

A Skill may instruct an agent to use a generator, then customize and review the output.

## Required Phase 1 Skills

Phase 1 must include at least these skills:

```text
creating-mllp-channels
generating-channel-tests
debugging-hl7v2-messages
reviewing-mappings
planning-replays
understanding-ghega-codebase
```

## Required Phase 2 Skills

Phase 2 should add:

```text
creating-http-channels
creating-file-and-sftp-channels
creating-db-channels
diagnosing-failed-deployments
diagnosing-operations
reviewing-security
```

## Required Phase 4 Skills

Migration-focused skills:

```text
migrating-from-mirth
writing-typed-rewrite-tasks
```

E4X conversion is part of `migrating-from-mirth`, not a standalone skill unless later evidence shows it should be split.

## Required Phase 6 Skills

FHIR-focused skills:

```text
creating-fhir-channels
mapping-hl7v2-to-fhir
```

The `mapping-hl7v2-to-fhir` skill must reference the HL7 v2-to-FHIR mapping guidance where practical and must document deviations.

## Recommended skill structure

Example:

```text
.claude/skills/creating-mllp-channels/
  SKILL.md
  references/
    channel-defaults.md
    mllp-patterns.md
    ack-policy.md
    testing-checklist.md
    safety.md
  assets/
    channel-template.yaml
    test-fixture-template/
```

Mirror or symlink to:

```text
.agents/skills/creating-mllp-channels/
ai/skills/creating-mllp-channels/
```

The project should document which directory is canonical.

Recommended canonical source:

```text
ai/skills/
```

Recommended compatibility symlinks:

```text
.claude/skills -> ../ai/skills
.agents/skills -> ../ai/skills
```

If symlinks are problematic on Windows, duplicate through a build/sync script.

## Skill descriptions must be pushy

Skill descriptions must include trigger phrases.

Bad:

```yaml
description: Helps with MLLP channels.
```

Good:

```yaml
description: Creates Ghega channel YAML, tests, and fixtures for HL7v2 MLLP integrations. Use when a user wants a new MLLP listener, MLLP sender, ADT feed, ORU feed, HL7v2 channel, or says "create a channel" involving MLLP.
```

## Skill: creating-mllp-channels

Purpose:

- Create HL7v2 MLLP listener/sender channels.
- Scaffold channel YAML.
- Scaffold tests.
- Recommend ACK policy.
- Recommend idempotency policy.
- Recommend message size limits.
- Recommend backpressure settings.

Must use safe defaults:

- ACK after persist.
- Idempotency on MSH-10, 24h default.
- Payload limit set explicitly.
- No PHI in generated examples.
- Tests include valid and invalid messages.
- Lenient HL7 mode unless strict conformance is requested.

May invoke:

```bash
ghega generate channel mllp-to-http ...
ghega channel validate ...
ghega channel test ...
```

## Skill: generating-channel-tests

Purpose:

- Generate test fixtures for existing channel YAML.
- Generate positive tests.
- Generate negative tests.
- Generate idempotency tests.
- Generate replay tests.
- Generate shadow comparison tests.

Must produce:

- input messages
- expected ACK/NACK
- expected destination payload
- expected status
- expected audit events where relevant

## Skill: debugging-hl7v2-messages

Purpose:

- Help users debug parse failures, ACK/NACK issues, MLLP framing, field path issues, charset issues, timestamp issues, and destination failures.

Workflow:

1. Inspect metadata first.
2. Avoid payload access unless necessary.
3. Check channel revision.
4. Check ACK policy.
5. Check parser mode.
6. Check MLLP framing.
7. Check destination attempts.
8. Check retry/dead-letter status.
9. Check idempotency result.
10. Recommend next action.

Must not expose PHI unless the user has permission and purpose.

## Skill: reviewing-mappings

Purpose:

- Review typed mappings, CEL expressions, JSON templates, HL7 templates, and response transforms.

Checks:

- required fields
- type mismatches
- missing null handling
- wrong HL7 path
- unsafe assumptions about repetitions
- timestamp policy
- charset issues
- FHIR Bundle reference consistency
- idempotency key consistency

## Skill: planning-replays

Purpose:

- Help users choose redeliver vs reprocess original revision vs reprocess current revision vs replay as new.
- Detect downstream channel effects.
- Detect idempotency interactions.
- Recommend rate limits.
- Recommend safe time windows.

Description must trigger on:

- replay
- redeliver
- reprocess
- retry
- dead letter
- stuck messages
- failed messages
- recovery
- "what should I do with these messages?"

References:

```text
references/replay-decision-tree.md
references/idempotency-interactions.md
references/downstream-scope.md
```

## Skill: understanding-ghega-codebase

Purpose:

- Help agents and contributors understand the Ghega repository.
- Explain where runtime, connectors, mapping, storage, UI, skills, and migration code live.
- Explain architecture boundaries.
- Explain no-Java/no-JavaScript runtime rules.

References:

```text
references/architecture-tour.md
references/where-to-find.md
references/phase-map.md
```

## Skill: migrating-from-mirth

Purpose:

- Interpret Mirth XML migration reports.
- Explain what was auto-converted.
- Explain what requires typed rewrite.
- Classify JavaScript/E4X patterns.
- Generate typed rewrite tasks.
- Suggest tests.

Important:

- Do not execute JavaScript.
- Do not preserve JavaScript as runnable code.
- E4X conversion is a sub-procedure of this skill.

References:

```text
references/e4x-patterns.md
references/code-template-migration.md
references/destination-set-patterns.md
references/migration-report-format.md
```

## Skill: reviewing-security

Purpose:

- Review Ghega channels and migration outputs for security risks.

Must reference:

```text
references/security-checklist.md
```

Checklist should map to:

- OWASP ASVS L2 where applicable
- HIPAA technical safeguards if US market is targeted
- GDPR Article 32
- IEC 62304 evidence where relevant
- Ghega PHI logging rules
- Ghega payload retention rules
- Ghega network policy rules
- Ghega replay permission rules

## Skill evaluation

Do not test Skills like deterministic generators.

Use two separate test categories.

### Generator tests

Command:

```bash
make test-generators
```

Verifies deterministic Go generators:

- expected files are generated
- generated channel validates
- generated tests run
- no secrets are embedded
- no PHI is introduced

### Skill evaluations

Command:

```bash
make eval-skills
```

Evaluates Agent Skills with representative prompts.

Each skill must have:

- 2 to 3 trigger prompts
- 1 negative prompt where it should not trigger
- expected behavior checklist
- grader instructions

Evaluation dimensions:

- discoverability
- correct skill triggering
- instruction following
- output usefulness
- safety behavior
- improvement over no-skill baseline

Skill evaluation may use an LLM provider. It is not a deterministic unit test.

## Skill validation

Command:

```bash
make validate-skills
```

Must verify:

- `SKILL.md` exists
- YAML frontmatter parses
- `name` exists
- `name` matches allowed pattern
- `description` exists
- `description` is specific and includes when to use the skill
- `license` exists
- referenced files exist
- no PHI in examples
- no executable JavaScript in skills
- no secrets in examples
- body is concise enough for progressive disclosure

## LLM-generated output manifest

When an agent creates or modifies channel YAML, tests, mappings, or migration tasks using a skill, it should create a manifest if practical.

Example:

```yaml
generatedBy:
  product: Ghega
  skill: creating-mllp-channels
  model: <model-name>
  modelVersion: <model-version-if-known>
  timestamp: 2026-04-26T00:00:00Z
  promptHash: sha256:...
  userReviewed: false
```

This supports review and audit, but it must not contain PHI or full prompts with PHI.

## No PHI rule

Skills must not include PHI in:

- examples
- prompts
- generated outputs
- manifests
- fixtures

Use synthetic data only.

## Commands

Add:

```bash
make validate-skills
make eval-skills
make test-generators

ghega generate channel mllp-to-http ...
ghega generate test-fixtures ...
ghega generate migration-task ...
```

Do not add `ghega skill run` as a primary workflow.

If a future command asks an LLM to use a skill, name it clearly:

```bash
ghega ai assist ...
```

Not:

```bash
ghega skill run ...
```

## MCP relationship

MCP comes later.

Skills teach agents how to work.
Generators create deterministic files.
MCP gives controlled access to a running Ghega instance.

MCP must reuse Admin API authentication and authorization.

Default policy for MCP payload access:

- no payload access by default
- per-call human approval for payload access in early versions
- all access audited

## Strategic demo

The first public demo should show:

1. User describes an ADT MLLP to HTTP/FHIR channel.
2. Agent loads `creating-mllp-channels`.
3. Agent runs `ghega generate channel`.
4. Agent customizes typed mapping.
5. Agent generates tests using `generating-channel-tests`.
6. `ghega channel test` passes.
7. `ghega serve` or `ghega channel deploy` runs.
8. First message is processed in under 5 seconds in dev.
9. UI shows the message and destination attempt.

This is the AI-native authoring story:

> From idea to tested channel in minutes, without JavaScript transformers.
