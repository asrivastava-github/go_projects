# go-s3-downloader — Design

**Date:** 2026-04-23
**Status:** Active

## Architecture Overview

Single Go binary with a linear pipeline: parse flags → configure AWS → authenticate → list S3 objects → download files.

```
main.go (flags, orchestration)
    │
    ├──▶  config/config.go      (environment → region/profile mapping)
    │
    ├──▶  awsauth/auth.go       (STS validation, aws-vault re-exec)
    │
    ├──▶  s3client/client.go    (AWS SDK v2 S3 client init, single-file download)
    │
    └──▶  handler/download.go   (ListObjectsV2 pagination, batch download loop)
                │
                └──▶  AWS S3 API
```

## Project Structure

```
go-s3-downloader/
├── main.go                       # Entry point: flag parsing, orchestration
├── internal/
│   ├── config/
│   │   └── config.go             # Config struct, env-based defaults (region, profile)
│   ├── awsauth/
│   │   └── auth.go               # STS credential validation, aws-vault re-exec
│   ├── s3client/
│   │   └── client.go             # S3 client wrapper (NewS3Client, DownloadFile)
│   └── handler/
│       └── download.go           # DownloadFiles: ListObjectsV2 pagination, batch download
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Components

### Component: main.go (Entry Point)

- **Responsibility:** Parse command-line flags (`-env`, `-profile`, `-bucketPrefix`), validate inputs, split bucket/prefix, orchestrate the pipeline (config → auth → S3 client → download).
- **Interfaces:** Calls `config.NewConfig()`, `awsauth.RenewCredentials()`, `s3client.NewS3Client()`, `handler.DownloadFiles()`.
- **Key Decisions:** Uses `flag` package directly (no cobra) — appropriate for a simple CLI with no subcommands.

### Component: config/config.go (Configuration)

- **Responsibility:** Map environment string to AWS region and profile. Provide config validation.
- **Interfaces:** `NewConfig(env, profile) → *Config`, `LoadConfig(filePath) → *Config`, `Config.Validate() → error`.
- **Key Decisions:** `prod` maps to `us-east-1`, all other environments default to `eu-west-1`. Profile defaults to the env name unless explicitly overridden. Also supports JSON config file loading via `LoadConfig`.

### Component: awsauth/auth.go (AWS Authentication)

- **Responsibility:** Validate AWS credentials via `sts:GetCallerIdentity`. If invalid, re-exec the current process under `aws-vault exec <profile>`.
- **Interfaces:** `AreAWSCredentialsValid() → bool`, `RenewCredentials(profile) → bool`.
- **Key Decisions:** Re-exec pattern passes all original `os.Args` to aws-vault so the program seamlessly restarts with valid credentials. Returns `true` when aws-vault was invoked (caller should `os.Exit(0)`).

### Component: s3client/client.go (S3 Client Wrapper)

- **Responsibility:** Initialize an AWS S3 client from the config region. Provide a single-file download method.
- **Interfaces:** `NewS3Client(cfg) → (*S3Client, error)`, `S3Client.GetS3Client() → *s3.Client`, `S3Client.DownloadFile(bucket, key) → error`.
- **Key Decisions:** Wraps `aws-sdk-go-v2/service/s3`. The `DownloadFile` method on `S3Client` is available but the main download path uses `handler.DownloadFiles` which works with the raw `*s3.Client` directly.

### Component: handler/download.go (Download Logic)

- **Responsibility:** List all objects under a bucket/prefix using `ListObjectsV2` with pagination, then download each file to `~/s3_files/<key>`.
- **Interfaces:** `DownloadFiles(s3Client, bucket, prefix) → error`.
- **Key Decisions:** Uses `s3.NewListObjectsV2Paginator` for automatic pagination. Skips directory placeholder objects (keys ending in `/`). Continues downloading on individual file failures (prints warning). Files are created with `0644` permissions.

## Data Flow

1. User provides `-env`, `-bucketPrefix`, and optional `-profile` via command-line flags
2. `main.go` parses flags and splits `bucketPrefix` into bucket name and prefix
3. `config.NewConfig` maps environment to region and profile
4. `awsauth.RenewCredentials` validates credentials via STS; re-execs under aws-vault if expired
5. `s3client.NewS3Client` creates an S3 client configured for the target region
6. `handler.DownloadFiles` paginates through `ListObjectsV2` results
7. Each object is downloaded via `GetObject` and written to `~/s3_files/<key>`
8. Success message printed to stdout

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| `github.com/aws/aws-sdk-go-v2` | AWS SDK v2 core | v1.36.3 |
| `github.com/aws/aws-sdk-go-v2/config` | AWS config loading | v1.29.14 |
| `github.com/aws/aws-sdk-go-v2/service/s3` | S3 API client | v1.79.3 |
| `github.com/aws/aws-sdk-go-v2/service/sts` | STS for credential validation | v1.33.19 |

## Authentication & Security

Uses the **AWS SDK v2 default credential chain** (env vars, shared config, IAM role, etc.).

1. `awsauth.AreAWSCredentialsValid()` calls `sts:GetCallerIdentity` to check credentials
2. If invalid/expired, `RenewCredentials()` re-execs the process under `aws-vault exec <profile> -- go-s3-downloader <original-args>`
3. aws-vault handles MFA/SSO prompts and injects session credentials as environment variables
4. The restarted process finds valid credentials and proceeds normally

## Error Handling Strategy

- Missing required flags (`-env`, `-bucketPrefix`) → `log.Fatalf`, exit 1
- Invalid bucket/prefix format → `log.Fatalf`, exit 1
- Invalid configuration → `log.Fatalf`, exit 1
- Expired AWS credentials → re-exec under aws-vault, exit 0
- S3 client creation failure → `log.Fatalf`, exit 1
- ListObjectsV2 pagination error → return error, `log.Fatalf`, exit 1
- Individual file download failure → print warning, continue with remaining files
- Directory creation failure → return error for that file, continue

## Future Considerations

- Add concurrent/parallel downloads for better performance on large file sets
- Add progress bar or download statistics
- Add `-output` flag to customize the local destination directory
- Add file size filtering or pattern matching (glob) for selective downloads
- Add dry-run mode to preview files without downloading
