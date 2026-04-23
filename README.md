# go_projects

A collection of Go CLI tools for day-to-day cloud operations.

## Projects

| Project | Description |
|---------|-------------|
| [awsct](awsct/) | AWS CloudTrail event finder (CLI + MCP server) |
| [gcpat](gcpat/) | GCP Cloud Audit Log event finder (CLI + MCP server) |
| [goto](goto/) | SSH into EC2 instances by Role tag |
| [goto-db](goto-db/) | SSH tunnel to databases via jump host + DbGate UI |
| [go-s3-downloader](go-s3-downloader/) | Download files from S3 buckets |
| [ip-finder](ip-finder/) | Identify AWS resources (EC2, Lambda, EKS pods) by IP address |

## Spec-Driven Development

This repo follows a spec-driven development workflow. Each project has a `specs/` directory with structured specifications that serve as the source of truth.

See **[SPEC-GUIDE.md](SPEC-GUIDE.md)** for the full workflow, and **[spec-templates/](spec-templates/)** for reusable templates when starting a new project.
