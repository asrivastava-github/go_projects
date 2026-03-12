package ui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ConnectionParams struct {
	Name   string
	Host   string
	Port   int
	Engine string
}

const (
	containerName = "goto-db-dbgate"
	imageName     = "dbgate/dbgate"
	ContainerPort = 8978
)

func dbgateEngine(engine string) string {
	switch engine {
	case "mysql":
		return "mysql@dbgate-plugin-mysql"
	default:
		return "postgres@dbgate-plugin-postgres"
	}
}

func StartUI(ctx context.Context, params ConnectionParams) error {
	if !isDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	if isContainerRunning() {
		fmt.Printf("🌐 DbGate already running at http://localhost:%d\n", ContainerPort)
		return nil
	}

	removeContainer()

	fmt.Println("🌐 Starting DbGate...")

	args := []string{
		"run", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:3000", ContainerPort),
		"--add-host=host.docker.internal:host-gateway",
		"-e", "CONNECTIONS=db",
		"-e", fmt.Sprintf("LABEL_db=%s", params.Name),
		"-e", "SERVER_db=host.docker.internal",
		"-e", fmt.Sprintf("PORT_db=%d", params.Port),
		"-e", fmt.Sprintf("ENGINE_db=%s", dbgateEngine(params.Engine)),
		"-e", "PASSWORD_MODE_db=askUser",
		imageName,
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start DbGate: %w", err)
	}

	if err := waitForReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: DbGate may still be starting: %v\n", err)
	}

	fmt.Printf("🌐 DbGate ready at http://localhost:%d\n", ContainerPort)
	return nil
}

func StopUI() {
	if isContainerRunning() {
		exec.Command("docker", "stop", containerName).Run()
		exec.Command("docker", "rm", containerName).Run()
		fmt.Println("🌐 DbGate stopped and removed")
	}
}

func BrowserURL(params ConnectionParams) string {
	return fmt.Sprintf("http://localhost:%d", ContainerPort)
}

func PrintConnectionInfo(localPort int, engine string) {
	fmt.Println()
	fmt.Println("  ┌─ DB UI Ready ─────────────────────────────────────┐")
	fmt.Println("  │                                                    │")
	fmt.Printf("  │  Server:  host.docker.internal:%-5d               │\n", localPort)
	fmt.Printf("  │  Engine:  %-10s                                │\n", engine)
	fmt.Println("  │  Connection pre-configured — enter credentials.    │")
	fmt.Println("  │                                                    │")
	fmt.Println("  └────────────────────────────────────────────────────┘")
}

func OpenBrowser(url string) {
	exec.Command("open", url).Run()
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func isContainerRunning() bool {
	out, err := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName)).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func removeContainer() {
	exec.Command("docker", "rm", "-f", containerName).Run()
}

func waitForReady(ctx context.Context) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	url := fmt.Sprintf("http://localhost:%d", ContainerPort)
	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for DbGate to start")
		case <-ticker.C:
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					return nil
				}
			}
		}
	}
}
