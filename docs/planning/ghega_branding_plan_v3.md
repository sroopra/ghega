# Ghega Branding Plan v3

## Status

The selected working name is **Ghega**.

Ghega should be used consistently in the project from now on, but trademark and namespace clearance are still required before public launch.

## Product name

```text
Ghega
```

## One-line description

Ghega is an open-source healthcare integration engine for teams that want to move beyond Mirth.

## Positioning

Ghega is infrastructure.

It is not:

- a care management product
- a patient engagement product
- an EHR module
- a generic workflow automation tool
- a Mirth clone

It is:

- an open-source healthcare integration engine
- typed
- testable
- observable
- migration-friendly
- AI-assisted
- runtime-safe
- built for HL7v2, FHIR, MLLP, HTTP, SFTP, files, and database integrations

## Naming rationale

Ghega is inspired by Carl Ritter von Ghega and the Semmering Railway.

The story:

> Ghega built reliable routes through difficult terrain. Ghega builds reliable routes through difficult healthcare systems.

Use this as a subtle infrastructure metaphor. Do not overdo the railway theme.

## Tagline

Primary:

```text
Move beyond Mirth.
```

Subline:

```text
Typed healthcare integration. AI-assisted operations.
```

Longer version:

```text
Move beyond Mirth with typed channels, deterministic tests, durable message processing, and AI-assisted operations.
```

## Brand voice

Ghega should sound:

- calm
- technical
- trustworthy
- operationally serious
- migration-friendly
- open-source friendly
- not hype-driven
- not “AI magic”
- not childish anti-Mirth messaging

Avoid:

- revolutionary
- magical
- kills Mirth
- fully automated clinical integration
- AI-powered everything
- care management wording

## Canonical names

```yaml
productName: Ghega
productSlug: ghega
cliName: ghega
repoName: ghega
githubOrg: ghega
containerImage: ghcr.io/ghega/ghega
helmChart: ghega
k8sNamespaceExample: ghega-system
envPrefix: GHEGA_
configDir: ~/.config/ghega/
uiName: Ghega Console
skillsName: Ghega Skills
mcpServerName: Ghega MCP Server
```

## Repository

Preferred:

```text
github.com/ghega/ghega
```

Fallback:

```text
github.com/ghega-project/ghega
```

Suggested organization layout:

```text
github.com/ghega/ghega
github.com/ghega/ghega-docs
github.com/ghega/ghega-examples
github.com/ghega/ghega-charts
```

Start with a monorepo unless there is a clear reason to split.

## Container registry

Canonical:

```text
ghcr.io/ghega/ghega
```

Optional later:

```text
docker.io/ghega/ghega
quay.io/ghega/ghega
```

Air-gapped customers should be able to retag the image into their private registry.

## CLI

Use one binary:

```bash
ghega
```

Do not use:

```text
ghegad
ghegactl
```

Examples:

```bash
ghega serve
ghega channel validate ./channels/adt/channel.yaml
ghega channel deploy ./channels/adt/channel.yaml --env dev
ghega message search --channel adt --status failed
ghega migrate mirth ./mirth-export --out ./migrated
ghega audit verify
ghega mcp serve
```

## Environment variables

Use:

```text
GHEGA_
```

Examples:

```bash
GHEGA_DATABASE_URL=
GHEGA_LOG_LEVEL=
GHEGA_CONFIG_FILE=
GHEGA_OIDC_ISSUER=
GHEGA_PUBLIC_URL=
```

## Config directory

Use XDG first:

```text
$XDG_CONFIG_HOME/ghega/
```

Fallback:

```text
~/.config/ghega/
```

Avoid using `~/.ghega/` as the primary documented path.

## Kubernetes

Example namespace:

```text
ghega-system
```

Use placeholders in production docs:

```text
<your-namespace>
```

Labels:

```yaml
app.kubernetes.io/name: ghega
app.kubernetes.io/part-of: ghega
app.kubernetes.io/component: engine
```

## Helm

Chart name:

```text
ghega
```

OCI chart example:

```text
oci://ghcr.io/ghega/charts/ghega
```

The Helm strategy may later use an umbrella chart with subcharts, but that is an operations decision, not a branding decision.

## UI

Name:

```text
Ghega Console
```

Suggested welcome text:

```text
Ghega Console
Typed healthcare integration. AI-assisted operations.
```

Navigation is not locked by branding.

Draft grouping:

- Channels
- Messages
- Operations
- Migrations
- Settings

## Skills

Marketing name:

```text
Ghega Skills
```

Technical wording:

```text
Agent Skills shipped with Ghega
```

Important:

Ghega Skills follow the Agent Skills standard. They are not a custom Ghega-specific format.

## MCP

Name:

```text
Ghega MCP Server
```

Command:

```bash
ghega mcp serve
```

Tool names should be domain-oriented:

```text
channel.list
message.search_metadata
migration.get_report
replay.preview
audit.search
```

Do not over-prefix every tool with `ghega.` unless required by a specific client.

## README opening

```markdown
# Ghega

Ghega is an open-source healthcare integration engine for teams that want to move beyond Mirth.

It provides typed channel definitions, deterministic tests, durable message processing, replay safety, observability, migration tooling, and AI-assisted authoring for HL7v2, FHIR, MLLP, HTTP, SFTP, files, and database integrations.
```

## Public launch checklist

Before public launch:

- trademark check: US, EUIPO, UK, DPMA, Madrid
- GitHub organization availability
- GHCR namespace
- Docker Hub namespace
- domain availability
- social handle sanity check
- target-language check in German, English, French, Italian
- buyer-association test with 5 to 10 target users

Buyer-association test question:

```text
What would you assume a product called Ghega does?
```

Pass condition:

Most answers should be close to:

- healthcare integration engine
- interoperability platform
- clinical data routing
- interface engine
- infrastructure

Fail condition:

Most answers suggest:

- care management
- patient app
- EHR module
- railway/tourism product only
- unrelated infrastructure

## Rename discipline

Ghega is the selected working name, but branding should remain centralized.

Add:

```text
branding/product.yaml
```

Add:

```bash
make rename-product OLD=ghega NEW=<new-slug>
```

It should update generated examples, docs, UI strings, Helm examples, Docker labels, env var examples, and MCP examples.

Do not rewrite historical Git commits.

## Visual identity

Create placeholder:

```text
branding/visual-identity.md
```

Do not block Phase 1 on logo, colors, or visual identity.

## Final guidance for agents

Use Ghega from now on.

Read `branding/product.yaml` before generating brand-sensitive output.

Do not hard-code alternative names.

Do not create `caremeld` artifacts.
