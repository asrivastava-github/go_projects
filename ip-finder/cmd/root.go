package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"ip-finder/pkg/aws"
	"ip-finder/pkg/finder"
	"ip-finder/pkg/logger"
	"ip-finder/pkg/output"
)

var opts finder.Options
var debug bool

var rootCmd = &cobra.Command{
	Use:   "ip-finder <ip-address>",
	Short: "Find AWS resources (EC2, Lambda, EKS pods) by internal IP address",
	Long: `IP Finder searches AWS ENIs and Kubernetes clusters to identify 
which resource owns a given internal IP address.

It searches:
- EC2 instances (primary and secondary IPs)
- Lambda functions (VPC ENIs)
- EKS pods (via VPC CNI secondary IPs)

Examples:
  ip-finder 10.0.1.50
  ip-finder 10.0.1.50 --region us-west-2
  ip-finder 10.0.1.50 --kube-context prod-cluster
  ip-finder 10.0.1.50 --all-contexts`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&opts.Region, "region", "r", "us-east-1", "AWS region (default: us-east-1)")
	rootCmd.Flags().StringVarP(&opts.AWSProfile, "profile", "p", "prod", "AWS profile (default: prod)")
	rootCmd.Flags().StringVarP(&opts.KubeContext, "kube-context", "k", "produe102", "Kubernetes context (default: produe102)")
	rootCmd.Flags().BoolVarP(&opts.AllContexts, "all-contexts", "a", false, "Search all available Kubernetes contexts")
	rootCmd.Flags().BoolVar(&opts.SkipK8s, "skip-k8s", false, "Skip Kubernetes pod search")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ip := args[0]

	logger.SetDebug(debug)
	logger.Debug("Debug mode enabled")
	logger.Debug("Options: region=%s, profile=%s, context=%s", opts.Region, opts.AWSProfile, opts.KubeContext)

	if !isValidIP(ip) {
		logger.Error("Invalid IP address format: %s", ip)
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	logger.Debug("Validated IP format: %s", ip)

	underVault := os.Getenv("AWS_VAULT") != ""
	profile := opts.AWSProfile
	if underVault {
		logger.Debug("Running under aws-vault, using environment credentials")
		profile = ""
	}

	logger.Debug("Validating AWS credentials for region=%s, profile=%s", opts.Region, profile)
	identity, err := aws.ValidateCredentials(ctx, opts.Region, profile)
	if err != nil {
		if underVault {
			logger.Error("AWS credentials failed under aws-vault")
			return fmt.Errorf("AWS credentials failed even under aws-vault. Check your aws-vault/SSO configuration.\n\nError: %v", err)
		}
		logger.Warn("AWS credentials expired, attempting aws-vault refresh")
		if reExecWithVault() {
			return nil
		}
		return fmt.Errorf(aws.FormatCredentialError(opts.AWSProfile, err))
	}
	logger.Info("Authenticated as %s (Account: %s)", identity.Arn, identity.Account)

	finderOpts := opts
	if underVault {
		finderOpts.AWSProfile = ""
	}

	logger.Debug("Initializing IP finder")
	ipFinder, err := finder.New(ctx, finderOpts)
	if err != nil {
		logger.Error("Failed to initialize finder: %v", err)
		return fmt.Errorf("failed to initialize finder: %w", err)
	}

	logger.Debug("Starting search for IP: %s", ip)
	result, err := ipFinder.Find(ctx, ip)
	if err != nil {
		logger.Error("Search failed: %v", err)
		return fmt.Errorf("search failed: %w", err)
	}

	logger.Debug("Search complete. ENIs found: %d", len(result.ENIs))
	printer := output.NewPrinter()
	printer.PrintResult(result)

	return nil
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func reExecWithVault() bool {
	if os.Getenv("AWS_VAULT") != "" {
		logger.Debug("Already running under aws-vault, skipping re-exec")
		return false
	}

	vaultPath, err := exec.LookPath("aws-vault")
	if err != nil {
		logger.Warn("aws-vault not found in PATH")
		return false
	}

	logger.Info("Session expired. Running aws-vault exec %s...", opts.AWSProfile)

	cmdArgs := append([]string{"exec", opts.AWSProfile, "--", os.Args[0]}, os.Args[1:]...)
	logger.Debug("Executing: aws-vault %v", cmdArgs)

	cmd := exec.Command(vaultPath, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error("aws-vault exec failed: %v", err)
		return false
	}

	return true
}
