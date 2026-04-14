# AuthzX CLI

Command-line tool for [AuthzX](https://authzx.com) — evaluate authorization checks, manage policies, and run the local agent.

## Install

```bash
# Homebrew
brew install authzx/tap/authzx

# Go
go install github.com/authzx/authzx-cli/cmd/authzx@latest
```

## Quickstart

```bash
# Store your API key
authzx login

# Test a policy check
authzx check --subject user:123 --action read --resource document:456

# ALLOWED
#   Reason: role_match
#   Policy: pol_abc
#   Path:   role
```

## Commands

### `authzx login`

Store API key and cloud URL in `~/.authzx/credentials`.

```bash
authzx login
# API Key: azx_...
# Cloud URL [https://api.authzx.com]:
```

### `authzx logout`

Remove stored credentials.

### `authzx check`

Authorize an authorization check against cloud or local agent.

```bash
authzx check \
  --subject user:123 \
  --action read \
  --resource document:456

# With roles and context
authzx check \
  --subject user:123 \
  --action read \
  --resource document:456 \
  --roles editor,viewer \
  --context '{"ip":"10.0.0.1"}'

# Against local agent instead of cloud
authzx check \
  --subject user:123 \
  --action read \
  --resource document:456 \
  --local
```

### `authzx policies list`

List all policies for your tenant.

### `authzx policies get <id>`

Get details of a specific policy.

### `authzx agent start`

Start the AuthzX agent locally. Requires the `authzx-agent` binary in PATH.

```bash
authzx agent start
authzx agent start --config ./my-config.yaml --listen 0.0.0.0:8181
```

### `authzx agent status`

Check if the local agent is running and healthy.

```bash
authzx agent status
# Agent is running at 127.0.0.1:8181
```

### `authzx version`

Print the CLI version.

## Credentials

Stored at `~/.authzx/credentials`:

```json
{
  "api_key": "azx_...",
  "cloud_url": "https://api.authzx.com"
}
```
