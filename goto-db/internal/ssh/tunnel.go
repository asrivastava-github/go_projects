package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
)

type Spec struct {
	AgentHost  string
	RemoteHost string
	RemotePort int
	LocalPort  int
	User       string
}

func CheckPortAvailable(port int) error {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("local port %d is already in use — use --local-port to specify a different one", port)
	}
	ln.Close()
	return nil
}

func RunTunnel(ctx context.Context, spec Spec) error {
	destination := spec.AgentHost
	if spec.User != "" {
		destination = fmt.Sprintf("%s@%s", spec.User, spec.AgentHost)
	}

	localForward := fmt.Sprintf("%d:%s:%d", spec.LocalPort, spec.RemoteHost, spec.RemotePort)

	// ControlPath=none forces a direct connection so the tunnel process stays alive.
	// Without this, SSH mux client exits immediately after setting up forwarding.
	args := []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=3",
		"-o", "ControlPath=none",
		"-L", localForward,
		destination,
	}

	fmt.Printf("🔗 Establishing tunnel: localhost:%d → %s:%d\n", spec.LocalPort, spec.RemoteHost, spec.RemotePort)
	fmt.Printf("   via: %s\n", spec.AgentHost)
	fmt.Println("   (2FA may be required)")
	fmt.Println()

	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("ssh tunnel failed: %w", err)
	}

	return nil
}
