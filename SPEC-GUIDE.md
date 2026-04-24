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

### Making Incremental Changes

When updating a spec, **tell the agent exactly what changed** — not "here's the whole spec, make it work." This prevents unnecessary rewrites.

**Do this:**
```
I added BR-007 to UC-WHO-001 in requirements.md. Update only the affected code.
```
```
I changed the Main Flow step 3 in UC-DL-001 — download files should go to
~/downloads/ instead of ~/s3_files/. Update the handler accordingly.
```
```
Read specs/requirements.md. I've added UC-RES-002 (a new use case). 
Implement only this new use case — don't touch existing code.
```

**Don't do this:**
```
Here's my updated requirements.md — implement it.
```
This risks the agent regenerating everything from scratch instead of making a targeted change.

**Why this works:** Use case IDs (`UC-XXX-NNN`), business rule IDs (`BR-NNN`), and the traceability in `tasks.md` (which maps tasks → use cases → files) give the agent precise scope. It knows *which* use case changed, *which* files implement it, and can make a surgical edit.

**After making changes:**
1. Update the spec first (requirements.md or design.md)
2. Tell the agent what specifically changed
3. Log the change in `tasks.md` Changelog

### Keeping Specs in Sync

- When code changes, ask: "Does the spec still match?"
- Specs don't need to document every implementation detail — they capture **intent and behavior**
- The Changelog in `tasks.md` tracks how the project evolves over time

### Splitting a Project to Its Own Repo

The `specs/` directory is self-contained. When you move a project to its own repository, the specs travel with it unchanged.

## Maintenance & Security

Dependency and version management applies equally to all projects. Individual project specs don't repeat this policy — they follow it.

### Dependencies

- **Quarterly review:** Check all `go.mod` dependencies for security advisories and updates
- **CVE monitoring:** When a CVE is reported for a dependency, upgrade promptly — track the work in the affected project's `tasks.md`
- **Go version:** Keep Go version current across all projects; update `go.mod` when a new stable release lands
- **Audit command:** Run `go mod tidy` and `govulncheck ./...` periodically to catch stale or vulnerable dependencies

### Version Pinning

- Each project's `design.md` Dependencies table records current versions
- When upgrading, update the Dependencies table in `design.md` and log the change in `tasks.md` Changelog

### Per-Project Security Requirements

Project-specific security constraints (e.g., "Must use AWS SDK v2; v1 is EOL", "No plaintext credentials in config") belong in that project's `requirements.md` under Non-Functional Requirements.

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
