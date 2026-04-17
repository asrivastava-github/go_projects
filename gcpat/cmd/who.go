package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"gcpat/pkg/auditlog"
)

var whoCmd = &cobra.Command{
	Use:   "who <method>",
	Short: "Find who performed a specific action",
	Long: `Query GCP Cloud Audit Logs to find who performed a specific API method.

Examples:
  gcpat who compute.instances.delete --project my-project --last 24h
  gcpat who storage.buckets.delete --project my-project --last 7d --json
  gcpat who iam.serviceAccounts.create --project my-project --last 1h`,
	Args: cobra.ExactArgs(1),
	RunE: runWho,
}

func init() {
	rootCmd.AddCommand(whoCmd)
}

func runWho(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ensureCredentials()
	project := requireProject()

	duration, err := auditlog.ParseDuration(flagLast)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	client, err := auditlog.NewClient(ctx, project)
	if err != nil {
		return err
	}
	defer client.Close()

	events, err := client.LookupByAction(ctx, args[0], auditlog.LookupParams{
		Duration:  duration,
		ProjectID: project,
		Limit:     flagLimit,
	})
	if err != nil {
		return err
	}

	printEvents(events, flagJSON)
	return nil
}
