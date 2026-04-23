# goto — Design

**Date:** 2025-04-23
**Status:** Active

## Architecture Overview

Single-file Go CLI that chains four stages: argument parsing, credential validation, EC2 instance lookup, and SSH connection.

```
CLI Args ──▶ parseArgs ──▶ areAWSCredentialsValid ──┬──▶ getEC2Instances ──▶ selectInstance ──▶ sshToInstance
                                                     │
                                                     └── (invalid) ──▶ aws-vault exec re-exec ──▶ exit
```

## Project Structure

```
goto/
├── main.go          # Entire implementation (all functions + main)
├── main_test.go     # Unit tests for parseArgs, selectInstance, extractFQDN
├── go.mod           # Module definition and dependencies
├── go.sum           # Dependency checksums
├── README.md        # Usage documentation
└── specs/           # Requirements, design, and task specifications
```

## Components

### Component: Argument Parser (`parseArgs`)

- **Responsibility:** Extracts role, env, and optional region from CLI arguments. Defaults region to `us-east-1`.
- **Interfaces:** Called with `os.Args`; returns `(role, env, region, error)`
- **Key Decisions:** Positional args instead of flags keeps the CLI minimal and fast to type.

### Component: Credential Checker (`areAWSCredentialsValid`)

- **Responsibility:** Validates current AWS credentials via STS GetCallerIdentity.
- **Interfaces:** Returns `bool`; no parameters (uses default AWS config chain).
- **Key Decisions:** Uses a separate STS call rather than attempting EC2 and catching auth errors, providing a clear point to trigger re-exec.

### Component: EC2 Instance Fetcher (`getEC2Instances`)

- **Responsibility:** Queries EC2 DescribeInstances filtered by `tag:Role` and `instance-state-name=running`.
- **Interfaces:** `(role, region) → ([]Instance, error)`
- **Key Decisions:** Filters are applied server-side via the EC2 API to minimize data transfer.

### Component: Instance Selector (`selectInstance`)

- **Responsibility:** If one instance, returns it. If multiple, displays a numbered list (FQDN + Instance ID) and prompts for user input via stdin.
- **Interfaces:** `([]Instance) → Instance`
- **Key Decisions:** Uses `log.Fatalf` on invalid selection for simplicity in a CLI context.

### Component: SSH Connector (`sshToInstance` + `extractFQDN`)

- **Responsibility:** Extracts the `FQDN` tag from the selected instance and execs `ssh <fqdn>` with stdin/stdout/stderr attached.
- **Interfaces:** Takes an `Instance`; does not return (execs into SSH or fatals).
- **Key Decisions:** Delegates to the system `ssh` binary rather than embedding an SSH client, keeping dependencies minimal.

## Data Flow

1. CLI arguments provide role, env, and optional region
2. STS GetCallerIdentity validates credentials; on failure, re-exec via `aws-vault exec`
3. EC2 DescribeInstances returns running instances matching the role tag
4. User selects an instance (or auto-selected if single match)
5. FQDN tag is extracted and passed to `ssh` for connection

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| aws-sdk-go-v2 | AWS SDK core | v1.36.3 |
| aws-sdk-go-v2/config | AWS config loading | v1.29.14 |
| aws-sdk-go-v2/service/ec2 | EC2 DescribeInstances API | v1.204.0 |
| aws-sdk-go-v2/service/sts | STS GetCallerIdentity for credential validation | v1.33.19 |
| aws-vault (external) | Credential management and session re-exec | — |
| ssh (external) | SSH client for instance connection | — |

## Authentication & Security

Credentials are managed entirely by `aws-vault`. The `env` argument doubles as the aws-vault profile name. When credentials are missing or expired, `goto` re-execs itself via `aws-vault exec <profile> -- goto <original-args>`, which injects temporary credentials as environment variables. No credentials are stored, logged, or passed through `goto` itself.

## Error Handling Strategy

- Invalid arguments → print usage, exit 1
- AWS credential failure → re-exec via aws-vault (transparent to user)
- EC2 API failure → `log.Fatalf` with descriptive error
- No matching instances → `log.Fatalf` with role and region context
- Missing FQDN tag → `log.Fatal` indicating the tag is missing
- Invalid instance selection → `log.Fatalf` with "Invalid selection"
- SSH connection failure → `log.Fatalf` with error from ssh process

## Future Considerations

- Support for SSM Session Manager as an alternative to direct SSH
- Tag-based filtering beyond Role (e.g., Name, Team)
- Output formatting options (JSON, table) for scripting
