package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"

	"awsct/pkg/cloudtrail"
)

var (
	flagLast    string
	flagRegion  string
	flagProfile string
	flagJSON    bool
	flagLimit   int
)

var rootCmd = &cobra.Command{
	Use:   "awsct",
	Short: "AWS CloudTrail event finder",
	Long: `awsct queries AWS CloudTrail to find:
  - Who performed a specific action (awsct who)
  - What a specific user has done (awsct user)
  - What happened to a specific resource (awsct resource)

It can also run as an MCP server (awsct serve-mcp) for AI assistant integration.

Examples:
  awsct who DeleteBucket --last 24h
  awsct user alice --last 1h --json
  awsct resource my-bucket --last 7d
  awsct serve-mcp`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagLast, "last", "24h", "Time window (e.g., 30m, 1h, 7d)")
	rootCmd.PersistentFlags().StringVar(&flagRegion, "region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "AWS profile for aws-vault")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().IntVar(&flagLimit, "limit", 50, "Max results to return")
}

func ensureCredentials(ctx context.Context) {
	underVault := os.Getenv("AWS_VAULT") != ""
	profile := flagProfile

	if underVault {
		profile = ""
	}

	if profile == "" && underVault {
		return
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(flagRegion),
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err == nil {
		stsClient := sts.NewFromConfig(cfg)
		_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	}

	if err != nil {
		if underVault {
			log.Fatalf("AWS credentials failed under aws-vault: %v", err)
		}
		if flagProfile == "" {
			log.Fatalf("AWS credentials invalid. Use --profile to specify an AWS profile: %v", err)
		}
		reExecWithVault()
	}
}

func reExecWithVault() {
	vaultPath, err := exec.LookPath("aws-vault")
	if err != nil {
		log.Fatalf("aws-vault not found in PATH. Install it or set valid AWS credentials.")
	}

	log.Printf("🔐 Session expired. Running aws-vault exec %s...", flagProfile)

	cmdArgs := append([]string{"exec", flagProfile, "--", os.Args[0]}, os.Args[1:]...)
	cmd := exec.Command(vaultPath, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("aws-vault exec failed: %v", err)
	}

	os.Exit(0)
}

func effectiveProfile() string {
	if os.Getenv("AWS_VAULT") != "" {
		return ""
	}
	return flagProfile
}

func printEvents(events []cloudtrail.Event, jsonOutput bool) {
	if len(events) == 0 {
		fmt.Println("No events found.")
		return
	}

	if jsonOutput {
		printJSON(events)
	} else {
		printTable(events)
	}
}

func printTable(events []cloudtrail.Event) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tUSER\tACTION\tRESOURCE\tSOURCE")
	fmt.Fprintln(w, "---------\t----\t------\t--------\t------")

	for _, e := range events {
		resource := "-"
		if len(e.Resources) > 0 {
			resource = e.Resources[0].Name
			if resource == "" {
				resource = e.Resources[0].Type
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp.Format(time.DateTime),
			e.Username,
			e.EventName,
			resource,
			e.EventSource,
		)
	}
	w.Flush()
}

func printJSON(events []cloudtrail.Event) {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal events: %v", err)
	}
	fmt.Println(string(data))
}
