# ip-finder — Tasks

**Date:** 2026-04-23
**Spec:** `specs/requirements.md`

## Status Legend

- [ ] Not started
- [x] Complete
- 🔄 In progress

## Task 1: Project Scaffold

**Traces to:** UC-IP-001
**Files:** `main.go`, `go.mod`, `Makefile`, `cmd/root.go`

- [x] Step 1: Create go.mod (`go mod init ip-finder`)
- [x] Step 2: Create main.go (entry point calling `cmd.Execute()`)
- [x] Step 3: Create cmd/root.go with cobra root command, flags (`--profile`, `--region`, `--kube-context`, `--all-contexts`, `--skip-k8s`, `--debug`), IP validation, and aws-vault re-exec logic
- [x] Step 4: Create Makefile (build, install, tidy, test, clean, lint targets)
- [x] Step 5: Add cobra dependency and verify build
- [x] Step 6: Commit

## Task 2: AWS Client & Credential Validation

**Traces to:** UC-IP-001
**Files:** `pkg/aws/client.go`, `pkg/aws/credentials.go`

- [x] Step 1: Create client.go — AWS client factory (EC2 client from SDK v2 config with region/profile)
- [x] Step 2: Create credentials.go — `ValidateCredentials` via STS GetCallerIdentity, `FormatCredentialError` with aws-vault instructions
- [x] Step 3: Add AWS SDK v2 dependencies (config, ec2, sts)
- [x] Step 4: Commit

## Task 3: ENI Lookup

**Traces to:** UC-IP-001
**Files:** `pkg/aws/eni.go`

- [x] Step 1: Create eni.go — `ENIFinder` with `FindByIP` method
- [x] Step 2: Implement secondary IP search via `DescribeNetworkInterfaces` with `addresses.private-ip-address` filter
- [x] Step 3: Implement prefix delegation fallback — list in-use ENIs, check IPv4 prefix CIDR containment
- [x] Step 4: Create `ENIResult` struct with all relevant fields (ID, type, description, instance ID, IPs, VPC, subnet, AZ, tags, prefix match info)
- [x] Step 5: Commit

## Task 4: Resource Classification

**Traces to:** UC-IP-001, UC-IP-002
**Files:** `pkg/aws/classifier.go`

- [x] Step 1: Define `ResourceType` enum (EC2, Lambda, ELB, NAT Gateway, RDS, EFS, VPC Endpoint, Transit Gateway, EKS Control Plane, Unknown)
- [x] Step 2: Create `ClassifyENI` function — classify by interface type and description string matching
- [x] Step 3: Set `MayBePodIP=true` only for EC2-attached and unknown ENIs
- [x] Step 4: Add `DisplayName()` and `ShouldSearchK8s()` methods
- [x] Step 5: Commit

## Task 5: EC2 Instance Details

**Traces to:** UC-IP-001
**Files:** `pkg/aws/instance.go`

- [x] Step 1: Create instance.go — `InstanceFinder` with `GetDetails` method via `DescribeInstances`
- [x] Step 2: Create `InstanceDetails` struct (ID, name, type, state, IPs, tags)
- [x] Step 3: Extract `Name` tag from instance tags
- [x] Step 4: Commit

## Task 6: Kubernetes Client & Pod Search

**Traces to:** UC-IP-001, UC-IP-003
**Files:** `pkg/k8s/client.go`, `pkg/k8s/pods.go`

- [x] Step 1: Create client.go — K8s client factory with kubeconfig loading, context selection, and `AWS_VAULT` env cleanup
- [x] Step 2: Create `GetAvailableContexts` for `--all-contexts` support
- [x] Step 3: Create pods.go — `PodFinder` with `FindByIP` (field selector `status.podIP`) and `FindByNodeIP` (node lookup → pod list by `spec.nodeName`)
- [x] Step 4: Create `PodResult` struct (name, namespace, pod IP, node name, node IP, status, hostNetwork, labels, annotations)
- [x] Step 5: Add client-go and k8s.io dependencies
- [x] Step 6: Commit

## Task 7: Finder Orchestrator

**Traces to:** UC-IP-001, UC-IP-002, UC-IP-003, UC-IP-004
**Files:** `pkg/finder/finder.go`

- [x] Step 1: Create finder.go — `IPFinder` struct with `New` constructor and `Find` method
- [x] Step 2: Implement search flow: ENI search → classification → instance details → smart K8s skip → K8s pod search
- [x] Step 3: Implement smart K8s skip logic (check `MayBePodIP` across all classified ENIs)
- [x] Step 4: Implement `--skip-k8s` flag handling
- [x] Step 5: Implement `--all-contexts` vs single context K8s search
- [x] Step 6: Implement hostNetwork detection — if all matched pods use hostNetwork, search for application pods on same node
- [x] Step 7: Create `Result`, `PodSearchResult`, `K8sError` structs
- [x] Step 8: Commit

## Task 8: Output Formatting

**Traces to:** UC-IP-001, UC-IP-002, UC-IP-003, UC-IP-004
**Files:** `pkg/output/printer.go`

- [x] Step 1: Create printer.go — `Printer` with `PrintResult` method
- [x] Step 2: Implement ENI table output (ID, resource type, status, instance ID, description, VPC/subnet/AZ, IPs)
- [x] Step 3: Implement EC2 instance details output
- [x] Step 4: Implement K8s pod table output (namespace, name, pod IP, node, status)
- [x] Step 5: Implement skip/error messages for K8s search
- [x] Step 6: Commit

## Task 9: Logger

**Traces to:** UC-IP-001
**Files:** `pkg/logger/logger.go`

- [x] Step 1: Create logger.go — leveled logging (DEBUG, INFO, WARN, ERROR) with `SetDebug` toggle
- [x] Step 2: Commit

## Task 10: README & Documentation

**Traces to:** UC-IP-001, UC-IP-002, UC-IP-003, UC-IP-004
**Files:** `README.md`

- [x] Step 1: Create README.md with architecture diagram, usage examples, options table, identified resource types, example output, limitations
- [x] Step 2: Final build and smoke test
- [x] Step 3: Commit

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-04-23 | Initial spec created; all tasks complete (retroactive documentation) | AS |
