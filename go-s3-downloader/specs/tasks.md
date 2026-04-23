# go-s3-downloader — Tasks

**Date:** 2026-04-23
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Project Scaffold

**Traces to:** UC-DL-001, UC-DL-002
**Files:** `main.go`, `go.mod`, `Makefile`

- [x] Step 1: Create `go.mod` (`go mod init go-s3-downloader`)
- [x] Step 2: Create `main.go` with flag parsing (`-env`, `-profile`, `-bucketPrefix`)
- [x] Step 3: Create `Makefile` with build, run, clean, test, fmt targets
- [x] Step 4: Add AWS SDK v2 dependencies and verify build

## Task 2: Configuration

**Traces to:** UC-DL-002, UC-DL-003
**Files:** `internal/config/config.go`

- [x] Step 1: Create Config struct (Environment, Region, Bucket, Prefix, Profile)
- [x] Step 2: Implement `NewConfig(env, profile)` with environment-based region defaults (`prod` → `us-east-1`, other → `eu-west-1`)
- [x] Step 3: Implement `LoadConfig(filePath)` for JSON config file support
- [x] Step 4: Implement `Config.Validate()` — ensure environment is non-empty
- [x] Step 5: Verify compilation

## Task 3: AWS Authentication

**Traces to:** UC-DL-001
**Files:** `internal/awsauth/auth.go`

- [x] Step 1: Implement `AreAWSCredentialsValid()` — validate via `sts:GetCallerIdentity`
- [x] Step 2: Implement `RenewCredentials(profile)` — re-exec under `aws-vault exec <profile>` preserving original args
- [x] Step 3: Integrate credential check in `main.go` (exit 0 after aws-vault re-exec)
- [x] Step 4: Verify compilation

## Task 4: S3 Client Wrapper

**Traces to:** UC-DL-001
**Files:** `internal/s3client/client.go`

- [x] Step 1: Create S3Client struct wrapping `*s3.Client`
- [x] Step 2: Implement `NewS3Client(cfg)` — load AWS config with region, create S3 client
- [x] Step 3: Implement `GetS3Client()` accessor
- [x] Step 4: Implement `DownloadFile(bucket, key)` — download single file to `~/s3_files/<key>`
- [x] Step 5: Verify compilation

## Task 5: Download Handler

**Traces to:** UC-DL-001
**Files:** `internal/handler/download.go`

- [x] Step 1: Implement `DownloadFiles(s3Client, bucket, prefix)` — ListObjectsV2 with pagination
- [x] Step 2: Skip directory placeholder objects (keys ending in `/`)
- [x] Step 3: Implement `downloadFileWithPath()` — download individual file with directory creation
- [x] Step 4: Handle individual download failures with warning (continue remaining files)
- [x] Step 5: Verify compilation

## Task 6: Integration & Main Flow

**Traces to:** UC-DL-001, UC-DL-002, UC-DL-003
**Files:** `main.go`

- [x] Step 1: Wire up full pipeline: flags → config → auth → S3 client → download
- [x] Step 2: Add input validation (required flags, bucket/prefix format)
- [x] Step 3: Build and verify end-to-end with `make build`

## Task 7: README & Documentation

**Traces to:** UC-DL-001, UC-DL-002, UC-DL-003
**Files:** `README.md`

- [x] Step 1: Create README.md with overview, project structure, installation, usage, and examples
- [x] Step 2: Final build and smoke test

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-04-23 | Initial spec created; all tasks complete (retroactive documentation) | AS |
