# goto — Tasks

**Date:** 2025-04-23
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Argument Parsing

**Traces to:** UC-GOTO-001
**Files:** `main.go`

- [x] Step 1: Implement `parseArgs` to extract role, env, and optional region from CLI args
- [x] Step 2: Default region to `us-east-1` when not provided
- [x] Step 3: Return error on insufficient arguments with usage message
- [x] Step 4: Unit tests for parseArgs (role+env, role+env+region, too few args)

## Task 2: Credential Validation and aws-vault Re-exec

**Traces to:** UC-GOTO-001
**Files:** `main.go`

- [x] Step 1: Implement `areAWSCredentialsValid` using STS GetCallerIdentity
- [x] Step 2: On invalid credentials, re-exec via `aws-vault exec <env> -- goto <args>`
- [x] Step 3: Attach stdin/stdout/stderr to re-exec subprocess
- [x] Step 4: Exit after re-exec to avoid duplicate execution

## Task 3: EC2 Instance Fetching

**Traces to:** UC-GOTO-001
**Files:** `main.go`

- [x] Step 1: Implement `getEC2Instances` with EC2 DescribeInstances API
- [x] Step 2: Apply server-side filters: `tag:Role` and `instance-state-name=running`
- [x] Step 3: Flatten reservations into a single instance slice
- [x] Step 4: Handle zero results with log message and return empty slice

## Task 4: Instance Selection

**Traces to:** UC-GOTO-001, UC-GOTO-002
**Files:** `main.go`

- [x] Step 1: Implement `selectInstance` — auto-return for single instance
- [x] Step 2: Display numbered list with FQDN and Instance ID for multiple instances
- [x] Step 3: Prompt user for selection via stdin
- [x] Step 4: Validate selection bounds, fatal on invalid input
- [x] Step 5: Unit test for single-instance auto-selection

## Task 5: SSH Connection

**Traces to:** UC-GOTO-001
**Files:** `main.go`

- [x] Step 1: Implement `extractFQDN` to pull FQDN tag value from instance
- [x] Step 2: Implement `sshToInstance` to exec `ssh <fqdn>` with stdin/stdout/stderr
- [x] Step 3: Fatal exit when FQDN tag is missing
- [x] Step 4: Unit tests for extractFQDN (tag present, missing, no tags)

## Task 6: Main Orchestration

**Traces to:** UC-GOTO-001, UC-GOTO-002
**Files:** `main.go`

- [x] Step 1: Wire parseArgs → credential check → getEC2Instances → selectInstance → sshToInstance
- [x] Step 2: Fatal exit when no instances found for role
- [x] Step 3: Log role, region, and env before EC2 query

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2025-04-23 | Initial spec created — all tasks complete (existing implementation) | AS |
