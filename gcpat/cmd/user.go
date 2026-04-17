package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"gcpat/pkg/auditlog"
)

var userCmd = &cobra.Command{
	Use:   "user <principal>",
	Short: "Find what a specific user has done",
	Long: `Query GCP Cloud Audit Logs to find all actions performed by a specific user.

Examples:
  gcpat user alice@example.com --project my-project --last 1h
  gcpat user alice@example.com --project my-project --last 2h --json
  gcpat user sa@project.iam.gserviceaccount.com --project my-project --last 7d`,
	Args: cobra.ExactArgs(1),
	RunE: runUser,
}

func init() {
	rootCmd.AddCommand(userCmd)
}

func runUser(cmd *cobra.Command, args []string) error {
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

	events, err := client.LookupByUser(ctx, args[0], auditlog.LookupParams{
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
