package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"gcpat/pkg/auditlog"
)

var resourceCmd = &cobra.Command{
	Use:   "resource <resource-name>",
	Short: "Find what happened to a specific resource",
	Long: `Query GCP Cloud Audit Logs to find all actions performed on a specific resource.

The resource name should match as Cloud Audit Logs records it.

Examples:
  gcpat resource my-bucket --project my-project --last 7d
  gcpat resource my-instance --project my-project --last 2h --json
  gcpat resource my-service-account --project my-project --last 24h`,
	Args: cobra.ExactArgs(1),
	RunE: runResource,
}

func init() {
	rootCmd.AddCommand(resourceCmd)
}

func runResource(cmd *cobra.Command, args []string) error {
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

	events, err := client.LookupByResource(ctx, args[0], auditlog.LookupParams{
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
