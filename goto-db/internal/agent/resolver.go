package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"goto-db/internal/cli"
	"goto-db/internal/config"
)

const (
	defaultAgent  = "jenkins-agent70215c.prod.svc.ue1.viatorsystems.com"
	agentDNSSuffix = ".prod.svc.ue1.viatorsystems.com"
)

// expandAgent auto-appends the domain suffix if a short name like "jenkins-agent70" is given.
func expandAgent(agent string) string {
	if !strings.Contains(agent, ".") {
		return agent + agentDNSSuffix
	}
	return agent
}

func Resolve(ctx context.Context, opts *cli.Options, cfg *config.Config) (string, error) {
	// Priority 1: explicit --agent flag
	if opts.JenkinsAgent != "" {
		agent := expandAgent(opts.JenkinsAgent)
		cfg.JenkinsAgent = agent
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to cache agent: %v\n", err)
		}
		return agent, nil
	}

	// Priority 2: --refresh flag (prompt user)
	if opts.Refresh {
		agent, err := promptForAgent()
		if err != nil {
			return "", err
		}
		cfg.JenkinsAgent = agent
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to cache agent: %v\n", err)
		}
		return agent, nil
	}

	// Priority 3: cached config
	if cfg.JenkinsAgent != "" {
		return cfg.JenkinsAgent, nil
	}

	// Priority 4: default agent
	cfg.JenkinsAgent = defaultAgent
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to cache agent: %v\n", err)
	}
	fmt.Printf("Using default Jenkins agent: %s\n", defaultAgent)
	return defaultAgent, nil
}

func promptForAgent() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter Jenkins agent hostname [%s]: ", defaultAgent)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultAgent, nil
	}
	return expandAgent(input), nil
}
