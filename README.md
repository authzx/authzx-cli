# AuthzX CLI

Command-line tool for [AuthzX](https://authzx.com) — configure an API key and run authorization checks from the shell.

## Install

Requires Go 1.21+.

```bash
go install github.com/authzx/authzx-cli/cmd/authzx@latest
```

Pre-built binaries and `brew install` support are planned for a future release.

## Quickstart

```bash
# Store your API key (paste when prompted — input is masked)
authzx configure

# Run an authorization check
authzx check --subject user-123 --resource doc-456 --action read
# allowed: true  (direct access)
```

## Commands

### `authzx configure`

Interactively store an AuthzX API key at `~/.authzx/config.yaml`. Input is
masked; the file is written with mode `0600` inside a `0700` directory.

```
$ authzx configure
AuthzX API Key: ****
Saved to ~/.authzx/config.yaml
```

Get a key from the AuthzX console under **Settings → API Keys**. API keys
start with `azx_`. Do not paste an OAuth client secret (which starts with
`azx_cs_`) — the CLI will refuse it.

### `authzx check`

Run a single authorization check.

```bash
authzx check --subject user-123 --resource doc-456 --action read

# With subject/resource types, roles, and context
authzx check \
  --subject user:123 \
  --action read \
  --resource document:456 \
  --roles editor,viewer \
  --context '{"ip":"10.0.0.1"}'

# Against a locally-running AuthzX agent
authzx check --subject user:123 --action read --resource document:456 --local
```

### `authzx version`

Print the CLI version.

## Authentication

When a command needs an API key, the CLI resolves it in this order (first
match wins):

1. `--api-key` flag
2. `AUTHZX_API_KEY` environment variable
3. `api_key` from `~/.authzx/config.yaml`

If none is set, commands print: `Not authenticated. Run 'authzx configure'
or set AUTHZX_API_KEY.`

## Config file

`~/.authzx/config.yaml`:

```yaml
api_key: azx_...
```

File mode: `0600`. Directory mode: `0700`.
