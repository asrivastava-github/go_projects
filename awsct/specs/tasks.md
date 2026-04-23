# awsct — Tasks

**Date:** 2026-04-17
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Project Scaffold

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `main.go`, `go.mod`, `Makefile`, `cmd/root.go`

- [x] Step 1: Create go.mod (`go mod init awsct`)
- [x] Step 2: Create main.go (entry point calling `cmd.Execute()`)
- [x] Step 3: Create cmd/root.go with global flags (`--last`, `--region`, `--profile`, `--json`, `--limit`) and aws-vault re-exec logic
- [x] Step 4: Create Makefile (build, install, tidy, test, clean, lint targets)
- [x] Step 5: Add cobra dependency and verify build
- [x] Step 6: Commit

## Task 2: Duration Parser

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** `pkg/cloudtrail/duration.go`

- [x] Step 1: Create duration.go — parse human-friendly duration strings (30m, 1h, 7d) with 90-day max clamp
- [x] Step 2: Verify compilation
- [x] Step 3: Commit

## Task 3: CloudTrail Client & Types

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** `pkg/cloudtrail/types.go`, `pkg/cloudtrail/client.go`

- [x] Step 1: Create types.go — Event struct, Resource struct, LookupParams struct
- [x] Step 2: Create client.go — AWS config loading, CloudTrail client creation (SDK v2)
- [x] Step 3: Add AWS SDK v2 CloudTrail dependency
- [x] Step 4: Commit

## Task 4: Core Lookup Logic

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `pkg/cloudtrail/lookup.go`

- [x] Step 1: Create lookup.go — LookupByAction, LookupByUser, LookupByResource with shared internal `lookup` function, auto-pagination, and throttle retry (exponential backoff)
- [x] Step 2: Verify compilation
- [x] Step 3: Commit

## Task 5: CLI Subcommands (who, user, resource)

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001
**Files:** `cmd/who.go`, `cmd/user.go`, `cmd/resource.go`

- [x] Step 1: Add shared output helpers to root.go (printTable, printJSON, printEvents)
- [x] Step 2: Create cmd/who.go — `who <action>` subcommand calling LookupByAction
- [x] Step 3: Create cmd/user.go — `user <username>` subcommand calling LookupByUser
- [x] Step 4: Create cmd/resource.go — `resource <resource-name>` subcommand calling LookupByResource
- [x] Step 5: Build and verify help output
- [x] Step 6: Commit

## Task 6: MCP Server

**Traces to:** UC-MCP-001
**Files:** `cmd/serve.go`

- [x] Step 1: Add mcp-go dependency
- [x] Step 2: Create cmd/serve.go — `serve-mcp` subcommand with three tools (lookup_by_action, lookup_by_user, lookup_by_resource) over stdio transport
- [x] Step 3: Build and verify
- [x] Step 4: Commit

## Task 7: README & MCP Config

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** `README.md`

- [x] Step 1: Create README.md with description, installation, CLI usage examples, MCP setup, and limitations
- [x] Step 2: Final build and smoke test
- [x] Step 3: Commit

## Task 8: Integration Test (Live AWS)

**Traces to:** UC-WHO-001, UC-USER-001, UC-RES-001, UC-MCP-001
**Files:** (none — manual testing)

- [x] Step 1: Test CLI with real CloudTrail (user, who, JSON output)
- [x] Step 2: Test MCP server (tools/list via stdin)
- [x] Step 3: Add MCP config to Claude Code
- [x] Step 4: Final commit

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-04-17 | Initial spec and plan created | AS |
| 2026-04-23 | Migrated to per-project specs/ structure; all tasks complete | AS |
