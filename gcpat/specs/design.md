# gcpat — Design

**Date:** 2026-04-23
**Status:** Active

## Architecture Overview

Single Go binary with two modes:
- **CLI mode** (default): cobra-based subcommands with table/JSON output
- **MCP mode** (`serve-mcp`): stdio-based MCP server exposing three tools

Both modes share core query logic in `pkg/auditlog/`.

```
CLI Layer (cobra, flags, tabwriter)  ──┐
                                       ├──▶  pkg/auditlog (core logic)  ──▶  GCP Cloud Logging API
MCP Layer (stdio, JSON-RPC)           ──┘
```

## Project Structure

```
gcpat/
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/
│   ├── root.go          # Cobra root command, global flags, credential check, output helpers
│   ├── who.go           # "who" subcommand — lookup by methodName
│   ├── user.go          # "user" subcommand — lookup by principalEmail
│   ├── resource.go      # "resource" subcommand — lookup by resourceName
│   └── serve.go         # "serve-mcp" subcommand — start MCP stdio server
├── pkg/
│   └── auditlog/
│       ├── client.go    # GCP Cloud Logging client creation
│       ├── lookup.go    # Core query functions (ListLogEntries wrapper, proto parsing)
│       ├── types.go     # Event struct, LookupParams
│       └── duration.go  # ParseDuration helper (30m, 1h, 7d) with 90-day max clamp
├── go.mod
├── go.sum
└── Makefile
```

## Components

### Component: CLI Layer

- **Responsibility:** Parse user input via cobra subcommands (`who`, `user`, `resource`), manage global flags (`--project`, `--last`, `--json`, `--limit`), verify GCP credentials, and format output as table or JSON.
- **Interfaces:** Calls `pkg/auditlog.Client` methods; outputs to stdout via `printEvents`.
- **Key Decisions:** Uses cobra for consistency across the monorepo. Credential check shells out to `gcloud auth application-default print-access-token`. Output helpers (`printTable`, `printJSON`) live in root.go alongside flags.

### Component: MCP Layer

- **Responsibility:** Expose Cloud Audit Log lookup functions as MCP tools over stdio transport for AI agent consumption.
- **Interfaces:** JSON-RPC over stdin/stdout; calls `pkg/auditlog.Client` methods. Returns JSON-formatted events via `mcp.NewToolResultText`.
- **Key Decisions:** Uses stdio transport (same pattern as awsct MCP server). Each tool requires `project` as a required parameter (since there's no ambient project context). Invalid durations silently default to 24h to avoid breaking agent workflows.

### Component: Core Query Logic (`pkg/auditlog/`)

- **Responsibility:** Wrap GCP Cloud Logging `ListLogEntries` API with typed Go functions, proto payload parsing, and duration handling.
- **Interfaces:** Three public lookup functions (`LookupByAction`, `LookupByUser`, `LookupByResource`), a client constructor, and a duration parser.
- **Key Decisions:** Single internal `lookup` function shared by all three public methods — they differ only in the filter string. Filters target the `cloudaudit.googleapis.com%2Factivity` log. Proto payloads are unmarshalled from `AuditLog` protobuf to extract method, principal, resource, service, caller IP, and status.

#### Core Query Functions

| Function              | Filter Field                                     | Match Type |
|-----------------------|--------------------------------------------------|------------|
| `LookupByAction()`    | `protoPayload.methodName`                        | Exact (`=`) |
| `LookupByUser()`      | `protoPayload.authenticationInfo.principalEmail` | Exact (`=`) |
| `LookupByResource()`  | `protoPayload.resourceName`                      | Substring (`:`) |

## Data Flow

1. User (CLI) or AI Agent (MCP) provides a query type, value, project ID, and optional time window
2. CLI layer parses flags / MCP layer extracts JSON-RPC params
3. `ParseDuration` converts the time window string (e.g., `7d`) to `time.Duration`, clamped to 90 days max
4. `pkg/auditlog.Client` builds a Cloud Logging filter combining the log name, query filter, and time range
5. Client calls `ListLogEntries` with `ResourceNames: ["projects/<project>"]`, `OrderBy: "timestamp desc"`, and `PageSize` set to limit
6. Iterator reads entries up to limit; each entry's `protoPayload` is unmarshalled as `AuditLog` protobuf
7. Extracted fields populate `[]Event` structs (timestamp, method, principal, resource, service, caller IP, status, severity, etc.)
8. CLI renders as table (tabwriter) or JSON; MCP returns indented JSON via `mcp.NewToolResultText`

### Output Formats

**Table mode** (default):
```
TIMESTAMP              PRINCIPAL              METHOD                        RESOURCE                    SERVICE
---------              ---------              ------                        --------                    -------
2026-04-23 10:45:03    alice@example.com      compute.instances.delete      my-instance                 compute.googleapis.com
2026-04-23 10:32:11    alice@example.com      storage.buckets.delete        my-bucket                   storage.googleapis.com
```

**JSON mode** (`--json`):
```json
[
  {
    "timestamp": "2026-04-23T10:45:03Z",
    "method": "compute.instances.delete",
    "service_name": "compute.googleapis.com",
    "principal": "alice@example.com",
    "caller_ip": "10.0.1.50",
    "resource_name": "my-instance",
    "resource_type": "gce_instance",
    "status": "OK",
    "project_id": "my-project",
    "severity": "NOTICE",
    "insert_id": "abc123def456"
  }
]
```

## MCP Server

Activated via `gcpat serve-mcp`. Uses **stdio** transport.

### Tools Exposed

| Tool                 | Description                          | Parameters                                              |
|----------------------|--------------------------------------|---------------------------------------------------------|
| `lookup_by_action`   | Who performed GCP method X?          | `action` (required), `project` (required), `last`, `limit` |
| `lookup_by_user`     | What did principal X do?             | `username` (required), `project` (required), `last`, `limit` |
| `lookup_by_resource` | What happened to resource X?         | `resource` (required), `project` (required), `last`, `limit` |

### Claude Code / Amp Integration

Add to MCP server config:

```json
"gcpat": {
  "type": "stdio",
  "command": "/path/to/gcpat",
  "args": ["serve-mcp"]
}
```

This enables natural language queries like:
- "Who deleted the compute instance in project my-project in the last 2 hours?"
- "What has alice@example.com done in project my-project today?"
- "Show me all actions on bucket my-bucket in the last week"

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| `cloud.google.com/go/logging` | GCP Cloud Logging client (ListLogEntries API) | v1.16.0 |
| `github.com/spf13/cobra` | CLI framework | v1.10.2 |
| `github.com/mark3labs/mcp-go` | MCP server SDK for Go (stdio transport) | v0.48.0 |
| `google.golang.org/api` | GCP API support | v0.276.0 |
| `google.golang.org/genproto` | AuditLog protobuf definitions | latest |
| `google.golang.org/protobuf` | Protobuf unmarshalling | v1.36.11 |

## Authentication & Security

Uses **GCP application-default credentials** (`gcloud auth application-default login`).

**CLI mode:**
1. Before each query, run `gcloud auth application-default print-access-token` to verify credentials
2. If credentials are not configured → fatal error with instructions to run `gcloud auth application-default login`

**MCP mode:**
- Relies on ambient credentials (application-default credentials must be configured before starting the server)
- No credential validation step (errors surface as API call failures)

## Error Handling Strategy

- Missing `--project` flag → log.Fatalf with "--project is required" message
- Invalid duration format → log.Fatalf with "Invalid duration" error (CLI); silently default to 24h (MCP)
- GCP credential errors → log.Fatalf with instructions to configure credentials (CLI); return error text in MCP response
- Cloud Logging API errors → return error wrapped with context (CLI); return error text via `mcp.NewToolResultError` (MCP)
- No events found → "No events found." message (CLI text or MCP text)
- Proto unmarshal failure → silently skip field extraction (event still added with available fields)

## CLI Global Flags

| Flag        | Default | Description                        |
|-------------|---------|------------------------------------|
| `--project` | (none)  | GCP project ID (required)          |
| `--last`    | `24h`   | Time window (e.g., 30m, 1h, 7d)   |
| `--json`    | `false` | Output as JSON instead of table    |
| `--limit`   | `50`    | Max results to return              |
