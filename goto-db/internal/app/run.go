package app

import (
	"context"
	"fmt"
	"os"

	"goto-db/internal/agent"
	"goto-db/internal/cli"
	"goto-db/internal/config"
	"goto-db/internal/db"
	"goto-db/internal/ssh"
	"goto-db/internal/ui"
)

func Run(ctx context.Context, args []string) error {
	opts, err := cli.Parse(args)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	agentHost, err := agent.Resolve(ctx, opts, cfg)
	if err != nil {
		return fmt.Errorf("resolving jenkins agent: %w", err)
	}

	// If --refresh was the only intent, we're done
	if opts.Refresh && opts.DBName == "" && opts.DBURL == "" {
		fmt.Printf("✅ Jenkins agent updated: %s\n", agentHost)
		return nil
	}

	target, err := db.ResolveTarget(opts)
	if err != nil {
		return fmt.Errorf("resolving db target: %w", err)
	}

	// Check port availability before starting anything
	if err := ssh.CheckPortAvailable(target.LocalPort); err != nil {
		return err
	}

	printConnectionInfo(target, agentHost)

	// Start DB UI
	connParams := ui.ConnectionParams{
		Name:   target.Host,
		Host:   "host.docker.internal",
		Port:   target.LocalPort,
		Engine: target.Engine,
	}
	if uiErr := ui.StartUI(ctx, connParams); uiErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not start DB UI: %v\n", uiErr)
	} else {
		ui.PrintConnectionInfo(target.LocalPort, target.Engine)
		ui.OpenBrowser(ui.BrowserURL(connParams))
	}

	// Run SSH tunnel (blocks until cancelled)
	tunnelErr := ssh.RunTunnel(ctx, ssh.Spec{
		AgentHost:  agentHost,
		RemoteHost: target.Host,
		RemotePort: target.Port,
		LocalPort:  target.LocalPort,
		User:       opts.User,
	})

	// Cleanup: stop and remove container
	ui.StopUI()

	return tunnelErr
}

func printConnectionInfo(target *db.Target, agentHost string) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Database:    %s:%d\n", target.Host, target.Port)
	fmt.Printf("  Engine:      %s\n", target.Engine)
	fmt.Printf("  Agent:       %s\n", agentHost)
	fmt.Printf("  Local:       localhost:%d\n", target.LocalPort)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Connect with: %s -h localhost -p %d\n", db.DefaultClient(target.Engine), target.LocalPort)
	fmt.Println()
}
