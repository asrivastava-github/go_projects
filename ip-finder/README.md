# IP Finder

A CLI tool to identify AWS resources (EC2, Lambda, EKS pods) by internal IP address.

## Architecture

```
ip-finder/
├── cmd/                    # CLI commands (Cobra)
│   └── root.go
├── pkg/
│   ├── aws/                # AWS SDK interactions
│   │   ├── client.go       # AWS client factory
│   │   ├── classifier.go   # Resource type classification (Lambda, ELB, etc.)
│   │   ├── credentials.go  # Credential validation (STS GetCallerIdentity)
│   │   ├── eni.go          # ENI lookup
│   │   └── instance.go     # EC2 instance details
│   ├── k8s/                # Kubernetes interactions
│   │   ├── client.go       # K8s client factory
│   │   └── pods.go         # Pod lookup
│   ├── finder/             # Core business logic
│   │   └── finder.go       # Orchestrates AWS + K8s searches
│   ├── logger/             # Logging utilities
│   │   └── logger.go       # Leveled logger (DEBUG, INFO, WARN, ERROR)
│   └── output/             # Output formatting
│       └── printer.go      # Tabular output
├── main.go
├── go.mod
└── Makefile
```

## How It Works

1. **Credential Check**: Validates AWS credentials via STS GetCallerIdentity
   - If expired and `aws-vault` is installed → auto re-runs with `aws-vault exec`
   - If expired and `aws-vault` not found → shows manual fix instructions
2. **ENI Search**: Queries `DescribeNetworkInterfaces` for the IP
   - First checks secondary IPs on ENIs
   - If not found, searches prefix delegations (EKS VPC CNI `/28` blocks)
   - If still not found → stops and skips K8s search
3. **Classification**: Identifies resource type (EC2, Lambda, ELB, NAT Gateway, etc.)
4. **Instance Lookup**: If ENI is attached to EC2, fetches instance details
5. **Smart K8s Search**: Only queries Kubernetes API if the IP could belong to a pod
   - Skips K8s search for Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoints
   - Searches K8s for EC2 instances (potential EKS nodes)

## Installation

```bash
cd code-snippets/aws/ip-finder

# Download dependencies
make tidy

# Build binary
make build

# Or install to $GOPATH/bin
make install
```

## Usage

```bash
# Prod defaults: just pass the IP (uses -p prod -r us-east-1 -k produe102)
./bin/ip-finder 100.64.5.42

# Override for different environment
./bin/ip-finder 10.0.1.50 -p nonprod -r us-west-2 -k nonprodue102

# AWS-only search (skip K8s)
./bin/ip-finder 10.0.1.50 --skip-k8s

# Search all K8s contexts
./bin/ip-finder 10.0.1.50 --all-contexts
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `prod` | AWS profile from ~/.aws/credentials |
| `--region` | `-r` | `us-east-1` | AWS region |
| `--kube-context` | `-k` | `produe102` | Kubernetes context |
| `--all-contexts` | `-a` | `false` | Search all available K8s contexts |
| `--skip-k8s` | | `false` | Skip Kubernetes pod search |
| `--debug` | `-d` | `false` | Enable debug logging |

## Read-Only Operations

This tool performs **only read operations**:

| Service | API Call | Permission Required |
|---------|----------|---------------------|
| EC2 | `DescribeNetworkInterfaces` | `ec2:DescribeNetworkInterfaces` |
| EC2 | `DescribeInstances` | `ec2:DescribeInstances` |
| K8s | `list pods` | `pods` list in RBAC |
| K8s | `list nodes` | `nodes` list in RBAC |

## Requirements

- Go 1.23+
- AWS credentials configured (`~/.aws/credentials` or environment)
- Kubernetes config (`~/.kube/config`) for pod search
- IAM permissions: `ec2:DescribeNetworkInterfaces`, `ec2:DescribeInstances`, `sts:GetCallerIdentity`
- (Optional) `aws-vault` for automatic credential refresh, otherwise use `aws sso login`

## What Gets Identified

| Resource Type | How It's Found |
|---------------|----------------|
| EC2 Instance | ENI primary/secondary IP → Instance attachment |
| Lambda | ENI with description containing "AWS Lambda" |
| EKS Pod | Direct pod IP match via K8s API |
| NAT Gateway | ENI type = "nat_gateway" |
| ELB/ALB | ENI type = "interface" with ELB description |

## Example Output

### EKS Pod IP (from secondary CIDR)
```
[Authenticated] arn:aws:sts::123456789:assumed-role/Admin/user@example.com (Account: 123456789)

[Search] IP: 100.64.5.42 | Region: us-east-1 | Profile: prod

━━━ AWS ENI Search ━━━
ENI ID              Resource Type   Status   Instance ID          Description
------              -------------   ------   -----------          -----------
eni-0abc123def456   EC2 Instance    in-use   i-0123456789abcdef0  aws-K8S-i-0123456789abcdef0

  VPC: vpc-12345678 | Subnet: subnet-abcd1234 | AZ: us-east-1a
  All Private IPs: 10.0.5.10, 100.64.5.42, 100.64.5.43, 100.64.5.44

━━━ EC2 Instance Details ━━━
Instance ID:  i-0123456789abcdef0
Name:         eks-node-group-abc123
Type:         m5.xlarge
State:        running
Private IP:   10.0.5.10

━━━ Kubernetes Pod Search ━━━

[Found] Context: prod-eks-cluster
NAMESPACE    NAME                      POD IP        NODE                  STATUS
---------    ----                      ------        ----                  ------
app          my-service-7d8f9b6c5-x2k  100.64.5.42   ip-10-0-5-10.ec2...   Running
```

### Lambda Function IP (K8s search skipped automatically)
```
[Authenticated] arn:aws:sts::123456789:assumed-role/Admin/user@example.com (Account: 123456789)

[Search] IP: 10.0.8.55 | Region: us-east-1 | Profile: prod

━━━ AWS ENI Search ━━━
ENI ID              Resource Type      Status   Instance ID   Description
------              -------------      ------   -----------   -----------
eni-0def456abc789   Lambda Function    in-use   -             AWS Lambda VPC ENI-my-function

  VPC: vpc-12345678 | Subnet: subnet-efgh5678 | AZ: us-east-1b

━━━ Kubernetes Pod Search ━━━
[Skipped] IP belongs to Lambda Function, not an EKS pod.
```

## Limitations

| Scope | Limitation |
|-------|------------|
| **AWS Account** | Searches only the account associated with the specified profile (`-p`) |
| **AWS Region** | Searches only the specified region (`-r`); IPs in other regions won't be found |
| **K8s Cluster** | Searches only the specified context (`-k`); pods in other clusters won't be found |
| **VPC CNI** | Supports both secondary IP and prefix delegation modes |
| **IP Not Found** | If IP not found in any ENI (secondary or prefix), K8s search is skipped |

**Search Flow:**
1. IP not in ENI (any mode) → **Stops here**, suggests checking region/account
2. IP in ENI but non-EC2 (Lambda, ELB, etc.) → **Stops here**, no K8s search
3. IP in ENI attached to EC2 → Searches K8s cluster for pod

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```
