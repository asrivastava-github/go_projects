# ip-finder — Design

**Date:** 2026-04-23
**Status:** Active

## Architecture Overview

Single Go binary with a cobra CLI entry point. The finder package orchestrates the search flow across AWS and Kubernetes packages, with results formatted by the output package.

```
CLI (cobra)  ──▶  finder (orchestrator)  ──┬──▶  aws pkg (eni, classifier, instance, credentials)  ──▶  AWS EC2/STS API
                                           └──▶  k8s pkg (pods, client)  ──▶  K8s API
                                                          │
                                                          ▼
                                                   output pkg (printer)  ──▶  stdout
```

## Project Structure

```
ip-finder/
├── main.go              # Entry point, calls cmd.Execute()
├── cmd/
│   └── root.go          # Cobra root command, flags, IP validation, aws-vault re-exec
├── pkg/
│   ├── aws/
│   │   ├── client.go    # AWS client factory (EC2 client from SDK config)
│   │   ├── credentials.go # Credential validation via STS GetCallerIdentity
│   │   ├── eni.go       # ENI lookup (secondary IP + prefix delegation)
│   │   ├── classifier.go # Resource type classification from ENI metadata
│   │   └── instance.go  # EC2 instance details via DescribeInstances
│   ├── k8s/
│   │   ├── client.go    # K8s client factory (kubeconfig loading, context selection)
│   │   └── pods.go      # Pod lookup by IP and by node IP
│   ├── finder/
│   │   └── finder.go    # Orchestrates AWS + K8s searches, smart skip logic
│   ├── logger/
│   │   └── logger.go    # Leveled logger (DEBUG, INFO, WARN, ERROR)
│   └── output/
│       └── printer.go   # Tabular output formatting
├── go.mod
├── go.sum
└── Makefile
```

## Components

### Component: CLI Layer (`cmd/root.go`)

- **Responsibility:** Parse user input via cobra (single root command with IP as positional arg), manage flags (`--profile`, `--region`, `--kube-context`, `--all-contexts`, `--skip-k8s`, `--debug`), validate IP format, handle aws-vault re-exec for credential refresh, and invoke the finder.
- **Interfaces:** Calls `aws.ValidateCredentials()`, creates `finder.IPFinder`, calls `finder.Find()`, passes result to `output.Printer`.
- **Key Decisions:** Single command (not subcommands) since the tool does one thing. aws-vault re-exec uses `os/exec` to re-run the entire binary under `aws-vault exec`. When running under aws-vault (`AWS_VAULT` env set), profile is set to empty to use environment credentials.

### Component: Credential Validator (`pkg/aws/credentials.go`)

- **Responsibility:** Validate AWS credentials by calling STS `GetCallerIdentity`. Format credential error messages with aws-vault fix instructions.
- **Interfaces:** `ValidateCredentials(ctx, region, profile) → (*CallerIdentity, error)`, `FormatCredentialError(profile, err) → string`.
- **Key Decisions:** Uses AWS SDK v2 config loading with optional profile. Returns structured `CallerIdentity` (Account, UserID, Arn).

### Component: ENI Finder (`pkg/aws/eni.go`)

- **Responsibility:** Find ENIs associated with a given IP address. Two-stage search: first by secondary IP filter, then by prefix delegation CIDR containment.
- **Interfaces:** `FindByIP(ctx, ip) → ([]ENIResult, error)`.
- **Key Decisions:** Secondary IP search uses the `addresses.private-ip-address` filter for efficient server-side filtering. Prefix delegation fallback lists all in-use ENIs and checks IPv4 prefix CIDRs client-side (necessary because AWS API doesn't support prefix containment filters).

### Component: Classifier (`pkg/aws/classifier.go`)

- **Responsibility:** Classify an ENI into a resource type (EC2, Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoint, Transit Gateway, EKS Control Plane, Unknown) and determine if the IP may belong to a K8s pod.
- **Interfaces:** `ClassifyENI(eni) → ClassifiedENI` (pure function). `ResourceType` enum with `DisplayName()` and `ShouldSearchK8s()` methods.
- **Key Decisions:** Classification uses ENI interface type and description string matching. `MayBePodIP` is true only for EC2-attached and unknown ENIs — all other types cannot host pods.

### Component: Instance Finder (`pkg/aws/instance.go`)

- **Responsibility:** Fetch EC2 instance details (name, type, state, IPs, tags) for a given instance ID.
- **Interfaces:** `GetDetails(ctx, instanceID) → (*InstanceDetails, error)`.
- **Key Decisions:** Returns nil (not error) for empty instance ID or no results — allows graceful handling when ENI has no instance attachment.

### Component: K8s Client (`pkg/k8s/client.go`)

- **Responsibility:** Create Kubernetes clientsets from kubeconfig with context selection. List available contexts for `--all-contexts` mode.
- **Interfaces:** `NewClient(kubeContext) → (*Client, error)`, `GetAvailableContexts() → ([]string, error)`.
- **Key Decisions:** Clears `AWS_VAULT` env var before creating K8s clients to prevent nested aws-vault conflicts with EKS authentication. Uses `clientcmd` non-interactive deferred loading for kubeconfig.

### Component: Pod Finder (`pkg/k8s/pods.go`)

- **Responsibility:** Find Kubernetes pods by pod IP (`status.podIP` field selector) and by node IP (node lookup → pod list by `spec.nodeName`).
- **Interfaces:** `FindByIP(ctx, ip) → ([]PodResult, error)`, `FindByNodeIP(ctx, nodeIP) → ([]PodResult, error)`.
- **Key Decisions:** Uses field selectors for server-side filtering. Node IP lookup finds the node name first via `NodeInternalIP` address match, then lists pods on that node.

### Component: Finder / Orchestrator (`pkg/finder/finder.go`)

- **Responsibility:** Orchestrate the complete search flow: ENI search → classification → instance details → smart K8s skip check → K8s pod search across context(s). Aggregate all results into a single `Result` struct.
- **Interfaces:** `New(ctx, opts) → (*IPFinder, error)`, `Find(ctx, ip) → (*Result, error)`.
- **Key Decisions:** Smart K8s skip checks `MayBePodIP` on all classified ENIs — skips only if none may be pod IPs. hostNetwork detection: if all matched pods use hostNetwork, additionally searches for application pods on the same node. Context iteration continues through errors (logs warnings, records errors in result).

### Component: Output Printer (`pkg/output/printer.go`)

- **Responsibility:** Format and display search results as tabular output to stdout (ENI table, EC2 instance details, K8s pod table, skip/error messages).
- **Interfaces:** `NewPrinter() → *Printer`, `PrintResult(result)`.

### Component: Logger (`pkg/logger/logger.go`)

- **Responsibility:** Leveled logging (DEBUG, INFO, WARN, ERROR) with debug mode toggle.
- **Interfaces:** Package-level functions: `SetDebug(bool)`, `Debug(fmt, args...)`, `Info(fmt, args...)`, `Warn(fmt, args...)`, `Error(fmt, args...)`.

## Data Flow

1. Developer provides an IP address as positional argument with optional flags
2. CLI validates IP format, then validates AWS credentials via STS GetCallerIdentity
3. If credentials are expired, attempts aws-vault re-exec; otherwise proceeds
4. ENI Finder searches for the IP: first via secondary IP filter, then via prefix delegation CIDR containment
5. If no ENI found → return result with K8s skip reason, stop
6. Classifier determines resource type and `MayBePodIP` flag for each ENI
7. If ENI has an attached EC2 instance, Instance Finder fetches instance details
8. Smart skip check: if `--skip-k8s` flag set or no ENI has `MayBePodIP=true` → skip K8s search
9. Pod Finder searches K8s cluster(s) for pods matching the IP; if only hostNetwork pods found, also searches for application pods on the same node
10. Output Printer renders ENI table, EC2 details, and K8s pod table to stdout

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| `github.com/aws/aws-sdk-go-v2` | AWS SDK v2 core | v1.41.1 |
| `github.com/aws/aws-sdk-go-v2/config` | AWS config loading (profiles, regions) | v1.29.0 |
| `github.com/aws/aws-sdk-go-v2/service/ec2` | EC2 API (DescribeNetworkInterfaces, DescribeInstances) | v1.281.0 |
| `github.com/aws/aws-sdk-go-v2/service/sts` | STS GetCallerIdentity (credential validation) | v1.33.8 (indirect) |
| `github.com/spf13/cobra` | CLI framework | v1.10.1 |
| `k8s.io/client-go` | Kubernetes client (kubeconfig, clientset) | v0.32.0 |
| `k8s.io/api` | Kubernetes API types (Pod, Node) | v0.32.0 |
| `k8s.io/apimachinery` | Kubernetes API machinery (metav1.ListOptions) | v0.32.0 |

## Authentication & Security

**AWS credentials:**
1. CLI validates credentials via `sts:GetCallerIdentity`
2. If expired and not already under aws-vault → re-exec under `aws-vault exec <profile> -- ip-finder <args>`
3. If expired and already under aws-vault → fail with auth error
4. If expired and aws-vault not found → print manual fix instructions

**Kubernetes credentials:**
- Uses kubeconfig (`~/.kube/config` or `$KUBECONFIG`)
- Clears `AWS_VAULT` env var before K8s client creation to prevent nested vault conflicts with EKS auth
- Context selection via `--kube-context` flag or `--all-contexts`

**All operations are strictly read-only** — no write permissions required.

## Error Handling Strategy

- Invalid IP format → immediate error with message, exit 1
- AWS credential failure → aws-vault re-exec attempt, or manual fix instructions
- No ENI found → informational message, K8s search skipped, suggest checking region/account
- K8s context connection failure → warning logged, error recorded in result, continue with other contexts
- EC2 DescribeInstances failure → silently ignored (instance details are optional enrichment)
- All errors propagate via Go error wrapping (`fmt.Errorf("...: %w", err)`)
