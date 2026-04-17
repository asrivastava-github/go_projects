package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"awsct/pkg/cloudtrail"
)

var userCmd = &cobra.Command{
	Use:   "user <username>",
	Short: "Find what a specific user has done",
	Long: `Query CloudTrail to find all actions performed by a specific user.

Examples:
  awsct user alice --last 1h
  awsct user alice --last 2h --json
  awsct user root --last 7d --profile prod`,
	Args: cobra.ExactArgs(1),
	RunE: runUser,
}

func init() {
	rootCmd.AddCommand(userCmd)
}

func runUser(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ensureCredentials(ctx)

	duration, err := cloudtrail.ParseDuration(flagLast)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	client, err := cloudtrail.NewClient(ctx, flagRegion, effectiveProfile())
	if err != nil {
		return err
	}

	events, err := client.LookupByUser(ctx, args[0], cloudtrail.LookupParams{
		Duration: duration,
		Region:   flagRegion,
		Limit:    flagLimit,
	})
	if err != nil {
		return err
	}

	printEvents(events, flagJSON)
	return nil
}
