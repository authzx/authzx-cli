package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var (
	agentConfig string
	agentListen string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the local AuthzX agent",
}

var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the AuthzX agent locally",
	Long: `Start the AuthzX agent locally. The agent syncs policy bundles from
the cloud and serves authorization decisions over HTTP.

Requires the authzx-agent binary in PATH.
Install: go install github.com/authzx/authzx-agent/cmd/agent@latest`,
	RunE: func(cmd *cobra.Command, args []string) error {
		binary, err := exec.LookPath("authzx-agent")
		if err != nil {
			return fmt.Errorf("authzx-agent binary not found in PATH\n\nInstall it with:\n  go install github.com/authzx/authzx-agent/cmd/agent@latest\n\nOr run via Docker:\n  docker run -p 8181:8181 -v ./authzx-agent.yaml:/etc/authzx/agent.yaml ghcr.io/authzx/agent")
		}

		proc := exec.Command(binary, "--config", agentConfig, "--listen", agentListen)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		proc.Stdin = os.Stdin

		fmt.Printf("Starting agent: %s --config %s --listen %s\n", binary, agentConfig, agentListen)
		return proc.Run()
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check agent health",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("http://%s/healthz", agentListen)
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("agent not reachable at %s: %w", agentListen, err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Agent is running at %s\n", agentListen)
			fmt.Printf("  %s\n", string(body))
		} else {
			fmt.Printf("Agent responded with status %d: %s\n", resp.StatusCode, string(body))
		}
		return nil
	},
}

func init() {
	agentStartCmd.Flags().StringVar(&agentConfig, "config", "authzx-agent.yaml", "Agent config file")
	agentStartCmd.Flags().StringVar(&agentListen, "listen", "127.0.0.1:8181", "Agent listen address")
	agentStatusCmd.Flags().StringVar(&agentListen, "listen", "127.0.0.1:8181", "Agent listen address")

	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStatusCmd)
}
