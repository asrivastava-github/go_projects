# gcpat — Requirements

**Date:** 2026-04-23
**Status:** Active
**Project:** go_projects/gcpat

## Overview

**Purpose:** A Go CLI tool and MCP server for quickly querying GCP Cloud Audit Logs.

**Problem Statement:** Investigating GCP activity requires navigating the Cloud Logging console or writing ad-hoc SDK scripts. gcpat provides instant answers to the three most common questions: who performed an action, what did a principal do, and what happened to a resource — from the terminal or via an AI agent.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | CLI user who runs gcpat subcommands to investigate Cloud Audit Log events |
| AI Agent | MCP consumer (e.g., Claude Code / Amp) that calls gcpat tools over stdio |

## Use Cases

### UC-WHO-001: Find who performed a specific GCP method

- **Actor:** Developer
- **Preconditions:** Valid GCP application-default credentials available; target project specified
- **Main Flow:**
  1. Developer runs `gcpat who <method> --project <project-id> --last <duration>`
  2. System verifies GCP credentials via `gcloud auth application-default print-access-token`
  3. System validates `--project` is provided (required flag)
  4. System parses the duration flag (default 24h) via `ParseDuration`
  5. System creates a Cloud Logging client for the specified project
  6. System queries Cloud Audit Logs with filter `protoPayload.methodName="<method>"`
  7. System returns results up to `--limit`, sorted newest first
  8. System displays results as a table (or JSON with `--json`)
- **Alternative Flows:**
  - 4a. Invalid duration format → log.Fatalf with "Invalid duration" error
  - 2a. No credentials configured → log.Fatalf with "Run: gcloud auth application-default login"
  - 3a. `--project` not provided → log.Fatalf with "--project is required"
  - 7a. No events found → print "No events found."
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN valid GCP credentials and Cloud Audit Log events exist for `compute.instances.delete` in the last 24h
  - WHEN the developer runs `gcpat who compute.instances.delete --project my-project --last 24h`
  - THEN the tool displays matching events with timestamp, principal, method, resource, and service columns sorted newest first

### UC-USER-001: Find what a specific principal did

- **Actor:** Developer
- **Preconditions:** Valid GCP application-default credentials available; target project specified
- **Main Flow:**
  1. Developer runs `gcpat user <principal> --project <project-id> --last <duration>`
  2. System verifies GCP credentials via `gcloud auth application-default print-access-token`
  3. System validates `--project` is provided (required flag)
  4. System parses the duration flag (default 24h) via `ParseDuration`
  5. System creates a Cloud Logging client for the specified project
  6. System queries Cloud Audit Logs with filter `protoPayload.authenticationInfo.principalEmail="<principal>"`
  7. System returns results up to `--limit`, sorted newest first
  8. System displays results as a table (or JSON with `--json`)
- **Alternative Flows:**
  - 4a. Invalid duration format → log.Fatalf with "Invalid duration" error
  - 2a. No credentials configured → log.Fatalf with "Run: gcloud auth application-default login"
  - 3a. `--project` not provided → log.Fatalf with "--project is required"
  - 7a. No events found → print "No events found."
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN valid GCP credentials and Cloud Audit Log events exist for principal `alice@example.com` in the last 1h
  - WHEN the developer runs `gcpat user alice@example.com --project my-project --last 1h`
  - THEN the tool displays all actions performed by that principal with timestamp, method, resource, and service columns

### UC-RES-001: Find what happened to a specific resource

- **Actor:** Developer
- **Preconditions:** Valid GCP application-default credentials available; target project specified
- **Main Flow:**
  1. Developer runs `gcpat resource <resource-name> --project <project-id> --last <duration>`
  2. System verifies GCP credentials via `gcloud auth application-default print-access-token`
  3. System validates `--project` is provided (required flag)
  4. System parses the duration flag (default 24h) via `ParseDuration`
  5. System creates a Cloud Logging client for the specified project
  6. System queries Cloud Audit Logs with filter `protoPayload.resourceName:"<resource-name>"` (substring match via `:`)
  7. System returns results up to `--limit`, sorted newest first
  8. System displays results as a table (or JSON with `--json`)
- **Alternative Flows:**
  - 4a. Invalid duration format → log.Fatalf with "Invalid duration" error
  - 2a. No credentials configured → log.Fatalf with "Run: gcloud auth application-default login"
  - 3a. `--project` not provided → log.Fatalf with "--project is required"
  - 7a. No events found → print "No events found."
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN valid GCP credentials and Cloud Audit Log events exist for resource `my-bucket` in the last 7d
  - WHEN the developer runs `gcpat resource my-bucket --project my-project --last 7d`
  - THEN the tool displays all actions on that resource with timestamp, principal, method, and service columns

### UC-MCP-001: Serve as MCP server for AI agents

- **Actor:** AI Agent
- **Preconditions:** Valid GCP application-default credentials available
- **Main Flow:**
  1. Process starts with `gcpat serve-mcp`
  2. System starts a stdio-based MCP server exposing three tools: `lookup_by_action`, `lookup_by_user`, `lookup_by_resource`
  3. AI Agent sends a JSON-RPC tool call (e.g., `lookup_by_action` with `action` and `project` params)
  4. System extracts required params (`project` always required, plus the query-specific param)
  5. System parses `last` (default 24h) and `limit` (default 50)
  6. System executes the corresponding Cloud Audit Log lookup
  7. System returns results as JSON via `mcp.NewToolResultText`
- **Alternative Flows:**
  - 3a. Missing required parameter → return MCP error response via `mcp.NewToolResultError`
  - 6a. Cloud Logging API error → return error text in MCP response
  - 7a. No events found → return "No events found." text
  - 5a. Invalid duration → silently default to 24h (MCP fallback behavior)
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN gcpat is running as an MCP server with valid ambient credentials
  - WHEN an AI Agent calls the `lookup_by_user` tool with `username: "alice@example.com"`, `project: "my-project"`, and `last: "1h"`
  - THEN the server returns a JSON array of matching Cloud Audit Log events

## Business Rules

| ID | Rule |
|----|------|
| BR-001 | Max lookback is 90 days; durations exceeding this are clamped to 90 days |
| BR-002 | Default time window is 24h when `--last` is not specified |
| BR-003 | Default result limit is 50; configurable via `--limit` |
| BR-004 | Results are sorted newest first (`timestamp desc`) |
| BR-005 | Queries are per-project — results only cover the specified GCP project's admin activity audit logs |

## Non-Functional Requirements

- **Security:** Authentication via `gcloud auth application-default` credentials. CLI mode validates credentials by running `gcloud auth application-default print-access-token` before each query. MCP mode relies on ambient credentials (no credential re-exec to avoid breaking the stdio session).
- **Performance:** Cloud Logging API has per-project quota limits. Results are paginated via the API's iterator; tool stops at `--limit`.
- **Portability:** Single Go binary, cross-compilable. Primary target: macOS (arm64).

## Constraints & Limitations

- Cloud Audit Logs `activity` log only contains **admin activity** events (not data access events)
- Max lookback is **90 days**; tool clamps ranges beyond this
- Queries are **per-project** — `--project` is always required
- `resourceName` lookup uses **substring match** (`:` operator) — may match multiple resources with similar names
- `methodName` and `principalEmail` lookups use **exact match** (`=` operator)
- Credential check relies on **gcloud CLI** being installed and available in PATH
- MCP mode silently defaults to 24h if duration parsing fails (no user-facing error)
