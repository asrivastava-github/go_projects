# Spec-Driven Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish a multi-document spec-driven workflow across all Go projects — create reusable templates, a workflow guide, migrate the existing awsct spec, and retroactively create specs for all other projects.

**Architecture:** Each project gets a `specs/` directory with `requirements.md`, `design.md`, `tasks.md`. Repo root gets `spec-templates/` (reusable templates) and `SPEC-GUIDE.md` (workflow documentation). The awsct spec is the reference implementation — migrated first, then used as the pattern for all others.

**Tech Stack:** Markdown

**Spec:** `docs/superpowers/specs/2026-04-23-spec-driven-workflow-design.md`

---

## File Structure

```
go_projects/
├── spec-templates/
│   ├── requirements.md         # Reusable template
│   ├── design.md               # Reusable template
│   └── tasks.md                # Reusable template
├── SPEC-GUIDE.md               # Workflow guide
├── awsct/specs/
│   ├── requirements.md         # Migrated + restructured from existing spec
│   ├── design.md               # Migrated + restructured from existing spec
│   └── tasks.md                # Migrated from existing plan
├── gcpat/specs/
│   ├── requirements.md         # New — created from code examination
│   ├── design.md               # New
│   └── tasks.md                # New
├── goto/specs/
│   ├── requirements.md         # New
│   ├── design.md               # New
│   └── tasks.md                # New
├── goto-db/specs/
│   ├── requirements.md         # New
│   ├── design.md               # New
│   └── tasks.md                # New
├── go-s3-downloader/specs/
│   ├── requirements.md         # New
│   ├── design.md               # New
│   └── tasks.md                # New
├── ip-finder/specs/
│   ├── requirements.md         # New
│   ├── design.md               # New
│   └── tasks.md                # New
└── docs/superpowers/specs/     # Old specs removed after migration
```

---

### Task 1: Create Reusable Templates

**Files:**
- Create: `spec-templates/requirements.md`
- Create: `spec-templates/design.md`
- Create: `spec-templates/tasks.md`

- [ ] **Step 1: Create `spec-templates/requirements.md`**

```markdown
# <Project Name> — Requirements

**Date:** YYYY-MM-DD
**Status:** Draft | Active | Deprecated
**Project:** <repo/project-path>

## Overview

**Purpose:** <One-line description of what this project does and why it exists>

**Problem Statement:** <What problem does this solve? What was the user doing before?>

## Actors

| Actor | Description |
|-------|-------------|
| <Actor Name> | <Who/what is this and how do they interact with the system> |

## Use Cases

### UC-<PREFIX>-001: <Title>

- **Actor:** <Who triggers this>
- **Preconditions:** <What must be true before this use case can execute>
- **Main Flow:**
  1. <Step 1>
  2. <Step 2>
  3. <Step 3>
- **Alternative Flows:**
  - <step>a. <condition> → <what happens>
- **Business Rules:**
  - BR-001: <Rule description>
- **Acceptance Criteria:**
  - GIVEN <context>
  - WHEN <action>
  - THEN <expected outcome>

<!-- Repeat for additional use cases: UC-<PREFIX>-002, UC-<PREFIX>-003, etc. -->

## Non-Functional Requirements

<!-- Remove sections that don't apply -->

- **Performance:** <Response time, throughput expectations>
- **Security:** <Authentication, authorization, credential handling>
- **Portability:** <OS support, deployment targets>

## Constraints & Limitations

- <Known API limits, platform restrictions, scope boundaries>
```

- [ ] **Step 2: Create `spec-templates/design.md`**

```markdown
# <Project Name> — Design

**Date:** YYYY-MM-DD
**Status:** Draft | Active | Deprecated

## Architecture Overview

<High-level description of how the system is structured. Include a text diagram showing component relationships.>

```
Component A  ──┐
               ├──▶  Core Logic  ──▶  External Service
Component B  ──┘
```

## Project Structure

```
project/
├── main.go              # <Description>
├── cmd/                  # <Description>
├── pkg/                  # <Description>
└── ...
```

## Components

### Component: <Name>

- **Responsibility:** <What it does>
- **Interfaces:** <How other components interact with it>
- **Key Decisions:** <Why this approach was chosen over alternatives>

<!-- Repeat for each component -->

## Data Flow

1. <Input source> provides <what>
2. <Component> processes <how>
3. <Output> is delivered as <format>

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| <package> | <why it's needed> | <version> |

## Authentication & Security

<!-- Remove if not applicable -->

<How credentials/secrets are handled. Auth flow description.>

## Error Handling Strategy

<!-- Remove if not applicable -->

<How errors propagate, what the user sees, retry logic if any.>

## Future Considerations

<!-- Optional: known extension points, areas for future growth -->

- <Potential enhancement and what it would require>
```

- [ ] **Step 3: Create `spec-templates/tasks.md`**

```markdown
# <Project Name> — Tasks

**Date:** YYYY-MM-DD
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: <Title>

**Traces to:** UC-<PREFIX>-NNN
**Files:** `path/to/file1.go`, `path/to/file2.go`

- [x] Step 1: <Description>
- [ ] Step 2: <Description>
- [ ] Step 3: <Description>

<!-- Repeat for additional tasks -->

## Changelog

| Date | Change | Author |
|------|--------|--------|
| YYYY-MM-DD | Initial spec created | <initials> |
```

- [ ] **Step 4: Verify all three templates are created**

```bash
ls spec-templates/
```

Expected: `requirements.md  design.md  tasks.md`

- [ ] **Step 5: Commit**

```bash
git add spec-templates/
git commit -m "docs: add reusable spec templates (requirements, design, tasks)"
```

---

### Task 2: Create SPEC-GUIDE.md

**Files:**
- Create: `SPEC-GUIDE.md`

- [ ] **Step 1: Create `SPEC-GUIDE.md`**

```markdown
# Spec-Driven Development Guide

This repository follows a **spec-driven development** workflow. Every project maintains structured specifications that serve as the source of truth for what the software does and how it's built.

## Why Specs?

- **Intent survives technology churn** — tools and frameworks change; specs capture *what* and *why*
- **AI agents work better** — structured specs with use cases, flows, and acceptance criteria give AI unambiguous context
- **Traceability** — every task links back to a requirement; you always know *why* code exists
- **Verifiability** — acceptance criteria in GIVEN/WHEN/THEN format are directly testable

## Spec Structure

Each project has a `specs/` directory with three files:

| File | Purpose | Answers |
|------|---------|---------|
| `requirements.md` | Use cases, actors, acceptance criteria | **What** should the software do? |
| `design.md` | Architecture, components, data flow, dependencies | **How** is it built? |
| `tasks.md` | Implementation tasks with traceability | **What work** was done and why? |

## Workflows

### Starting a New Project

1. Copy templates: `cp -r spec-templates/ <project>/specs/`
2. Fill in `requirements.md` first — define actors, use cases, acceptance criteria
3. Fill in `design.md` — decide architecture, components, dependencies
4. Create tasks in `tasks.md` — link each task to use case IDs
5. Start coding — the spec is your guide

### Working on an Existing Project

1. **Read specs first** — understand the intent before changing code
2. **Update specs when behavior changes** — if you change what the software does, update `requirements.md`; if you change how it's built, update `design.md`
3. **Log changes** — add entries to the Changelog in `tasks.md`

### Using Specs with AI Agents

Point the agent at `requirements.md` + `design.md` before asking it to generate code:

```
Read specs/requirements.md and specs/design.md, then implement UC-XXX-003
```

The structured use cases give AI agents actors, flows, business rules, and acceptance criteria — far more effective than freeform prompts.

### Keeping Specs in Sync

- When code changes, ask: "Does the spec still match?"
- Specs don't need to document every implementation detail — they capture **intent and behavior**
- The Changelog in `tasks.md` tracks how the project evolves over time

### Splitting a Project to Its Own Repo

The `specs/` directory is self-contained. When you move a project to its own repository, the specs travel with it unchanged.

## Use Case ID Convention

- Format: `UC-<PREFIX>-NNN` (e.g., `UC-WHO-001`, `UC-DL-001`)
- Business Rules: `BR-NNN` scoped per use case
- Choose a short, memorable prefix per project

## Templates

Reusable templates live in `spec-templates/`. Copy them to start a new project's specs.

## Influences

This workflow draws from:
- **AI Unified Process (AIUP)** by Simon Martinelli — structured use cases with actors, flows, business rules
- **Spec-Driven Development** principles — specs as source of truth, not code
- **GIVEN/WHEN/THEN** acceptance criteria from BDD
```

- [ ] **Step 2: Commit**

```bash
git add SPEC-GUIDE.md
git commit -m "docs: add spec-driven development workflow guide"
```

---

### Task 3: Migrate awsct Spec

**Files:**
- Create: `awsct/specs/requirements.md`
- Create: `awsct/specs/design.md`
- Create: `awsct/specs/tasks.md`
- Remove: `docs/superpowers/specs/2026-04-17-awsct-design.md`
- Remove: `docs/superpowers/plans/2026-04-17-awsct.md`

Split the existing monolithic `2026-04-17-awsct-design.md` into the three-file format and restructure content to match the template. This serves as the **reference implementation** that validates the templates.

- [ ] **Step 1: Create `awsct/specs/requirements.md`**

Extract from the existing spec: Purpose, CLI usage (as use cases), MCP tools (as use cases), global flags, authentication, constraints & limitations. Restructure into actors, use cases with flows, business rules, and acceptance criteria.

**Actors:**
- Developer (CLI user)
- AI Agent (MCP consumer)

**Use Cases to define:**
- UC-WHO-001: Find who performed a specific action
- UC-USER-001: Find what a specific user did
- UC-RES-001: Find what happened to a specific resource
- UC-MCP-001: Serve as MCP server for AI agents

**Business Rules from existing spec:**
- BR-001: Max lookback 90 days
- BR-002: Default time window 24h
- BR-003: Max 50 events per page (auto-paginated)
- BR-004: One LookupAttribute per query
- BR-005: Results sorted newest first

- [ ] **Step 2: Create `awsct/specs/design.md`**

Extract from existing spec: Architecture diagram, project structure, core query logic table, output formats (table + JSON), MCP server details, dependencies table.

- [ ] **Step 3: Create `awsct/specs/tasks.md`**

Migrate from `docs/superpowers/plans/2026-04-17-awsct.md`. Convert the 8 tasks into the new format with traceability links. Mark all tasks as complete (since awsct is already built).

- [ ] **Step 4: Remove old spec files**

```bash
git rm docs/superpowers/specs/2026-04-17-awsct-design.md
git rm docs/superpowers/plans/2026-04-17-awsct.md
```

- [ ] **Step 5: Verify migration**

```bash
ls awsct/specs/
```

Expected: `requirements.md  design.md  tasks.md`

- [ ] **Step 6: Commit**

```bash
git add awsct/specs/ && git rm docs/superpowers/specs/2026-04-17-awsct-design.md docs/superpowers/plans/2026-04-17-awsct.md
git commit -m "docs(awsct): migrate spec to per-project specs/ structure"
```

---

### Task 4: Create gcpat Specs

**Files:**
- Create: `gcpat/specs/requirements.md`
- Create: `gcpat/specs/design.md`
- Create: `gcpat/specs/tasks.md`

gcpat is the GCP equivalent of awsct — queries GCP Cloud Audit Logs by method, principal, or resource. Has CLI + MCP modes. Uses cobra, GCP Logging API, tabwriter.

**Reference source files to examine:**
- `gcpat/cmd/root.go` — flags, output helpers
- `gcpat/cmd/who.go`, `user.go`, `resource.go` — CLI subcommands
- `gcpat/cmd/serve.go` — MCP server
- `gcpat/pkg/auditlog/` — core query logic, types, duration parser

**Actors:** Developer (CLI), AI Agent (MCP)

**Use Cases:**
- UC-WHO-001: Find who performed a specific GCP method
- UC-USER-001: Find what a principal did
- UC-RES-001: Find what happened to a resource
- UC-MCP-001: Serve as MCP server

**Key flags:** `--project` (required), `--last`, `--json`, `--limit`

- [ ] **Step 1: Read all gcpat source files to extract behavior**

Read: `gcpat/cmd/who.go`, `gcpat/cmd/user.go`, `gcpat/cmd/resource.go`, `gcpat/cmd/serve.go`, `gcpat/pkg/auditlog/lookup.go`, `gcpat/pkg/auditlog/types.go`, `gcpat/pkg/auditlog/client.go`, `gcpat/pkg/auditlog/duration.go`

- [ ] **Step 2: Create `gcpat/specs/requirements.md`**

- [ ] **Step 3: Create `gcpat/specs/design.md`**

- [ ] **Step 4: Create `gcpat/specs/tasks.md`** (all tasks marked complete)

- [ ] **Step 5: Commit**

```bash
git add gcpat/specs/
git commit -m "docs(gcpat): add spec-driven requirements, design, and tasks"
```

---

### Task 5: Create goto Specs

**Files:**
- Create: `goto/specs/requirements.md`
- Create: `goto/specs/design.md`
- Create: `goto/specs/tasks.md`

goto is a single-file Go CLI that fetches EC2 instances by Role tag and SSHs into them via FQDN tag. Uses aws-vault re-exec for credential refresh.

**Reference source files:**
- `goto/main.go` — entire implementation
- `goto/README.md` — usage docs
- `goto/main_test.go` — existing tests

**Actors:** Developer

**Use Cases:**
- UC-GOTO-001: SSH into an EC2 instance by role and environment
- UC-GOTO-002: Select from multiple matching instances

**Key behavior:** args are `<role> <env> [region]`, auto re-exec via aws-vault, filters by Role tag + running state, prompts if multiple instances, SSHs via FQDN tag.

- [ ] **Step 1: Read `goto/main.go` and `goto/main_test.go`**

- [ ] **Step 2: Create `goto/specs/requirements.md`**

- [ ] **Step 3: Create `goto/specs/design.md`**

- [ ] **Step 4: Create `goto/specs/tasks.md`** (all tasks marked complete)

- [ ] **Step 5: Commit**

```bash
git add goto/specs/
git commit -m "docs(goto): add spec-driven requirements, design, and tasks"
```

---

### Task 6: Create goto-db Specs

**Files:**
- Create: `goto-db/specs/requirements.md`
- Create: `goto-db/specs/design.md`
- Create: `goto-db/specs/tasks.md`

goto-db creates SSH tunnels to databases via jump host + Jenkins agent, with built-in DbGate UI in Docker. Well-structured with `internal/` packages.

**Reference source files:**
- `goto-db/README.md` — comprehensive docs with architecture diagram
- `goto-db/internal/` — app, cli, config, agent, db, ssh, ui packages

**Actors:** Developer

**Use Cases:**
- UC-DB-001: Connect to a database via SSH tunnel
- UC-DB-002: Resolve Jenkins agent (default/cache/flag/prompt)
- UC-DB-003: Launch DbGate UI in Docker
- UC-DB-004: Refresh cached Jenkins agent

**Key behavior:** DB domain convention (`prod.primary.audit.db.viatorsystems.com`), tunnel chain (localhost → jump → jenkins-agent → db), Docker DbGate, signal handling (Ctrl+C cleanup).

- [ ] **Step 1: Read key goto-db source files**

Read: `goto-db/cmd/goto-db/main.go`, `goto-db/internal/app/run.go`, `goto-db/internal/cli/`, `goto-db/internal/db/`, `goto-db/internal/ssh/tunnel.go`, `goto-db/internal/ui/dbgate.go`

- [ ] **Step 2: Create `goto-db/specs/requirements.md`**

- [ ] **Step 3: Create `goto-db/specs/design.md`**

- [ ] **Step 4: Create `goto-db/specs/tasks.md`** (all tasks marked complete)

- [ ] **Step 5: Commit**

```bash
git add goto-db/specs/
git commit -m "docs(goto-db): add spec-driven requirements, design, and tasks"
```

---

### Task 7: Create go-s3-downloader Specs

**Files:**
- Create: `go-s3-downloader/specs/requirements.md`
- Create: `go-s3-downloader/specs/design.md`
- Create: `go-s3-downloader/specs/tasks.md`

go-s3-downloader is a CLI that downloads files from S3 buckets, supporting prod/dev environments. Uses AWS SDK v2, internal packages for config, s3client, and download handler.

**Reference source files:**
- `go-s3-downloader/README.md` — usage docs
- `go-s3-downloader/main.go`
- `go-s3-downloader/internal/config/config.go`
- `go-s3-downloader/internal/s3client/client.go`
- `go-s3-downloader/internal/handler/download.go`

**Actors:** Developer

**Use Cases:**
- UC-DL-001: Download files from S3 bucket/prefix
- UC-DL-002: Select environment (prod/dev)

**Key behavior:** `-env` flag (prod/dev), `-bucket` flag (bucket/prefix), downloads to `~/s3_files/`, preserves S3 folder structure.

- [ ] **Step 1: Read go-s3-downloader source files**

- [ ] **Step 2: Create `go-s3-downloader/specs/requirements.md`**

- [ ] **Step 3: Create `go-s3-downloader/specs/design.md`**

- [ ] **Step 4: Create `go-s3-downloader/specs/tasks.md`** (all tasks marked complete)

- [ ] **Step 5: Commit**

```bash
git add go-s3-downloader/specs/
git commit -m "docs(go-s3-downloader): add spec-driven requirements, design, and tasks"
```

---

### Task 8: Create ip-finder Specs

**Files:**
- Create: `ip-finder/specs/requirements.md`
- Create: `ip-finder/specs/design.md`
- Create: `ip-finder/specs/tasks.md`

ip-finder identifies AWS resources (EC2, Lambda, EKS pods) by internal IP address. Most complex project — has AWS ENI lookup, resource classification, K8s pod search, smart skip logic.

**Reference source files:**
- `ip-finder/README.md` — comprehensive docs
- `ip-finder/cmd/root.go` — CLI flags
- `ip-finder/pkg/aws/` — eni.go, classifier.go, instance.go, credentials.go
- `ip-finder/pkg/k8s/` — pods.go
- `ip-finder/pkg/finder/finder.go` — orchestration logic

**Actors:** Developer

**Use Cases:**
- UC-IP-001: Identify an AWS resource by IP address
- UC-IP-002: Find EC2 instance details for an IP
- UC-IP-003: Find Kubernetes pod by IP
- UC-IP-004: Smart skip K8s search for non-EC2 resources

**Key behavior:** ENI search (secondary IPs + prefix delegation), resource classification (Lambda, ELB, NAT GW, etc.), smart K8s skip, aws-vault auto re-exec, tabular output.

- [ ] **Step 1: Read ip-finder source files**

- [ ] **Step 2: Create `ip-finder/specs/requirements.md`**

- [ ] **Step 3: Create `ip-finder/specs/design.md`**

- [ ] **Step 4: Create `ip-finder/specs/tasks.md`** (all tasks marked complete)

- [ ] **Step 5: Commit**

```bash
git add ip-finder/specs/
git commit -m "docs(ip-finder): add spec-driven requirements, design, and tasks"
```

---

### Task 9: Update Repo README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update `README.md` to reference the spec-driven workflow**

Add a section pointing to `SPEC-GUIDE.md` and mentioning that each project has a `specs/` directory.

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with spec-driven workflow reference"
```
