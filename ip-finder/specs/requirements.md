# ip-finder — Requirements

**Date:** 2026-04-23
**Status:** Active
**Project:** go_projects/ip-finder

## Overview

**Purpose:** A Go CLI tool that identifies AWS resources (EC2, Lambda, EKS pods) by internal IP address.

**Problem Statement:** Investigating which AWS resource owns a private IP requires navigating multiple AWS consoles (EC2, Lambda, VPC) and Kubernetes dashboards. ip-finder answers the question "what owns this IP?" in a single command by searching ENIs, classifying the resource type, and optionally searching Kubernetes clusters for matching pods.

## Actors

| Actor | Description |
|-------|-------------|
| Developer | CLI user who runs ip-finder with an IP address to identify the owning AWS resource and/or Kubernetes pod |

## Use Cases

### UC-IP-001: Identify AWS resource by IP

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials available; target region and account configured
- **Main Flow:**
  1. Developer runs `ip-finder <ip-address>` with optional flags
  2. System validates the IP address format
  3. System validates AWS credentials via STS GetCallerIdentity
  4. System searches ENIs by secondary IP address (`DescribeNetworkInterfaces` with `addresses.private-ip-address` filter)
  5. If no ENI found via secondary IP, system searches prefix delegations (EKS VPC CNI `/28` blocks) by listing in-use ENIs and checking IPv4 prefix CIDR containment
  6. System classifies each ENI by resource type (EC2, Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoint, Transit Gateway, EKS Control Plane)
  7. If ENI is attached to an EC2 instance, system fetches instance details via `DescribeInstances`
  8. If resource type may be a pod IP (EC2 or unknown), system searches Kubernetes cluster(s) for pods matching the IP
  9. System displays results in tabular format (ENI details, EC2 instance details, K8s pod details)
- **Alternative Flows:**
  - 2a. Invalid IP format → print error, exit 1
  - 3a. Credentials expired and `aws-vault` installed → re-exec under `aws-vault exec <profile>` automatically
  - 3b. Credentials expired and `aws-vault` not installed → print manual fix instructions, exit 1
  - 3c. Already running under `aws-vault` and credentials fail → print auth error, exit 1
  - 4a–5a. No ENI found in any mode → print "No ENI found" message, skip K8s search, suggest checking region/account
  - 8a. K8s connection fails for a context → log warning, continue with other contexts
  - 8b. Pod IP matches only hostNetwork pods → also search for application (non-hostNetwork) pods on the same node
- **Business Rules:**
  - BR-001, BR-002, BR-003, BR-004, BR-005, BR-006
- **Acceptance Criteria:**
  - GIVEN valid AWS credentials and an EC2-attached ENI with secondary IP `100.64.5.42`
  - WHEN the developer runs `ip-finder 100.64.5.42`
  - THEN the tool displays the ENI details, EC2 instance details, and matching K8s pod (if found)

### UC-IP-002: Smart skip K8s for non-EC2 resources

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials; ENI found for the IP
- **Main Flow:**
  1. System classifies the ENI resource type
  2. System checks `MayBePodIP` flag on the classified ENI
  3. If no ENI has `MayBePodIP=true` (Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoint, Transit Gateway, EKS Control Plane), system skips K8s search
  4. System displays skip reason (e.g., "IP belongs to Lambda Function, not an EKS pod.")
- **Alternative Flows:**
  - 2a. ENI is EC2-attached or unknown type → `MayBePodIP=true`, proceed with K8s search
- **Business Rules:**
  - BR-002
- **Acceptance Criteria:**
  - GIVEN a Lambda function ENI for IP `10.0.8.55`
  - WHEN the developer runs `ip-finder 10.0.8.55`
  - THEN the tool displays ENI details and "[Skipped] IP belongs to Lambda Function, not an EKS pod."

### UC-IP-003: Search all K8s contexts

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials; ENI found with `MayBePodIP=true`; multiple K8s contexts in kubeconfig
- **Main Flow:**
  1. Developer runs `ip-finder <ip> --all-contexts`
  2. System reads all available contexts from kubeconfig
  3. System searches each context for pods matching the IP
  4. System displays results from all contexts (including errors for unreachable contexts)
- **Alternative Flows:**
  - 2a. Failed to read kubeconfig → fall back to default context
  - 3a. Context connection fails → log error, continue with remaining contexts
- **Business Rules:**
  - BR-003, BR-006
- **Acceptance Criteria:**
  - GIVEN 3 K8s contexts in kubeconfig and a pod exists in context `produe102`
  - WHEN the developer runs `ip-finder 100.64.5.42 --all-contexts`
  - THEN the tool searches all 3 contexts and displays the pod found in `produe102`

### UC-IP-004: AWS-only search

- **Actor:** Developer
- **Preconditions:** Valid AWS credentials
- **Main Flow:**
  1. Developer runs `ip-finder <ip> --skip-k8s`
  2. System performs ENI search and classification
  3. System skips K8s pod search entirely
  4. System displays ENI and EC2 details only, with skip reason "Skipped by user (--skip-k8s flag)."
- **Business Rules:**
  - BR-001
- **Acceptance Criteria:**
  - GIVEN an EC2-attached ENI for IP `10.0.1.50`
  - WHEN the developer runs `ip-finder 10.0.1.50 --skip-k8s`
  - THEN the tool displays ENI and EC2 details without K8s pod search

## Business Rules

| ID | Rule |
|----|------|
| BR-001 | All operations are read-only (`DescribeNetworkInterfaces`, `DescribeInstances`, `GetCallerIdentity`, K8s list pods/nodes) |
| BR-002 | Smart K8s skip: K8s search is skipped for resource types that cannot be EKS pods (Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoint, Transit Gateway, EKS Control Plane) |
| BR-003 | Supports both secondary IP and prefix delegation (EKS VPC CNI `/28` blocks) for ENI discovery |
| BR-004 | aws-vault re-exec: if credentials are expired and `aws-vault` is installed, the tool re-execs itself under `aws-vault exec <profile>` automatically |
| BR-005 | Single account/region per query — results only cover the AWS account associated with the profile and the specified region |
| BR-006 | When running under aws-vault (`AWS_VAULT` env set), the tool clears the `AWS_VAULT` env var before creating K8s clients to avoid nested aws-vault conflicts |

## CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `prod` | AWS profile from `~/.aws/credentials` |
| `--region` | `-r` | `us-east-1` | AWS region |
| `--kube-context` | `-k` | `produe102` | Kubernetes context |
| `--all-contexts` | `-a` | `false` | Search all available K8s contexts |
| `--skip-k8s` | | `false` | Skip Kubernetes pod search |
| `--debug` | `-d` | `false` | Enable debug logging |

## Non-Functional Requirements

- **Security:** Authentication via AWS SDK v2 default credential chain. Supports aws-vault re-exec for credential refresh. All operations are strictly read-only.
- **Performance:** ENI prefix delegation search lists all in-use ENIs when secondary IP search fails — scope is bounded by single region/account.
- **Portability:** Single Go binary, cross-compilable. Primary target: macOS (arm64).

## Constraints & Limitations

- Searches only the AWS account associated with the specified profile (`-p`)
- Searches only the specified AWS region (`-r`); IPs in other regions won't be found
- K8s search uses only the specified context (`-k`) unless `--all-contexts` is set
- Prefix delegation search lists all in-use ENIs in the region — may be slow in accounts with many ENIs
- If IP is not found in any ENI, K8s search is skipped (assumes IP doesn't exist in this account/region)
- hostNetwork pod detection: if all matched pods use hostNetwork, tool additionally searches for application pods on the same node
