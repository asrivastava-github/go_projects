# gcpat — Tasks

**Date:** 2026-04-23
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Project Scaffold

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `main.go`, `go.mod`, `Makefile`, `cmd/root.go`

- [x] Step 1: Create go.mod (`go mod init gcpat`)
- [x] Step 2: Create main.go (entry point calling `cmd.Execute()`)
- [x] Step 3: Create cmd/root.go with global flags (`--project`, `--last`, `--json`, `--limit`), credential check (`ensureCredentials`), project validation (`requireProject`), and output helpers (`printEvents`, `printTable`, `printJSON`)
- [x] Step 4: Create Makefile (build, install, tidy, test, clean targets)
- [x] Step 5: Add cobra dependency and verify build
- [x] Step 6: Commit

## Task 2: Duration Parser

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** `pkg/auditlog/duration.go`

- [x] Step 1: Create duration.go — parse human-friendly duration strings (30m, 1h, 7d) with 90-day max clamp (`MaxLookback`)
- [x] Step 2: Handle empty string default (24h), day suffix conversion, and standard Go duration parsing
- [x] Step 3: Verify compilation
- [x] Step 4: Commit

## Task 3: Cloud Logging Client & Types

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** `pkg/auditlog/types.go`, `pkg/auditlog/client.go`

- [x] Step 1: Create types.go — Event struct (timestamp, method, service_name, principal, caller_ip, resource_name, resource_type, status, project_id, severity, insert_id) and LookupParams struct (Duration, ProjectID, Limit)
- [x] Step 2: Create client.go — GCP Cloud Logging client creation via `logging.NewClient(ctx)`, Close method
- [x] Step 3: Add GCP Cloud Logging and protobuf dependencies
- [x] Step 4: Commit

## Task 4: Core Lookup Logic

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `pkg/auditlog/lookup.go`

- [x] Step 1: Create lookup.go — `LookupByAction`, `LookupByUser`, `LookupByResource` with shared internal `lookup` function
- [x] Step 2: Implement Cloud Logging filter construction (logName for admin activity, time range, query-specific filter)
- [x] Step 3: Implement `ListLogEntries` iteration with limit enforcement
- [x] Step 4: Implement AuditLog proto payload unmarshalling to extract method, service, resource, principal, caller IP, and status
- [x] Step 5: Verify compilation
- [x] Step 6: Commit

## Task 5: CLI Subcommands (who, user, resource)

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `cmd/who.go`, `cmd/user.go`, `cmd/resource.go`

- [x] Step 1: Create cmd/who.go — `who <method>` subcommand (ExactArgs(1)) calling `LookupByAction`
- [x] Step 2: Create cmd/user.go — `user <principal>` subcommand (ExactArgs(1)) calling `LookupByUser`
- [x] Step 3: Create cmd/resource.go — `resource <resource-name>` subcommand (ExactArgs(1)) calling `LookupByResource`
- [x] Step 4: Each subcommand: ensureCredentials → requireProject → ParseDuration → NewClient → Lookup → printEvents
- [x] Step 5: Build and verify help output
- [x] Step 6: Commit

## Task 6: MCP Server

**Traces to:** UC-MCP-001
**Files:** `cmd/serve.go`

- [x] Step 1: Add mcp-go dependency
- [x] Step 2: Create cmd/serve.go — `serve-mcp` subcommand with three tools (`lookup_by_action`, `lookup_by_user`, `lookup_by_resource`) over stdio transport
- [x] Step 3: Implement `parseMCPParams` helper — extract `project` (required), `last` (default 24h), `limit` (default 50) with duration fallback
- [x] Step 4: Implement `eventsToResult` helper — JSON marshal events for MCP response, handle empty results
- [x] Step 5: Build and verify
- [x] Step 6: Commit

## Task 7: Integration Test (Live GCP)

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** (none — manual testing)

- [x] Step 1: Test CLI with real Cloud Audit Logs (who, user, resource, JSON output)
- [x] Step 2: Test MCP server (tools/list via stdin)
- [x] Step 3: Add MCP config to Claude Code / Amp
- [x] Step 4: Final commit

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-04-23 | Initial spec created; all tasks complete (retroactive documentation) | AS |
