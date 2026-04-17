package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"awsct/pkg/cloudtrail"
)

var resourceCmd = &cobra.Command{
	Use:   "resource <resource-name>",
	Short: "Find what happened to a specific resource",
	Long: `Query CloudTrail to find all actions performed on a specific resource.

The resource name should match exactly as CloudTrail records it
(e.g., bucket name, instance ID, role name).

Examples:
  awsct resource my-bucket --last 7d
  awsct resource i-0abc123def456 --last 2h --json
  awsct resource my-role --last 24h --profile prod`,
	Args: cobra.ExactArgs(1),
	RunE: runResource,
}

func init() {
	rootCmd.AddCommand(resourceCmd)
}

func runResource(cmd *cobra.Command, args []string) error {
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

	events, err := client.LookupByResource(ctx, args[0], cloudtrail.LookupParams{
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
