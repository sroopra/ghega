# Ghega Agent Harness Plan Update

## Status

This document updates the main agent-harness brief to use the new working product name: **Ghega**.

Ghega is the product name used in code examples, documentation, CLI examples, container names, UI copy, skills, and MCP examples from this point forward.

Trademark and namespace clearance are still required before public launch.

## Product identity

Product name: **Ghega**

Positioning:

> Ghega is an open-source healthcare integration engine for teams that want to move beyond Mirth.

Core promise:

> Typed healthcare integration, deterministic tests, durable message processing, safe replay, observability, migration tooling, and AI-assisted authoring.

Short tagline:

> Move beyond Mirth.

Subline:

> Typed healthcare integration. AI-assisted operations.

## Why Ghega

The name is inspired by Carl Ritter von Ghega, the Austrian engineer associated with the Semmering Railway.

The metaphor is intentional:

> Ghega built reliable routes through difficult terrain. This product builds reliable routes through difficult healthcare systems.

Use the metaphor carefully. Ghega is an infrastructure brand, not a tourism or heritage brand.

## Canonical naming

Use these values unless later changed through the central branding file.

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

## Central branding file

Add:

```text
branding/product.yaml
```

Initial content:

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

All generated docs, UI strings, Helm examples, Docker labels, CLI examples, skills, and MCP examples should read from or align with this file.

## Repository and registry

Preferred public open-source locations:

```text
github.com/ghega/ghega
ghcr.io/ghega/ghega
```

Fallback if the GitHub org is unavailable:

```text
github.com/ghega-project/ghega
ghcr.io/ghega-project/ghega
```

Do not publish to public registries before namespace and trademark checks are complete.

## Binary and CLI

Use one binary:

```bash
ghega
```

Do not use `ghegad` or `ghegactl`.

CLI style:

```bash
ghega serve
ghega channel validate ./channels/adt/channel.yaml
ghega channel deploy ./channels/adt/channel.yaml --env dev
ghega channel diff ./channels/adt/channel.yaml --env prod
ghega channel rollback adt --to-revision sha256:abc123

ghega message search --channel adt --status failed
ghega message show <message-id>
ghega message redeliver <message-id> --destination ris-api
ghega message reprocess <message-id> --revision current
ghega message replay <message-id> --as-new
ghega message replay-preview <message-id>

ghega migrate mirth ./mirth-export --samples ./samples --expected ./goldens --out ./migrated
ghega audit verify --tenant default
ghega mcp serve
```

## Environment variables

Use the `GHEGA_` prefix.

Examples:

```bash
GHEGA_DATABASE_URL=
GHEGA_LOG_LEVEL=
GHEGA_CONFIG_FILE=
GHEGA_OIDC_ISSUER=
GHEGA_PUBLIC_URL=
```

## Configuration directory

Use XDG first:

```text
$XDG_CONFIG_HOME/ghega/
```

Fallback:

```text
~/.config/ghega/
```

Avoid using `~/.ghega/` as the primary documented path.

## Kubernetes and Helm

Example namespace:

```text
ghega-system
```

Do not hard-code the production namespace to `ghega`. Examples may use:

```text
<your-namespace>
ghega-system
```

Helm chart:

```text
ghega
```

OCI chart example:

```text
oci://ghcr.io/ghega/charts/ghega
```

Common labels:

```yaml
app.kubernetes.io/name: ghega
app.kubernetes.io/part-of: ghega
app.kubernetes.io/component: engine
```

## UI naming

Use:

```text
Ghega Console
```

The UI should show:

```text
Ghega
Move beyond Mirth.
```

Navigation should not be locked by branding. Current suggested high-level grouping:

- Channels
- Messages
- Operations
- Migrations
- Settings

## AI naming

Use:

```text
Ghega Skills
Ghega MCP Server
Ghega Migration Assistant
```

Technical phrasing:

> Ghega ships a curated library of Agent Skills following the Agent Skills standard.

Do not invent a custom skills format.

## MCP naming

MCP server display name:

```text
Ghega MCP Server
```

Command:

```bash
ghega mcp serve
```

MCP tool names should be domain-oriented and not over-prefixed:

```text
channel.list
message.search_metadata
migration.get_report
replay.preview
audit.search
```

The MCP server namespace provides product context.

## README opening

Use this as the default README opening:

```markdown
# Ghega

Ghega is an open-source healthcare integration engine for teams that want to move beyond Mirth.

It provides typed channel definitions, deterministic tests, durable message processing, replay safety, observability, migration tooling, and AI-assisted authoring for HL7v2, FHIR, MLLP, HTTP, SFTP, files, and database integrations.
```

## Rename discipline

Although Ghega is now the selected working name, keep branding centralized.

Required:

```bash
make rename-product OLD=ghega NEW=<new-slug>
```

The command should update generated examples, docs, UI strings, Helm examples, Docker labels, env var examples, and MCP examples.

Do not rewrite Git history.

## Agent instruction

From now on, agents should use Ghega consistently.

If an agent creates or updates a file containing product branding, it must use the Ghega naming rules in this document or read `branding/product.yaml`.

