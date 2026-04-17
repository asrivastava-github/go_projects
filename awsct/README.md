# awsct — AWS CloudTrail Finder

A CLI tool and MCP server for quickly querying AWS CloudTrail events.

## Install

```bash
cd awsct
make build      # builds to bin/awsct
# or
make install    # installs to $GOPATH/bin
```

## CLI Usage

```bash
# Who performed a specific action?
awsct who DeleteBucket --last 24h
awsct who RunInstances --last 7d --json

# What did a specific user do?
awsct user alice --last 1h
awsct user alice --last 30m --json

# What happened to a specific resource?
awsct resource my-bucket --last 7d
awsct resource i-0abc123def456 --last 2h

# With a specific AWS profile (triggers aws-vault if needed)
awsct user alice --last 1h --profile prod

# Different region
awsct who DeleteBucket --last 24h --region us-west-2
```

### Flags

| Flag        | Default     | Description                       |
|-------------|-------------|-----------------------------------|
| `--last`    | `24h`       | Time window (30m, 1h, 7d, etc.)  |
| `--region`  | `us-east-1` | AWS region                        |
| `--profile` | (none)      | AWS profile for aws-vault         |
| `--json`    | `false`     | Output as JSON                    |
| `--limit`   | `50`        | Max results to return             |

## MCP Server

Start as an MCP server for AI assistant integration:

```bash
awsct serve-mcp
```

### Claude Code / Amp Setup

Add to `~/.claude.json` under `mcpServers`:

```json
"awsct": {
  "type": "stdio",
  "command": "/path/to/awsct",
  "args": ["serve-mcp"]
}
```

For profiles requiring aws-vault:

```json
"awsct": {
  "type": "stdio",
  "command": "aws-vault",
  "args": ["exec", "your-profile", "--", "/path/to/awsct", "serve-mcp"]
}
```

### MCP Tools

| Tool                 | Description                  | Required Params |
|----------------------|------------------------------|-----------------|
| `lookup_by_action`   | Who performed action X?      | `action`        |
| `lookup_by_user`     | What did user X do?          | `username`      |
| `lookup_by_resource` | What happened to resource X? | `resource`      |

All tools accept optional `last` (default: 24h), `region` (default: us-east-1), and `limit` (default: 50) params.

## Limitations

- CloudTrail `LookupEvents` returns **management events only** (not data events)
- Max lookback: **90 days**
- Queries are **per-account, per-region**
- API throttled to **2 requests/second** per account/region
- Only **one lookup attribute** per query (no compound filters)
- Resource lookup uses **exact name** as CloudTrail records it
