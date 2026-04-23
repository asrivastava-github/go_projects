# goto-db — Tasks

**Date:** 2026-04-23
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Project Scaffold

**Traces to:** UC-DB-001
**Files:** `go.mod`, `cmd/goto-db/main.go`

- [x] Step 1: Create go.mod (`go mod init goto-db`)
- [x] Step 2: Create cmd/goto-db/main.go with signal handling (SIGINT/SIGTERM) and delegation to app.Run()
- [x] Step 3: Commit

## Task 2: CLI Parsing & Validation

**Traces to:** UC-DB-001, UC-DB-002, UC-DB-003, UC-DB-004
**Files:** `internal/cli/options.go`, `internal/cli/parse.go`, `internal/cli/usage.go`, `internal/cli/parse_test.go`

- [x] Step 1: Create options.go — Options struct (DBName, Environment, Engine, JenkinsAgent, Refresh, LocalPort, User, DBURL)
- [x] Step 2: Create parse.go — flag.FlagSet parsing with defaults (env=prod, engine=postgres, user=current), validation rules (--db/--db-url required, mutually exclusive pairs)
- [x] Step 3: Create usage.go — help text with usage patterns, examples, and tunnel chain description
- [x] Step 4: Create parse_test.go — tests for valid args, db-url mode, missing db, mutual exclusivity, invalid engine, refresh-only, mysql engine
- [x] Step 5: Verify tests pass
- [x] Step 6: Commit

## Task 3: Config Management

**Traces to:** UC-DB-002, UC-DB-004
**Files:** `internal/config/config.go`

- [x] Step 1: Create config.go — Config struct (JenkinsAgent, UpdatedAt), Load/Save functions using os.UserConfigDir()
- [x] Step 2: Handle missing file gracefully (return empty config), create directory on first save
- [x] Step 3: Verify compilation
- [x] Step 4: Commit

## Task 4: Jenkins Agent Resolver

**Traces to:** UC-DB-002, UC-DB-004
**Files:** `internal/agent/resolver.go`

- [x] Step 1: Create resolver.go — Resolve function with priority chain: --agent flag → --refresh prompt → cached config → default agent
- [x] Step 2: Implement expandAgent for short name auto-expansion (append .prod.svc.ue1.viatorsystems.com)
- [x] Step 3: Implement promptForAgent for interactive --refresh flow
- [x] Step 4: Verify compilation
- [x] Step 5: Commit

## Task 5: DB Target Resolution

**Traces to:** UC-DB-001, UC-DB-003
**Files:** `internal/db/target.go`, `internal/db/target_test.go`, `internal/db/engine.go`, `internal/db/engine_test.go`

- [x] Step 1: Create engine.go — DefaultPort, DefaultLocalPort, DefaultClient functions for postgres and mysql
- [x] Step 2: Create engine_test.go — table-driven tests for all engine functions
- [x] Step 3: Create target.go — ResolveTarget function with domain convention (<env>.primary.<db>.db.viatorsystems.com) and --db-url bypass
- [x] Step 4: Create target_test.go — tests for short name resolution, db-url mode, custom local port, mysql engine
- [x] Step 5: Verify tests pass
- [x] Step 6: Commit

## Task 6: SSH Tunnel

**Traces to:** UC-DB-001
**Files:** `internal/ssh/tunnel.go`

- [x] Step 1: Create tunnel.go — Spec struct, CheckPortAvailable (net.Listen probe), RunTunnel (exec ssh -N -L with ControlPath=none, ExitOnForwardFailure, ServerAlive keepalive)
- [x] Step 2: Handle context cancellation (return nil on ctx.Err())
- [x] Step 3: Verify compilation
- [x] Step 4: Commit

## Task 7: DbGate UI (Docker)

**Traces to:** UC-DB-001
**Files:** `internal/ui/dbgate.go`

- [x] Step 1: Create dbgate.go — ConnectionParams struct, StartUI (docker run with env vars for pre-configured connection)
- [x] Step 2: Implement StopUI (docker stop + rm), isDockerAvailable, isContainerRunning checks
- [x] Step 3: Implement waitForReady (HTTP polling with 30s timeout)
- [x] Step 4: Implement BrowserURL, PrintConnectionInfo, OpenBrowser
- [x] Step 5: Verify compilation
- [x] Step 6: Commit

## Task 8: Orchestration

**Traces to:** UC-DB-001, UC-DB-002, UC-DB-003, UC-DB-004
**Files:** `internal/app/run.go`

- [x] Step 1: Create run.go — Run function wiring all components: parse → config → agent → target → port check → UI → tunnel → cleanup
- [x] Step 2: Handle --refresh early exit (no tunnel when only refreshing agent)
- [x] Step 3: Handle UI failure gracefully (warning, continue without UI)
- [x] Step 4: Verify build
- [x] Step 5: Commit

## Task 9: README & Documentation

**Traces to:** UC-DB-001, UC-DB-002, UC-DB-003, UC-DB-004
**Files:** `README.md`

- [x] Step 1: Create README.md with install, usage examples, tunnel chain, DB domain convention, default ports, architecture diagram, and config info
- [x] Step 2: Commit

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-04-23 | Initial spec created; all tasks complete | AS |
