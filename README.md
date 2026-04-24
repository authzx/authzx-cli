# AuthzX CLI

Command-line tool for [AuthzX](https://authzx.com) — configure an API key and run authorization checks from the shell.

The CLI binary is `azx`.

## Install

### Homebrew (macOS / Linux)

```bash
brew tap authzx/tap
brew install azx
```

### Binary download

Grab the latest release from [GitHub Releases](https://github.com/authzx/authzx-cli/releases) — binaries available for macOS, Linux, and Windows (amd64/arm64).

### From source

Requires Go 1.21+.

```bash
go install github.com/authzx/authzx-cli/cmd/azx@latest
```

## Quickstart

```bash
# Store your API key (paste when prompted — input is masked)
azx configure

# Run an authorization check
azx check --subject user-123 --resource doc-456 --action read
# allowed: true  (direct access)
```

## Commands

### `azx configure`

Interactively store an AuthzX API key at `~/.authzx/config.yaml`. Input is
masked; the file is written with mode `0600` inside a `0700` directory.

```
$ azx configure
AuthzX API Key: ****
Saved to ~/.authzx/config.yaml
```

Get a key from the AuthzX console under **Settings → API Keys**. API keys
start with `azx_`. Do not paste an OAuth client secret (which starts with
`azx_cs_`) — the CLI will refuse it.

### `azx check`

Run a single authorization check.

```bash
azx check --subject user-123 --resource doc-456 --action read

# With subject/resource types, roles, and context
azx check \
  --subject user:123 \
  --action read \
  --resource document:456 \
  --roles editor,viewer \
  --context '{"ip":"10.0.0.1"}'

# Against a locally-running AuthzX agent
azx check --subject user:123 --action read --resource document:456 --local
```

### `azx agent`

Manage a local AuthzX agent running in Docker.

```bash
# Generate agent config with your API key prefilled
azx agent config

# Start the agent (detached by default)
azx agent start

# Start a specific version on a custom port
azx agent start --image authzx/agent:v0.1.3 --port 9090

# Start with extra env vars and in foreground mode
azx agent start --foreground --env AUTHZX_LOG_LEVEL=debug

# Check agent health
azx agent status

# Tail agent logs
azx agent logs
azx agent logs --tail 50

# Stop and remove the container
azx agent stop
```

| Flag | Default | Description |
|------|---------|-------------|
| `--image` | `authzx/agent:latest` | Docker image to run |
| `--port` | `8181` | Host port mapped to the agent |
| `--name` | `authzx-agent` | Container name |
| `--config` | `authzx-agent.yaml` | Config file to mount |
| `--foreground` | `false` | Run attached (not detached) |
| `-e, --env` | — | Environment variables (repeatable) |
| `--tail` | — | Number of log lines to show (logs subcommand) |

### `azx version`

Print the CLI version.

## Authentication

When a command needs an API key, the CLI resolves it in this order (first
match wins):

1. `--api-key` flag
2. `AUTHZX_API_KEY` environment variable
3. `api_key` from `~/.authzx/config.yaml`

If none is set, commands print: `Not authenticated. Run 'azx configure'
or set AUTHZX_API_KEY.`

## Config file

`~/.authzx/config.yaml`:

```yaml
api_key: azx_...
```

File mode: `0600`. Directory mode: `0700`.
