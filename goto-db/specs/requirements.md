# goto-db — Requirements

**Date:** 2026-04-23
**Status:** Active
**Project:** go_projects/goto-db

## Overview

**Purpose:** A Go CLI tool that creates SSH tunnels to databases via a jump host and Jenkins agent, with a built-in DbGate UI running in Docker.

**Problem Statement:** Connecting to internal databases requires manually constructing multi-hop SSH tunnels through a jump host and Jenkins agent, then configuring a DB client. goto-db automates the entire chain — resolving the Jenkins agent, building the tunnel, launching a pre-configured DbGate web UI in Docker, and opening the browser — all in one command.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | CLI user who runs goto-db to connect to internal databases via SSH tunnel |

## Use Cases

### UC-DB-001: Connect to a database via SSH tunnel

- **Actor:** Developer
- **Preconditions:** SSH config with jump host entry (`~/.ssh/config`), Docker running, Jenkins agent reachable
- **Main Flow:**
  1. Developer runs `goto-db --db <name>` (optionally with `--env`, `--engine`, `--agent`)
  2. System parses CLI flags and validates input
  3. System loads config from `~/.config/goto-db/config.json`
  4. System resolves the Jenkins agent (flag → cache → default)
  5. System resolves DB target host from short name using domain convention
  6. System checks local port availability
  7. System prints connection info (host, engine, agent, local port)
  8. System starts DbGate Docker container with pre-configured connection
  9. System opens browser to DbGate UI
  10. System establishes SSH tunnel (blocks until cancelled)
  11. Developer presses Ctrl+C
  12. System stops SSH tunnel and removes Docker container
- **Alternative Flows:**
  - 2a. Missing `--db` and `--db-url` (and not `--refresh`) → print usage, exit 1
  - 2b. Both `--db` and `--db-url` provided → print error (mutually exclusive), exit 1
  - 2c. Unsupported engine → print error, exit 1
  - 6a. Local port already in use → print error suggesting `--local-port`, exit 1
  - 8a. Docker not available → print warning, continue without UI
- **Business Rules:**
  - BR-001, BR-002, BR-005
- **Acceptance Criteria:**
  - GIVEN Docker is running and SSH config is valid
  - WHEN the developer runs `goto-db --db audit`
  - THEN the system resolves `prod.primary.audit.db.viatorsystems.com`, starts DbGate, establishes an SSH tunnel on port 15432, and opens the browser

### UC-DB-002: Resolve Jenkins agent

- **Actor:** Developer
- **Preconditions:** None
- **Main Flow:**
  1. System checks for `--agent` flag — if set, uses it (auto-expands short names) and caches
  2. If `--refresh` flag, prompts user for agent hostname and caches
  3. If cached agent exists in `~/.config/goto-db/config.json`, uses it
  4. Otherwise, uses default agent (`jenkins-agent70215c.prod.svc.ue1.viatorsystems.com`) and caches
- **Alternative Flows:**
  - 1a. Both `--agent` and `--refresh` provided → print error (mutually exclusive), exit 1
- **Business Rules:**
  - BR-003, BR-006
- **Acceptance Criteria:**
  - GIVEN no cached agent exists
  - WHEN the developer runs `goto-db --db audit`
  - THEN the system uses the default agent and caches it in `~/.config/goto-db/config.json`

### UC-DB-003: Use fully qualified DB URL

- **Actor:** Developer
- **Preconditions:** SSH config with jump host, Docker running
- **Main Flow:**
  1. Developer runs `goto-db --db-url mydb.us-east-1.rds.amazonaws.com`
  2. System uses the URL directly as the DB host (bypasses domain convention)
  3. Continues with UC-DB-001 flow from step 3
- **Business Rules:**
  - BR-002, BR-005
- **Acceptance Criteria:**
  - GIVEN valid SSH config
  - WHEN the developer runs `goto-db --db-url mydb.us-east-1.rds.amazonaws.com`
  - THEN the system tunnels to `mydb.us-east-1.rds.amazonaws.com` using default postgres ports

### UC-DB-004: Refresh cached Jenkins agent

- **Actor:** Developer
- **Preconditions:** None
- **Main Flow:**
  1. Developer runs `goto-db --refresh`
  2. System prompts for Jenkins agent hostname (with default shown)
  3. System auto-expands short name if needed
  4. System saves to `~/.config/goto-db/config.json`
  5. System prints confirmation and exits (no tunnel started)
- **Business Rules:**
  - BR-003, BR-006
- **Acceptance Criteria:**
  - GIVEN an existing cached agent
  - WHEN the developer runs `goto-db --refresh` and enters `jenkins-agent80`
  - THEN the system expands to `jenkins-agent80.prod.svc.ue1.viatorsystems.com` and saves to config

## Business Rules

| ID | Rule |
|----|------|
| BR-001 | DB domain convention: `<env>.primary.<db>.db.viatorsystems.com` |
| BR-002 | Default engine is postgres (remote port 5432, local port 15432) |
| BR-003 | Jenkins agent is cached in `~/.config/goto-db/config.json` with timestamp |
| BR-004 | On Ctrl+C: stop SSH tunnel, stop and remove Docker container |
| BR-005 | Default ports — postgres: 5432/15432, mysql: 3306/13306 |
| BR-006 | Short agent names auto-expand with suffix `.prod.svc.ue1.viatorsystems.com` |

## Non-Functional Requirements

- **Security:** SSH authentication via system SSH (`~/.ssh/config`); DB credentials entered by user in DbGate UI (never handled by goto-db). SSH tunnel uses `ControlPath=none` to force a direct connection.
- **Portability:** Single Go binary, macOS (arm64) primary target. Requires Docker for DbGate UI.

## Constraints & Limitations

- Requires Docker to be running for DbGate UI (gracefully degrades with a warning if unavailable)
- SSH tunnel relies on system `ssh` binary and `~/.ssh/config` for jump host configuration
- Only postgres and mysql engines are supported
- One tunnel at a time per engine (default local ports would conflict)
- DbGate container uses fixed name `goto-db-dbgate` — only one instance at a time
- 2FA may be required during SSH connection (interactive prompt via stdin passthrough)
