package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"awsct/pkg/cloudtrail"
)

var whoCmd = &cobra.Command{
	Use:   "who <action>",
	Short: "Find who performed a specific action",
	Long: `Query CloudTrail to find who performed a specific AWS API action.

Examples:
  awsct who DeleteBucket --last 24h
  awsct who RunInstances --last 7d --json
  awsct who CreateUser --last 1h --profile prod`,
	Args: cobra.ExactArgs(1),
	RunE: runWho,
}

func init() {
	rootCmd.AddCommand(whoCmd)
}

func runWho(cmd *cobra.Command, args []string) error {
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

	events, err := client.LookupByAction(ctx, args[0], cloudtrail.LookupParams{
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


