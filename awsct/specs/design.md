# awsct — Design

**Date:** 2026-04-17
**Status:** Active

## Architecture Overview

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
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/
│   ├── root.go          # Cobra root command, global flags, aws-vault re-exec
│   ├── who.go           # "who" subcommand — lookup by EventName
│   ├── user.go          # "user" subcommand — lookup by Username
│   ├── resource.go      # "resource" subcommand — lookup by ResourceName
│   └── serve.go         # "serve-mcp" subcommand — start MCP stdio server
├── pkg/
│   └── cloudtrail/
│       ├── client.go    # AWS config loading, CloudTrail client creation
│       ├── lookup.go    # Core query functions (LookupEvents wrapper, pagination, throttle retry)
│       ├── types.go     # Event struct, LookupParams
│       └── duration.go  # ParseDuration helper (30m, 1h, 7d)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Components

### Component: CLI Layer

- **Responsibility:** Parse user input via cobra subcommands (`who`, `user`, `resource`), manage global flags (`--last`, `--region`, `--profile`, `--json`, `--limit`), handle aws-vault re-exec for credential refresh, and format output as table or JSON.
- **Interfaces:** Calls `pkg/cloudtrail.Client` methods; outputs to stdout.
- **Key Decisions:** Uses cobra (same as `ip-finder`) for consistency across the monorepo. aws-vault re-exec pattern reused from `ip-finder/cmd/root.go`.

### Component: MCP Layer

- **Responsibility:** Expose CloudTrail lookup functions as MCP tools over stdio transport for AI agent consumption.
- **Interfaces:** JSON-RPC over stdin/stdout; calls `pkg/cloudtrail.Client` methods.
- **Key Decisions:** Uses stdio transport (same as existing `gitlab` and `jenkins` MCP servers). No aws-vault re-exec in MCP mode — would break the stdio session; relies on ambient credentials.

### Component: Core Query Logic (`pkg/cloudtrail/`)

- **Responsibility:** Wrap CloudTrail `LookupEvents` API with typed Go functions, auto-pagination, throttle retry, and duration parsing.
- **Interfaces:** Three public lookup functions, a client constructor, and a duration parser.
- **Key Decisions:** Single internal `lookup` function shared by all three public methods — they differ only in `LookupAttribute` type. Exponential backoff for throttling (base 500ms, max 8s, max 5 retries).

#### Core Query Functions

| Function              | LookupAttribute | Filters by                |
|-----------------------|-----------------|---------------------------|
| `LookupByAction()`    | `EventName`     | Action like `DeleteBucket`|
| `LookupByUser()`      | `Username`      | IAM user/role name        |
| `LookupByResource()`  | `ResourceName`  | Exact resource name/ID    |

## Data Flow

1. User (CLI) or AI Agent (MCP) provides a query type, value, and optional time window
2. CLI layer parses flags / MCP layer extracts JSON-RPC params
3. `ParseDuration` converts the time window string (e.g., `7d`) to `time.Duration`, clamped to 90 days max
4. `pkg/cloudtrail.Client` builds `LookupEventsInput` with one `LookupAttribute`, `StartTime`, and `EndTime`
5. Client paginates via `NextToken` until limit reached or no more pages, with throttle backoff
6. Events are parsed from CloudTrail response into `[]Event` structs (extracting `sourceIPAddress`, `readOnly` from the `CloudTrailEvent` JSON string)
7. CLI renders as table (tabwriter) or JSON; MCP returns JSON via `mcp.NewToolResultText`

### Output Formats

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

Activated via `awsct serve-mcp`. Uses **stdio** transport.

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

This enables natural language queries like:
- "Who deleted the S3 bucket prod-data in the last 2 hours?"
- "What has alice done today?"
- "Show me all actions on the production database instance"

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| `github.com/aws/aws-sdk-go-v2` | AWS SDK v2 (cloudtrail, sts, config) | latest |
| `github.com/spf13/cobra` | CLI framework (same as `ip-finder`) | latest |
| `github.com/olekukonenko/tablewriter` | Table output formatting | latest |
| `github.com/mark3labs/mcp-go` | MCP server SDK for Go (stdio transport) | latest |

## Authentication & Security

Uses the **AWS SDK v2 default credential chain** (env vars, shared config, IAM role, etc.).

**CLI mode:**
1. Check if AWS credentials are valid via `sts:GetCallerIdentity`
2. If invalid/expired, re-exec under `aws-vault exec <profile> -- awsct <args>` (same pattern as `ip-finder`)

**MCP mode:**
- Relies on ambient credentials (e.g., run under `aws-vault exec --server <profile> -- awsct serve-mcp`)
- No aws-vault re-exec in MCP mode (would break stdio session)

## Error Handling Strategy

- Invalid duration format → user-facing error with examples of valid formats
- AWS credential errors → CLI re-execs under aws-vault if `--profile` is set; MCP returns error text
- CloudTrail API throttling → exponential backoff (base 500ms, max 8s, max 5 retries)
- No events found → "No events found" message (CLI) or empty JSON array (MCP)

## CLI Global Flags

| Flag       | Default      | Description                        |
|------------|--------------|------------------------------------|
| `--last`   | `24h`        | Time window (e.g., 30m, 1h, 7d)   |
| `--region` | `us-east-1`  | AWS region                         |
| `--profile`| (none)       | AWS profile for aws-vault          |
| `--json`   | `false`      | Output as JSON instead of table    |
| `--limit`  | `50`         | Max results to return              |
