package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"gcpat/pkg/auditlog"
)

var serveCmd = &cobra.Command{
	Use:   "serve-mcp",
	Short: "Start MCP server (stdio transport)",
	Long:  `Start gcpat as an MCP server over stdio for AI assistant integration.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	s := server.NewMCPServer(
		"gcpat",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// lookup_by_action
	actionTool := mcp.NewTool("lookup_by_action",
		mcp.WithDescription("Find who performed a specific GCP API method. Returns Cloud Audit Log events matching the method name."),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("GCP API method name, e.g. compute.instances.delete, storage.buckets.delete, iam.serviceAccounts.create"),
		),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("GCP project ID"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max results to return. Default: 50"),
		),
	)
	s.AddTool(actionTool, handleLookupByAction)

	// lookup_by_user
	userTool := mcp.NewTool("lookup_by_user",
		mcp.WithDescription("Find all actions performed by a specific user. Returns Cloud Audit Log events for the given principal email."),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Principal email address (user or service account)"),
		),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("GCP project ID"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max results to return. Default: 50"),
		),
	)
	s.AddTool(userTool, handleLookupByUser)

	// lookup_by_resource
	resourceTool := mcp.NewTool("lookup_by_resource",
		mcp.WithDescription("Find all actions performed on a specific resource. Uses resource name substring match in Cloud Audit Logs."),
		mcp.WithString("resource",
			mcp.Required(),
			mcp.Description("Resource name as Cloud Audit Logs records it"),
		),
		mcp.WithString("project",
			mcp.Required(),
			mcp.Description("GCP project ID"),
		),
		mcp.WithString("last",
			mcp.Description("Time window, e.g. 30m, 1h, 24h, 7d. Default: 24h"),
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
	params, err := parseMCPParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client, err := auditlog.NewClient(ctx, params.ProjectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	defer client.Close()
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
	params, err := parseMCPParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client, err := auditlog.NewClient(ctx, params.ProjectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	defer client.Close()
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
	params, err := parseMCPParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client, err := auditlog.NewClient(ctx, params.ProjectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create client: %v", err)), nil
	}
	defer client.Close()
	events, err := client.LookupByResource(ctx, resource, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Lookup failed: %v", err)), nil
	}
	return eventsToResult(events)
}

func parseMCPParams(request mcp.CallToolRequest) (auditlog.LookupParams, error) {
	project, err := request.RequireString("project")
	if err != nil {
		return auditlog.LookupParams{}, err
	}

	last := mcp.ParseString(request, "last", "24h")
	limit := mcp.ParseInt(request, "limit", 50)

	duration, err := auditlog.ParseDuration(last)
	if err != nil {
		duration = 24 * 60 * 60 * 1e9 // 24h in nanoseconds
	}

	return auditlog.LookupParams{
		Duration:  duration,
		ProjectID: project,
		Limit:     limit,
	}, nil
}

func eventsToResult(events []auditlog.Event) (*mcp.CallToolResult, error) {
	if len(events) == 0 {
		return mcp.NewToolResultText("No events found."), nil
	}
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal events: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
