package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"awsct/pkg/cloudtrail"
)

var serveCmd = &cobra.Command{
	Use:   "serve-mcp",
	Short: "Start MCP server (stdio transport)",
	Long:  `Start awsct as an MCP server over stdio for AI assistant integration.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	s := server.NewMCPServer(
		"awsct",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// lookup_by_action
	actionTool := mcp.NewTool("lookup_by_action",
		mcp.WithDescription("Find who performed a specific AWS API action. Returns CloudTrail events matching the action name."),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("AWS API action name, e.g. DeleteBucket, RunInstances, CreateUser"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
		),
		mcp.WithString("region",
			mcp.Description("AWS region. Default: us-east-1"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max results to return. Default: 50"),
		),
	)
	s.AddTool(actionTool, handleLookupByAction)

	// lookup_by_user
	userTool := mcp.NewTool("lookup_by_user",
		mcp.WithDescription("Find all actions performed by a specific user. Returns CloudTrail events for the given username."),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("IAM username or role session name"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
		),
		mcp.WithString("region",
			mcp.Description("AWS region. Default: us-east-1"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max results to return. Default: 50"),
		),
	)
	s.AddTool(userTool, handleLookupByUser)

	// lookup_by_resource
	resourceTool := mcp.NewTool("lookup_by_resource",
		mcp.WithDescription("Find all actions performed on a specific resource. Uses exact resource name as CloudTrail records it."),
		mcp.WithString("resource",
			mcp.Required(),
			mcp.Description("Resource name as CloudTrail records it (e.g., bucket name, instance ID, role name)"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
		),
		mcp.WithString("region",
			mcp.Description("AWS region. Default: us-east-1"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max results to return. Default: 50"),
		),
	)
	s.AddTool(resourceTool, handleLookupByResource)

	return server.ServeStdio(s)
}

func handleLookupByAction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, err := request.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := parseMCPParams(request)
	client, err := cloudtrail.NewClient(ctx, params.Region, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	events, err := client.LookupByAction(ctx, action, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Lookup failed: %v", err)), nil
	}
	return eventsToResult(events)
}

func handleLookupByUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := parseMCPParams(request)
	client, err := cloudtrail.NewClient(ctx, params.Region, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	events, err := client.LookupByUser(ctx, username, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Lookup failed: %v", err)), nil
	}
	return eventsToResult(events)
}

func handleLookupByResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resource, err := request.RequireString("resource")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := parseMCPParams(request)
	client, err := cloudtrail.NewClient(ctx, params.Region, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	events, err := client.LookupByResource(ctx, resource, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Lookup failed: %v", err)), nil
	}
	return eventsToResult(events)
}

func parseMCPParams(request mcp.CallToolRequest) cloudtrail.LookupParams {
	last := mcp.ParseString(request, "last", "24h")
	region := mcp.ParseString(request, "region", "us-east-1")
	limit := mcp.ParseInt(request, "limit", 50)

	duration, err := cloudtrail.ParseDuration(last)
	if err != nil {
		duration = 24 * 60 * 60 * 1e9 // 24h in nanoseconds
	}

	return cloudtrail.LookupParams{
		Duration: duration,
		Region:   region,
		Limit:    limit,
	}
}

func eventsToResult(events []cloudtrail.Event) (*mcp.CallToolResult, error) {
	if len(events) == 0 {
		return mcp.NewToolResultText("No events found."), nil
	}
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal events: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
