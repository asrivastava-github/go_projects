# awsct — AWS CloudTrail Finder

**Date:** 2026-04-17
**Status:** Draft
**Repository:** go_projects/awsct

## Purpose

A Go CLI tool and MCP server for quickly querying AWS CloudTrail events. Answers three questions:
1. **Who did action X?** — e.g., who called `DeleteBucket` in the last 24h?
2. **What did user X do?** — e.g., what has `alice` done in the last 2h?
3. **What happened to resource X?** — e.g., all actions on `arn:aws:s3:::my-bucket`

## Architecture

Single Go binary with two modes:
- **CLI mode** (default): cobra-based subcommands with table/JSON output
- **MCP mode** (`serve-mcp`): stdio-based MCP server exposing three tools

Both modes share core query logic in `pkg/cloudtrail/`.

```
CLI Layer (cobra, flags, table)  ──┐
                                   ├──▶  pkg/cloudtrail (core logic)  ──▶  AWS CloudTrail API
MCP Layer (stdio, JSON-RPC)      ──┘
```

## Project Structure

```
awsct/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Cobra root, aws-vault re-exec, global flags
│   ├── who.go           # "who" subcommand — who did action X?
│   ├── user.go          # "user" subcommand — what did user X do?
│   ├── resource.go      # "resource" subcommand — what happened to resource X?
│   └── serve.go         # "serve-mcp" subcommand — start MCP server
├── pkg/
│   └── cloudtrail/
│       ├── client.go    # AWS session + CloudTrail client setup
│       └── lookup.go    # Core query functions (LookupEvents wrapper, pagination, throttle retry)
├── go.mod
├── go.sum
├── Makefile
└── README.md
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
awsct resource arn:aws:s3:::my-bucket --last 7d
awsct resource i-0abc123def456 --last 2h

# Start as MCP server (stdio mode)
awsct serve-mcp
```

### Global Flags

| Flag       | Default      | Description                        |
|------------|--------------|------------------------------------|
| `--last`   | `24h`        | Time window (e.g., 30m, 1h, 7d)   |
| `--region` | `us-east-1`  | AWS region                         |
| `--profile`| (none)       | AWS profile for aws-vault          |
| `--json`   | `false`      | Output as JSON instead of table    |
| `--limit`  | `50`         | Max results to return              |

## Authentication

Uses the **AWS SDK v2 default credential chain** (env vars, shared config, IAM role, etc.).

**CLI mode:**
1. Check if AWS credentials are valid via `sts:GetCallerIdentity`
2. If invalid/expired, re-exec under `aws-vault exec <profile> -- awsct <args>` (same pattern as `goto`)

**MCP mode:**
- Relies on ambient credentials (e.g., run under `aws-vault exec --server <profile> -- awsct serve-mcp`)
- No aws-vault re-exec in MCP mode (would break stdio session)

## Core Query Logic (`pkg/cloudtrail/`)

Three functions wrapping CloudTrail `LookupEvents` API:

| Function              | LookupAttribute | Filters by                |
|-----------------------|-----------------|---------------------------|
| `LookupByAction()`    | `EventName`     | Action like `DeleteBucket`|
| `LookupByUser()`      | `Username`      | IAM user/role name        |
| `LookupByResource()`  | `ResourceName`  | Exact resource name/ID    |

Shared behavior:
- Auto-pagination (API returns max 50 per page)
- Time window via `StartTime` / `EndTime` parameters
- Max lookback: 90 days (CloudTrail API limit)
- Results sorted by time (most recent first)

### Output Format

**Table mode** (default):
```
TIMESTAMP              USER          ACTION                RESOURCE                          SOURCE
2026-04-17 10:45:03    alice         DeleteBucket          arn:aws:s3:::my-bucket            s3.amazonaws.com
2026-04-17 10:32:11    alice         StopInstances         i-0abc123def456                   ec2.amazonaws.com
```

**JSON mode** (`--json`):
```json
[
  {
    "timestamp": "2026-04-17T10:45:03Z",
    "event_name": "DeleteBucket",
    "event_source": "s3.amazonaws.com",
    "username": "alice",
    "source_ip": "10.0.1.50",
    "read_only": false,
    "resources": [
      {"type": "AWS::S3::Bucket", "name": "my-bucket"}
    ],
    "event_id": "abc123-def456",
    "aws_region": "us-east-1"
  }
]
```

## MCP Server

Activated via `awsct serve-mcp`. Uses **stdio** transport (like existing `gitlab` and `jenkins` MCP servers).

### Tools Exposed

| Tool                 | Description                  | Parameters                              |
|----------------------|------------------------------|-----------------------------------------|
| `lookup_by_action`   | Who performed action X?      | `action` (required), `last` (optional)  |
| `lookup_by_user`     | What did user X do?          | `username` (required), `last` (optional)|
| `lookup_by_resource` | What happened to resource X? | `resource` (required), `last` (optional)|

### Claude Code Integration

Add to `~/.claude.json` mcpServers:

```json
"awsct": {
  "type": "stdio",
  "command": "/path/to/awsct",
  "args": ["serve-mcp"]
}
```

This enables natural language queries in Amp like:
- "Who deleted the S3 bucket prod-data in the last 2 hours?"
- "What has alice done today?"
- "Show me all actions on the production database instance"

## Dependencies

- `github.com/aws/aws-sdk-go-v2` — AWS SDK v2 (v1 reached end-of-support July 2025)
- `github.com/spf13/cobra` — CLI framework (same as `ip-finder`)
- `github.com/olekukonenko/tablewriter` — table output
- `github.com/mark3labs/mcp-go` — MCP server SDK for Go (stdio transport)

## Constraints & Limitations

- CloudTrail `LookupEvents` only returns **management events** (not data events)
- Max lookback is **90 days**; tool will reject/clamp ranges beyond this
- API returns max **50 events per page** (auto-paginated, stops at `--limit`)
- Only **one `LookupAttribute` per query** (API limitation) — compound filters (e.g., action + resource) are not supported server-side
- Queries are **per-account, per-region** — results only cover the configured AWS account and region
- Some global-service events (IAM, STS, CloudFront, Route53) may only appear in `us-east-1`
- `LookupEvents` is throttled to **2 requests/second per account/region** — tool uses exponential backoff on throttling
- `ResourceName` lookup uses **exact name as CloudTrail records it** (may be bucket name, instance ID, etc. — not always full ARN)
- `Username` in CloudTrail may be IAM user, role session name, service principal, or `root` — for SSO/assumed-role, richer identity is parsed from event details
