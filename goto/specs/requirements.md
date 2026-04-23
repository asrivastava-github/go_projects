# goto — Requirements

**Date:** 2025-04-23
**Status:** Active
**Project:** goto

## Overview

**Purpose:** CLI tool that fetches EC2 instances by Role tag and SSHs into them via their FQDN tag, using aws-vault for credential management.

**Problem Statement:** Developers need a fast way to SSH into EC2 instances by role and environment without manually looking up instance hostnames in the AWS console. `goto` automates the lookup-and-connect workflow.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | Runs `goto` from a terminal to SSH into EC2 instances by role and environment |

## Use Cases

### UC-GOTO-001: SSH into an EC2 instance by role and environment

- **Actor:** Developer
- **Preconditions:** `aws-vault` is installed and configured with a profile matching the target environment. The target EC2 instance has `Role` and `FQDN` tags and is in a running state.
- **Main Flow:**
  1. Developer runs `goto <role> <env> [aws_region]`
  2. System parses arguments; region defaults to `us-east-1` if omitted
  3. System checks AWS credentials via STS GetCallerIdentity
  4. System queries EC2 for running instances matching `tag:Role = <role>`
  5. System selects the single matching instance
  6. System extracts the `FQDN` tag value and executes `ssh <fqdn>`
- **Alternative Flows:**
  - 3a. Credentials are missing or expired → system re-execs itself via `aws-vault exec <env> -- goto <args>` and exits
  - 4a. No instances match → system exits with error: "No instances found for role"
  - 5a. Multiple instances match → triggers UC-GOTO-002
  - 6a. No `FQDN` tag on selected instance → system exits with error: "No FQDN tag found"
- **Business Rules:**
  - BR-001: Only running instances are returned from EC2 queries
  - BR-002: If no FQDN tag is found on the selected instance, exit with error
  - BR-003: If no instances are found for the given role, exit with error
  - BR-004: Region defaults to `us-east-1` when not specified
  - BR-005: The `env` argument is used as the aws-vault profile name
- **Acceptance Criteria:**
  - GIVEN valid credentials and a single running instance with Role=webserver and FQDN=web1.example.com
  - WHEN the developer runs `goto webserver prod`
  - THEN the system SSHs into web1.example.com

### UC-GOTO-002: Select from multiple matching instances

- **Actor:** Developer
- **Preconditions:** Multiple running EC2 instances match the given role tag
- **Main Flow:**
  1. System displays a numbered list of matching instances showing FQDN and Instance ID
  2. Developer enters the number of the desired instance
  3. System connects via SSH to the selected instance's FQDN
- **Alternative Flows:**
  - 2a. Developer enters an invalid selection → system exits with "Invalid selection"
- **Business Rules:**
  - BR-001: Only running instances are returned
  - BR-002: If no FQDN tag is found on the selected instance, exit with error
- **Acceptance Criteria:**
  - GIVEN valid credentials and 3 running instances with Role=api
  - WHEN the developer runs `goto api prod` and selects instance 2
  - THEN the system SSHs into the FQDN of the second listed instance

## Non-Functional Requirements

- **Security:** Credentials are never stored or logged by `goto`; all credential management is delegated to `aws-vault`. The re-exec pattern avoids holding expired credentials in-process.
- **Portability:** macOS and Linux (requires `ssh` and `aws-vault` on PATH)

## Constraints & Limitations

- Requires `aws-vault` to be installed and configured
- Requires `ssh` binary on PATH
- EC2 instances must have `Role` and `FQDN` tags to be usable
- Single-file CLI; no configuration file support
