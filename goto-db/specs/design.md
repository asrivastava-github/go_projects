# goto-db — Design

**Date:** 2026-04-23
**Status:** Active

## Architecture Overview

Single Go binary with a linear orchestration flow: parse CLI → resolve agent → resolve DB target → start UI → run tunnel → cleanup on exit.

```
main.go (entrypoint + signals)
  └──▶ app/run.go (orchestration)
         ├──▶ cli/         (parse flags, validate)
         ├──▶ config/      (load/save ~/.config/goto-db/)
         ├──▶ agent/       (resolve Jenkins agent)
         ├──▶ db/          (resolve DB host + ports)
         ├──▶ ui/          (Docker DbGate container)
         └──▶ ssh/         (SSH tunnel via system ssh)
```

## Project Structure

```
goto-db/
├── cmd/goto-db/
│   └── main.go              # Entrypoint, signal handling (SIGINT/SIGTERM)
├── internal/
│   ├── app/
│   │   └── run.go           # Orchestration: parse → resolve → UI → tunnel → cleanup
│   ├── cli/
│   │   ├── options.go       # Options struct (DBName, Environment, Engine, etc.)
│   │   ├── parse.go         # Flag parsing with flag.FlagSet, validation
│   │   ├── parse_test.go    # CLI parsing tests
│   │   └── usage.go         # Help/usage text with examples
│   ├── config/
│   │   └── config.go        # Config struct, Load/Save to ~/.config/goto-db/config.json
│   ├── agent/
│   │   └── resolver.go      # Jenkins agent resolution (flag → refresh → cache → default)
│   ├── db/
│   │   ├── target.go        # ResolveTarget: short name → FQDN, port mapping
│   │   ├── target_test.go   # Target resolution tests
│   │   ├── engine.go        # DefaultPort, DefaultLocalPort, DefaultClient per engine
│   │   └── engine_test.go   # Engine defaults tests
│   ├── ssh/
│   │   └── tunnel.go        # CheckPortAvailable, RunTunnel (ssh -N -L with ControlPath=none)
│   └── ui/
│       └── dbgate.go        # StartUI, StopUI, Docker container lifecycle, browser open
├── go.mod
└── README.md
```

## Components

### Component: Entrypoint (`cmd/goto-db/main.go`)

- **Responsibility:** Set up signal context (SIGINT/SIGTERM) and delegate to `app.Run()`.
- **Interfaces:** `app.Run(ctx, os.Args[1:])` — passes context and args.
- **Key Decisions:** Uses `signal.NotifyContext` for clean cancellation propagation through the entire call chain.

### Component: Orchestration (`internal/app`)

- **Responsibility:** Coordinate the full workflow: parse CLI → load config → resolve agent → resolve DB target → check port → start UI → run tunnel → cleanup.
- **Interfaces:** Single public function `Run(ctx, args) error`.
- **Key Decisions:** Linear sequential flow (no goroutines) — each step depends on the previous. If `--refresh` is the only intent (no `--db`/`--db-url`), exits early after agent resolution.

### Component: CLI (`internal/cli`)

- **Responsibility:** Parse and validate CLI flags using `flag.FlagSet`.
- **Interfaces:** `Parse(args) (*Options, error)` returns typed options.
- **Key Decisions:** Uses stdlib `flag` (no cobra needed — single command, no subcommands). Validation rules: `--db` and `--db-url` are mutually exclusive; `--agent` and `--refresh` are mutually exclusive; engine must be `postgres` or `mysql`; `--user` is required.

### Component: Config (`internal/config`)

- **Responsibility:** Persist and load Jenkins agent hostname in `~/.config/goto-db/config.json`.
- **Interfaces:** `Load() (*Config, error)`, `Save(*Config) error`.
- **Key Decisions:** Uses `os.UserConfigDir()` for cross-platform config path. Creates directory on first save. Returns empty config (not error) when file doesn't exist.

### Component: Agent Resolver (`internal/agent`)

- **Responsibility:** Resolve the Jenkins agent hostname with priority: explicit `--agent` flag → `--refresh` (prompt) → cached config → hardcoded default.
- **Interfaces:** `Resolve(ctx, opts, cfg) (string, error)`.
- **Key Decisions:** Short agent names (no dots) auto-expand with `.prod.svc.ue1.viatorsystems.com` suffix. Every resolution path caches the result.

### Component: DB Target (`internal/db`)

- **Responsibility:** Resolve the database hostname and port mapping from CLI options.
- **Interfaces:** `ResolveTarget(opts) (*Target, error)` returns host, remote port, local port, engine.
- **Key Decisions:** Short name resolution uses convention `<env>.primary.<db>.db.viatorsystems.com`. `--db-url` bypasses convention and uses the URL directly. `--local-port` overrides the engine default.

### Component: SSH Tunnel (`internal/ssh`)

- **Responsibility:** Check local port availability and run the SSH tunnel process.
- **Interfaces:** `CheckPortAvailable(port) error`, `RunTunnel(ctx, Spec) error`.
- **Key Decisions:** Shells out to system `ssh` (not a Go SSH library) to leverage `~/.ssh/config` (jump hosts, 2FA, keys). Uses `ControlPath=none` to force a direct connection — without this, SSH mux client exits immediately. `ExitOnForwardFailure=yes` ensures clean error on port conflicts. `ServerAliveInterval=60` with `ServerAliveCountMax=3` for keepalive.

### Component: DB UI (`internal/ui`)

- **Responsibility:** Manage the DbGate Docker container lifecycle and browser launch.
- **Interfaces:** `StartUI(ctx, ConnectionParams) error`, `StopUI()`, `OpenBrowser(url)`, `BrowserURL(params) string`.
- **Key Decisions:** Uses `dbgate/dbgate` Docker image with environment variables for pre-configured connection. Container maps port 3000 to host port 8978. Uses `host.docker.internal` for the container to reach the host's SSH tunnel. `PASSWORD_MODE_db=askUser` so credentials are entered in the UI. Waits up to 30 seconds for DbGate readiness (HTTP polling).

## Data Flow

1. User provides `--db <name>` (or `--db-url`) with optional `--env`, `--engine`, `--agent`
2. CLI parses and validates flags into `Options` struct
3. Config loads cached Jenkins agent from `~/.config/goto-db/config.json`
4. Agent resolver determines the Jenkins agent hostname (flag/prompt/cache/default)
5. DB target resolver builds FQDN (`prod.primary.audit.db.viatorsystems.com`) and maps ports
6. SSH module checks local port availability
7. UI module starts DbGate Docker container with connection pre-configured
8. SSH module runs `ssh -N -L <local>:<db-host>:<remote> <agent>` (blocks until Ctrl+C)
9. On cancellation: UI module stops and removes Docker container

### Tunnel Chain

```
localhost:<local-port> → jump (via ~/.ssh/config) → <jenkins-agent> → <db-host>:<db-port>
```

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| Go stdlib (`flag`, `os/exec`, `net`, `net/http`) | CLI parsing, process exec, port checks, HTTP polling | Go 1.26.1 |
| System `ssh` | SSH tunnel (leverages ~/.ssh/config) | system |
| Docker | DbGate UI container runtime | system |
| `dbgate/dbgate` | Web-based database UI (Docker image) | latest |

## Error Handling Strategy

- Missing required flags → print usage and error message, exit 1
- Mutually exclusive flags → print specific error, exit 1
- Unsupported engine → print error with valid options, exit 1
- Config file missing → return empty config (not error), use defaults
- Config file corrupt → return parse error
- Local port in use → print error suggesting `--local-port` flag, exit 1
- Docker not available → print warning, continue without UI
- DbGate startup timeout (30s) → print warning, continue
- SSH tunnel failure → print error, exit 1
- Context cancellation (Ctrl+C) → clean shutdown (stop tunnel, remove container), exit 0

## Future Considerations

- Support for additional database engines (e.g., MongoDB, Redis)
- Multiple simultaneous tunnels with different local ports
- Connection profiles/bookmarks beyond just agent caching
