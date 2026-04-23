# go-s3-downloader — Requirements

**Date:** 2026-04-23
**Status:** Active
**Project:** go_projects/go-s3-downloader

## Overview

**Purpose:** A Go CLI tool that downloads files from Amazon S3 buckets, supporting multiple environments (prod/dev) with automatic credential management via aws-vault.

**Problem Statement:** Downloading files from S3 requires navigating the AWS console or writing ad-hoc scripts. go-s3-downloader provides a single command to bulk-download all objects under a bucket/prefix, preserving the S3 folder structure locally.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | CLI user who runs go-s3-downloader to download files from S3 buckets |

## Use Cases

### UC-DL-001: Download files from S3 bucket/prefix

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials available; target S3 bucket exists and is accessible
- **Main Flow:**
  1. Developer runs `go-s3-downloader -env=<env> -bucketPrefix=<bucket/prefix>`
  2. System validates required flags (`-env`, `-bucketPrefix`)
  3. System parses bucket name and prefix from the `-bucketPrefix` flag
  4. System loads configuration for the specified environment (region, profile)
  5. System checks AWS credentials via STS and renews via aws-vault if expired
  6. System initializes S3 client with environment-specific region
  7. System lists all objects under the specified prefix using `ListObjectsV2` with pagination
  8. System downloads each object to `~/s3_files/<key>`, preserving folder structure
  9. System prints "Files downloaded successfully."
- **Alternative Flows:**
  - 2a. Missing `-env` flag → print error, exit 1
  - 2b. Missing `-bucketPrefix` flag → print error, exit 1
  - 5a. Credentials expired → re-exec under `aws-vault exec <profile>`, exit 0 (aws-vault restarts the program)
  - 7a. S3 API error → print error, exit 1
  - 8a. Individual file download fails → print warning, continue downloading remaining files
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005
- **Acceptance Criteria:**
  - GIVEN valid AWS credentials and objects exist under `my-bucket/data/`
  - WHEN the developer runs `go-s3-downloader -env=prod -bucketPrefix=my-bucket/data`
  - THEN all objects under `data/` are downloaded to `~/s3_files/data/` preserving the S3 folder structure

### UC-DL-002: Select environment (prod/dev)

- **Actor:** Developer
- **Preconditions:** None
- **Main Flow:**
  1. Developer specifies `-env=prod` or `-env=dev`
  2. System creates configuration with environment-specific defaults:
     - `prod` → region `us-east-1`
     - `dev` (or any other value) → region `eu-west-1`
  3. System sets AWS profile to the environment name (unless `-profile` overrides it)
- **Business Rules:**
  - BR-001, BR-002, BR-006
- **Acceptance Criteria:**
  - GIVEN the developer runs with `-env=prod`
  - WHEN the S3 client is initialized
  - THEN the AWS region is `us-east-1` and the profile is `prod`

### UC-DL-003: Override AWS profile

- **Actor:** Developer
- **Preconditions:** None
- **Main Flow:**
  1. Developer specifies `-profile=<custom-profile>` alongside `-env`
  2. System uses the custom profile for AWS credential resolution instead of the environment name
- **Business Rules:**
  - BR-006
- **Acceptance Criteria:**
  - GIVEN the developer runs with `-env=prod -profile=my-custom-profile`
  - WHEN credentials are checked/renewed
  - THEN the profile `my-custom-profile` is used for aws-vault

## Business Rules

| ID | Rule |
|----|------|
| BR-001 | Environment determines the AWS region: `prod` → `us-east-1`, other → `eu-west-1` |
| BR-002 | Both `-env` and `-bucketPrefix` flags are required |
| BR-003 | Files are saved to `~/s3_files/` preserving the full S3 key as the relative path |
| BR-004 | Uses AWS SDK v2 with `ListObjectsV2` pagination to handle large numbers of objects |
| BR-005 | Individual file download failures are logged as warnings; remaining files continue downloading |
| BR-006 | AWS profile defaults to the `-env` value; `-profile` flag overrides it |

## Non-Functional Requirements

- **Security:** Authentication via AWS SDK v2 default credential chain. Credentials are validated via `sts:GetCallerIdentity`; if expired, the process re-execs under `aws-vault exec <profile>`.
- **Portability:** Single Go binary, primary target macOS (arm64). Installed to `/usr/local/bin/` via `make build`.

## Constraints & Limitations

- S3 `ListObjectsV2` returns max 1000 objects per page (auto-paginated)
- Directory placeholder objects (keys ending in `/`) are skipped
- No progress bar or download size reporting
- No concurrent/parallel downloads — files are downloaded sequentially
- aws-vault re-exec causes the process to exit and restart, which is expected behavior
