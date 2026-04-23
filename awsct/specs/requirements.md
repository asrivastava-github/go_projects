# awsct — Requirements

**Date:** 2026-04-17
**Status:** Active
**Project:** go_projects/awsct

## Overview

**Purpose:** A Go CLI tool and MCP server for quickly querying AWS CloudTrail events.

**Problem Statement:** Investigating AWS activity requires navigating the CloudTrail console or writing ad-hoc SDK scripts. awsct provides instant answers to the three most common questions: who did an action, what did a user do, and what happened to a resource — from the terminal or via an AI agent.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | CLI user who runs awsct subcommands to investigate CloudTrail events |
| AI Agent | MCP consumer (e.g., Claude Code / Amp) that calls awsct tools over stdio |

## Use Cases

### UC-WHO-001: Find who performed a specific action

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials available; target region and account configured
- **Main Flow:**
  1. Developer runs `awsct who <action> --last <duration>`
  2. System parses the duration flag (default 24h)
  3. System queries CloudTrail LookupEvents with LookupAttribute `EventName` = `<action>`
  4. System auto-paginates results up to `--limit`
  5. System displays results as a table (or JSON with `--json`) sorted newest first
- **Alternative Flows:**
  - 2a. Invalid duration format → print error with valid examples, exit 1
  - 3a. No credentials / expired credentials → re-exec under `aws-vault exec <profile>` if `--profile` set, otherwise print auth error
  - 4a. No events found → print "No events found" message
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005, BR-006
- **Acceptance Criteria:**
  - GIVEN valid AWS credentials and CloudTrail events exist for `DeleteBucket` in the last 24h
  - WHEN the developer runs `awsct who DeleteBucket --last 24h`
  - THEN the tool displays matching events with timestamp, user, action, resource, and source columns sorted newest first

### UC-USER-001: Find what a specific user did

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials available; target region and account configured
- **Main Flow:**
  1. Developer runs `awsct user <username> --last <duration>`
  2. System parses the duration flag (default 24h)
  3. System queries CloudTrail LookupEvents with LookupAttribute `Username` = `<username>`
  4. System auto-paginates results up to `--limit`
  5. System displays results as a table (or JSON with `--json`) sorted newest first
- **Alternative Flows:**
  - 2a. Invalid duration format → print error with valid examples, exit 1
  - 3a. No credentials / expired credentials → re-exec under `aws-vault exec <profile>` if `--profile` set, otherwise print auth error
  - 4a. No events found → print "No events found" message
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005, BR-006
- **Acceptance Criteria:**
  - GIVEN valid AWS credentials and CloudTrail events exist for user `alice` in the last 1h
  - WHEN the developer runs `awsct user alice --last 1h`
  - THEN the tool displays all actions performed by `alice` with timestamp, action, resource, and source columns

### UC-RES-001: Find what happened to a specific resource

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials available; target region and account configured
- **Main Flow:**
  1. Developer runs `awsct resource <resource-name> --last <duration>`
  2. System parses the duration flag (default 24h)
  3. System queries CloudTrail LookupEvents with LookupAttribute `ResourceName` = `<resource-name>`
  4. System auto-paginates results up to `--limit`
  5. System displays results as a table (or JSON with `--json`) sorted newest first
- **Alternative Flows:**
  - 2a. Invalid duration format → print error with valid examples, exit 1
  - 3a. No credentials / expired credentials → re-exec under `aws-vault exec <profile>` if `--profile` set, otherwise print auth error
  - 4a. No events found → print "No events found" message
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005, BR-006
- **Acceptance Criteria:**
  - GIVEN valid AWS credentials and CloudTrail events exist for `arn:aws:s3:::my-bucket` in the last 7d
  - WHEN the developer runs `awsct resource arn:aws:s3:::my-bucket --last 7d`
  - THEN the tool displays all actions on that resource with timestamp, user, action, and source columns

### UC-MCP-001: Serve as MCP server for AI agents

- **Actor:** AI Agent
- **Preconditions:** Ambient AWS credentials available (e.g., via `aws-vault exec --server`)
- **Main Flow:**
  1. Process starts with `awsct serve-mcp`
  2. System starts a stdio-based MCP server exposing three tools: `lookup_by_action`, `lookup_by_user`, `lookup_by_resource`
  3. AI Agent sends a JSON-RPC tool call (e.g., `lookup_by_action` with `action` param)
  4. System executes the corresponding CloudTrail lookup
  5. System returns results as JSON via `mcp.NewToolResultText`
- **Alternative Flows:**
  - 3a. Missing required parameter → return MCP error response
  - 4a. AWS API error → return error text in MCP response
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN awsct is running as an MCP server with valid ambient credentials
  - WHEN an AI Agent calls the `lookup_by_user` tool with `username: "alice"` and `last: "1h"`
  - THEN the server returns a JSON array of matching CloudTrail events

## Business Rules

| ID | Rule |
|----|------|
| BR-001 | Max lookback is 90 days (CloudTrail LookupEvents API limit) |
| BR-002 | Default time window is 24h when `--last` is not specified |
| BR-003 | Max 50 events per page (CloudTrail API limit); results are auto-paginated up to `--limit` |
| BR-004 | Only one LookupAttribute per query (CloudTrail API limitation); compound filters are not supported server-side |
| BR-005 | Results are sorted newest first |
| BR-006 | Queries are per-account, per-region — results only cover the configured AWS account and region |

## Non-Functional Requirements

- **Security:** Authentication via AWS SDK v2 default credential chain (env vars, shared config, IAM role). CLI mode supports aws-vault re-exec when `--profile` is set and credentials are expired. MCP mode relies on ambient credentials only (no aws-vault re-exec to avoid breaking the stdio session).
- **Performance:** CloudTrail `LookupEvents` is throttled to 2 requests/second per account/region. Tool uses exponential backoff on throttling (base 500ms, max 8s, max 5 retries).
- **Portability:** Single Go binary, cross-compilable. Primary target: macOS (arm64).

## Constraints & Limitations

- CloudTrail `LookupEvents` only returns **management events** (not data events)
- Max lookback is **90 days**; tool will reject/clamp ranges beyond this
- API returns max **50 events per page** (auto-paginated, stops at `--limit`)
- Only **one `LookupAttribute` per query** (API limitation) — compound filters not supported server-side
- Queries are **per-account, per-region** — results only cover the configured AWS account and region
- Some global-service events (IAM, STS, CloudFront, Route53) may only appear in `us-east-1`
- `LookupEvents` is throttled to **2 requests/second per account/region** — tool uses exponential backoff
- `ResourceName` lookup uses **exact name as CloudTrail records it** (may be bucket name, instance ID, etc. — not always full ARN)
- `Username` in CloudTrail may be IAM user, role session name, service principal, or `root` — for SSO/assumed-role, richer identity is parsed from event details
