# Spec-Driven Development Workflow for go_projects

**Date:** 2026-04-23
**Status:** Draft
**Scope:** All projects in go_projects repository

## Purpose

Establish a spec-driven development workflow across all Go projects in this repository. Each project becomes a self-contained unit with its own multi-document specification (requirements, design, tasks) that serves as the living source of truth. The specs travel with the project if it moves to a separate repo.

This draws from the **AI Unified Process (AIUP)** by Simon Martinelli — specifically structured use cases with actors, flows, business rules, and traceability — while staying lightweight enough for CLI utility tools.

## Goals

1. **Spec-anchored development** — specs are created before code and kept in sync as features evolve
2. **Self-contained projects** — each project carries its own specs in `docs/`
3. **Standardized structure** — reusable templates ensure consistency across projects
4. **Traceability** — every task traces back to a requirement; every requirement is testable
5. **Scalability** — the three-file minimum can grow with additional spec files as needed
6. **AI-agent friendly** — structured specs give AI coding agents unambiguous context for code generation

## Design

### Per-Project Spec Structure

Each project gets a `docs/` directory with three markdown files:

```
<project>/
├── docs/
│   ├── requirements.md    # WHAT — use cases, actors, acceptance criteria
│   ├── design.md          # HOW — architecture, components, data flow
│   └── tasks.md           # TRACK — implementation tasks with traceability
├── main.go
├── cmd/
├── pkg/
└── ...
```

**Extensibility:** Projects may add additional spec files (e.g., `api-contract.md`, `testing-strategy.md`, `migration.md`) alongside the three core files without breaking the standard.

### requirements.md — What the Project Does

Captures intent, actors, use cases, and acceptance criteria. Borrows structured use case format from AIUP.

**Sections:**

| Section | Purpose | Required |
|---------|---------|----------|
| Overview | One-line purpose + target users | Yes |
| Actors | Who or what interacts with the system | Yes |
| Use Cases | Structured use cases with IDs, flows, business rules | Yes |
| Non-Functional Requirements | Performance, security, portability | If applicable |
| Constraints & Limitations | API limits, known restrictions | If applicable |

**Use Case Format (AIUP-influenced):**

Each use case follows this structure:

```markdown
### UC-<PREFIX>-NNN: <Title>
- **Actor:** <Who triggers this>
- **Preconditions:** <What must be true before>
- **Main Flow:**
  1. Step 1
  2. Step 2
  3. Step 3
- **Alternative Flows:**
  - <step>a. <condition> → <what happens>
- **Business Rules:**
  - BR-NNN: <rule description>
- **Acceptance Criteria:**
  - GIVEN <context>
  - WHEN <action>
  - THEN <expected outcome>
```

**Use Case ID Convention:**
- Format: `UC-<PROJECT_PREFIX>-NNN`
- Examples: `UC-WHO-001` (awsct who command), `UC-DL-001` (go-s3-downloader), `UC-GOTO-001` (goto)
- Business Rules: `BR-NNN` scoped per use case

### design.md — How the Project is Built

Captures architecture decisions, component design, data flow, and dependencies.

**Sections:**

| Section | Purpose | Required |
|---------|---------|----------|
| Architecture Overview | High-level system description + text diagram | Yes |
| Project Structure | File tree with descriptions | Yes |
| Components | Per-component responsibility, interfaces, key decisions | Yes |
| Data Flow | Input → processing → output flow | Yes |
| Dependencies | Table of external dependencies with purpose | Yes |
| Authentication & Security | How credentials/secrets are handled | If applicable |
| Error Handling Strategy | How errors propagate, what users see | If applicable |
| Future Considerations | Known extension points, scalability notes | Optional |

### tasks.md — Implementation Tracking with Traceability

Tracks implementation work with explicit links back to requirements.

**Sections:**

| Section | Purpose | Required |
|---------|---------|----------|
| Status Legend | Checkbox conventions | Yes |
| Tasks | Grouped implementation steps with traceability | Yes |
| Changelog | Date-stamped record of spec/task changes | Yes |

**Task Format:**

```markdown
## Task N: <Title>
**Traces to:** UC-<PREFIX>-NNN, UC-<PREFIX>-NNN
**Files:** `path/to/file1.go`, `path/to/file2.go`

- [x] Step 1: Description
- [ ] Step 2: Description
```

The **"Traces to"** line is the key AIUP influence — every task links back to one or more use case IDs, maintaining traceability from requirements → tasks → code.

### Repo-Level Artifacts

```
go_projects/
├── spec-templates/
│   ├── requirements.md    # Copy-and-fill template
│   ├── design.md          # Copy-and-fill template
│   └── tasks.md           # Copy-and-fill template
├── SPEC-GUIDE.md          # Workflow guide: how to use spec-driven development
├── awsct/docs/            # Project-level specs
├── gcpat/docs/            # Project-level specs
├── goto/docs/             # Project-level specs
├── goto-db/docs/          # Project-level specs
├── go-s3-downloader/docs/ # Project-level specs
├── ip-finder/docs/        # Project-level specs
└── ...
```

### SPEC-GUIDE.md Contents

A short workflow guide covering:

1. **Starting a new project** — copy templates, fill in requirements first, then design, then tasks
2. **Working on an existing project** — read specs before changing code; update specs when behavior changes
3. **Spec-first with AI agents** — point the agent at `requirements.md` + `design.md` before asking it to generate code
4. **Keeping specs in sync** — when code changes, check if specs need updating; the changelog in `tasks.md` tracks evolution
5. **Splitting a project to its own repo** — the `docs/` directory travels with the project; it is fully self-contained

### Migration Plan

1. **Move existing `awsct` spec** from `docs/superpowers/specs/2026-04-17-awsct-design.md` → split into `awsct/docs/requirements.md`, `awsct/docs/design.md`, `awsct/docs/tasks.md`
2. **Move existing `awsct` plan** from `docs/superpowers/plans/2026-04-17-awsct.md` → fold into `awsct/docs/tasks.md`
3. **Create specs for existing projects** by examining their code:
   - `gcpat` — GCP Audit Trail CLI
   - `goto` — AWS SSM/SSH connection tool
   - `goto-db` — Database connection tool
   - `go-s3-downloader` — S3 file downloader
   - `ip-finder` — IP lookup tool
4. **Create repo-level templates** in `spec-templates/`
5. **Create `SPEC-GUIDE.md`** at repo root

### What We Borrow from AIUP

| AIUP Element | How We Use It |
|-------------|---------------|
| Structured use cases (actors, flows, rules) | Core format of `requirements.md` |
| Use case IDs (UC-XXX-NNN) | Traceability from requirements → tasks → code |
| Business rules (BR-NNN) | Explicit, testable rules in each use case |
| Acceptance criteria (GIVEN/WHEN/THEN) | Unambiguous, testable success definitions |
| Traceability | Every task references the use cases it implements |

### What We Skip from AIUP

| AIUP Element | Why We Skip It |
|-------------|----------------|
| 4-phase ceremony (Inception/Elaboration/Construction/Transition) | Too heavy for CLI utility tools |
| Separate entity model diagrams | Folded into `design.md` when needed |
| Use case diagrams (UML) | Structured markdown is sufficient |
| Formal stakeholder alignment phase | Single-developer projects |

## Scalability

- **File-level:** Each of the three files grows independently
- **Project-level:** New spec files can be added alongside the three core files
- **Template-level:** Templates can evolve; existing project specs are not forced to match newer template versions
- **Cross-repo:** When a project moves to its own repo, specs travel with it unchanged
