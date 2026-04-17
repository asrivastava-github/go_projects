package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"gcpat/pkg/auditlog"
)

var (
	flagLast    string
	flagProject string
	flagJSON    bool
	flagLimit   int
)

var rootCmd = &cobra.Command{
	Use:   "gcpat",
	Short: "GCP Audit Trail event finder",
	Long: `gcpat queries GCP Cloud Audit Logs to find:
  - Who performed a specific action (gcpat who)
  - What a specific user has done (gcpat user)
  - What happened to a specific resource (gcpat resource)

It can also run as an MCP server (gcpat serve-mcp) for AI assistant integration.

Examples:
  gcpat who compute.instances.delete --project my-project --last 24h
  gcpat user alice@example.com --project my-project --last 1h --json
  gcpat resource my-bucket --project my-project --last 7d
  gcpat serve-mcp`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagLast, "last", "24h", "Time window (e.g., 30m, 1h, 7d)")
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "GCP project ID")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().IntVar(&flagLimit, "limit", 50, "Max results to return")
}

func ensureCredentials() {
	cmd := exec.Command("gcloud", "auth", "application-default", "print-access-token")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		log.Fatalf("GCP credentials not configured. Run: gcloud auth application-default login")
	}
}

func requireProject() string {
	if flagProject == "" {
		log.Fatalf("--project is required. Specify a GCP project ID.")
	}
	return flagProject
}

func printEvents(events []auditlog.Event, jsonOutput bool) {
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

func printTable(events []auditlog.Event) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tPRINCIPAL\tMETHOD\tRESOURCE\tSERVICE")
	fmt.Fprintln(w, "---------\t---------\t------\t--------\t-------")

	for _, e := range events {
		resource := e.ResourceName
		if resource == "" {
			resource = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp.Format(time.DateTime),
			e.Principal,
			e.Method,
			resource,
			e.ServiceName,
		)
	}
	w.Flush()
}

func printJSON(events []auditlog.Event) {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal events: %v", err)
	}
	fmt.Println(string(data))
}
