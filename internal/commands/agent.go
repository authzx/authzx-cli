package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
)

const (
	defaultImage         = "authzx/agent:latest"
	defaultContainerName = "authzx-agent"
	defaultPort          = "8181"
	defaultConfigFile    = "authzx-agent.yaml"
)

var (
	agentImage      string
	agentName       string
	agentPort       string
	agentConfig     string
	agentForeground bool
	agentEnvVars    []string
	agentLogsTail   string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the local AuthzX agent (Docker)",
	Long: `Manage a local AuthzX agent running in Docker.

Commands:
  azx agent config    Generate agent config file with your API key
  azx agent start     Start the agent container
  azx agent status    Check agent health
  azx agent stop      Stop and remove the agent container
  azx agent logs      Tail agent logs`,
}

// --- azx agent config ---

var agentConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate agent config file",
	Long: `Generate an authzx-agent.yaml config file with your API key prefilled.

The API key is read from ~/.authzx/config.yaml (set by 'azx configure').
All other settings use sensible defaults and can be edited after generation.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath := agentConfig

		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("%s already exists — delete it first or edit it directly", outPath)
		}

		apiKey := ""
		cfg, err := credentials.Load()
		if err == nil && cfg.APIKey != "" {
			apiKey = cfg.APIKey
		}
		if k := os.Getenv("AUTHZX_API_KEY"); k != "" {
			apiKey = k
		}

		if apiKey == "" {
			apiKey = "<your-api-key-here>"
			fmt.Fprintf(os.Stderr, "Warning: no API key found. Run 'azx configure' first, or edit the file manually.\n\n")
		}

		yaml := fmt.Sprintf(`# AuthzX Agent Configuration
# Env vars (AUTHZX_*) override these values.
# Docs: https://docs.authzx.com/agent

api_key: "%s"

# AuthzX cloud URL (where bundles are pulled from)
cloud_url: "https://api.authzx.com"

# Agent listen address (inside the container, 0.0.0.0 for Docker)
listen_addr: "0.0.0.0:8181"

# Bundle sync interval
poll_interval: "30s"

# Local bundle cache (persisted via Docker volume)
cache_dir: "/var/lib/authzx/bundles"

# Log level: debug, info, warn, error
log_level: "info"

# Structured JSON decision logs (stdout). Enable for debugging.
# decision_log: false

# Audit forwarding — push decisions to cloud for observability dashboard
audit_forwarding: true

# Agent identity (auto-detected from hostname if not set)
# agent_name: "my-agent"
# agent_region: "us-east-1"

# Bundle signature verification (optional, for regulated environments)
# bundle_signature_required: false
`, apiKey)

		if err := os.WriteFile(outPath, []byte(yaml), 0600); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		abs, _ := filepath.Abs(outPath)
		fmt.Printf("Created %s\n", abs)
		if apiKey != "<your-api-key-here>" {
			fmt.Printf("  API key: %s\n", credentials.MaskKey(apiKey))
		}
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  azx agent start          # start the agent")
		fmt.Println("  azx agent status         # verify it's running")
		return nil
	},
}

// --- azx agent start ---

var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the AuthzX agent in Docker",
	Long: `Start a local AuthzX agent container.

Requires Docker to be installed and running.

Examples:
  azx agent start
  azx agent start --image authzx/agent:v0.1.3
  azx agent start --port 9090 --env AUTHZX_LOG_LEVEL=debug
  azx agent start --foreground`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDocker(); err != nil {
			return err
		}

		if isContainerRunning(agentName) {
			return fmt.Errorf("agent container %q is already running\n\nUse 'azx agent stop' to stop it first, or 'azx agent status' to check health", agentName)
		}

		removeStoppedContainer(agentName)

		configAbs, err := filepath.Abs(agentConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve config path: %w", err)
		}
		if _, err := os.Stat(configAbs); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s\n\nRun 'azx agent config' to generate one", configAbs)
		}

		dockerArgs := []string{"run"}

		if agentForeground {
			dockerArgs = append(dockerArgs, "--rm")
		} else {
			dockerArgs = append(dockerArgs, "-d")
		}

		dockerArgs = append(dockerArgs,
			"--name", agentName,
			"-p", agentPort+":8181",
			"-v", configAbs+":/etc/authzx/agent.yaml:ro",
		)

		for _, env := range agentEnvVars {
			dockerArgs = append(dockerArgs, "-e", env)
		}

		dockerArgs = append(dockerArgs, agentImage, "--config", "/etc/authzx/agent.yaml")

		proc := exec.Command("docker", dockerArgs...)

		if agentForeground {
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr
			proc.Stdin = os.Stdin
			fmt.Printf("Starting agent (foreground): %s on port %s\n", agentImage, agentPort)
			return proc.Run()
		}

		out, err := proc.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to start agent container: %w\n%s", err, string(out))
		}

		containerID := strings.TrimSpace(string(out))
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}

		fmt.Printf("Agent started: %s\n", containerID)
		fmt.Printf("  Image:     %s\n", agentImage)
		fmt.Printf("  Port:      localhost:%s\n", agentPort)
		fmt.Printf("  Config:    %s\n", configAbs)
		fmt.Printf("  Container: %s\n", agentName)
		fmt.Println()
		fmt.Println("Next:")
		fmt.Println("  azx agent status         # check health")
		fmt.Println("  azx agent logs           # tail logs")
		fmt.Println("  azx check --local ...    # test a decision")
		fmt.Println("  azx agent stop           # stop the agent")
		return nil
	},
}

// --- azx agent status ---

var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check agent health",
	Long: `Check the health of a running AuthzX agent.

Hits the /healthz endpoint and displays the agent's status,
bundle revision, sync age, and degradation state.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("http://localhost:%s/healthz", agentPort)
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			// Check if container exists but isn't healthy
			if isContainerRunning(agentName) {
				return fmt.Errorf("agent container is running but /healthz is not responding at %s\n\nCheck logs: azx agent logs", url)
			}
			return fmt.Errorf("agent not reachable at localhost:%s\n\nStart it with: azx agent start", agentPort)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var health map[string]interface{}
		if err := json.Unmarshal(body, &health); err != nil {
			fmt.Printf("Agent responded (status %d):\n  %s\n", resp.StatusCode, string(body))
			return nil
		}

		fmt.Printf("Agent is running at localhost:%s\n\n", agentPort)

		if status, ok := health["status"]; ok {
			fmt.Printf("  Status:       %v\n", status)
		}
		if rev, ok := health["bundle_revision"]; ok {
			fmt.Printf("  Bundle:       %v\n", rev)
		}
		if age, ok := health["sync_age_seconds"]; ok {
			fmt.Printf("  Last sync:    %.0fs ago\n", age)
		}
		if degraded, ok := health["degraded"]; ok {
			if d, ok := degraded.(bool); ok && d {
				fmt.Printf("  Degraded:     \033[31myes\033[0m\n")
				if failures, ok := health["consecutive_failures"]; ok {
					fmt.Printf("  Failures:     %.0f consecutive\n", failures)
				}
			} else {
				fmt.Printf("  Degraded:     no\n")
			}
		}

		return nil
	},
}

// --- azx agent stop ---

var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the agent container",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDocker(); err != nil {
			return err
		}

		if !containerExists(agentName) {
			fmt.Printf("No agent container %q found.\n", agentName)
			return nil
		}

		if isContainerRunning(agentName) {
			stop := exec.Command("docker", "stop", agentName)
			if out, err := stop.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to stop container: %w\n%s", err, string(out))
			}
			fmt.Printf("Agent stopped: %s\n", agentName)
		}

		rm := exec.Command("docker", "rm", agentName)
		if out, err := rm.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to remove container: %w\n%s", err, string(out))
		}
		fmt.Printf("Container removed: %s\n", agentName)
		return nil
	},
}

// --- azx agent logs ---

var agentLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail agent logs",
	Long: `Stream logs from the running agent container.

Examples:
  azx agent logs              # follow all logs
  azx agent logs --tail 50    # last 50 lines then follow`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireDocker(); err != nil {
			return err
		}

		if !containerExists(agentName) {
			return fmt.Errorf("no agent container %q found\n\nStart it with: azx agent start", agentName)
		}

		dockerArgs := []string{"logs", "-f"}
		if agentLogsTail != "" {
			dockerArgs = append(dockerArgs, "--tail", agentLogsTail)
		}
		dockerArgs = append(dockerArgs, agentName)

		proc := exec.Command("docker", dockerArgs...)
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		return proc.Run()
	},
}

// --- Docker helpers ---

func requireDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("Docker is required but not found in PATH\n\nInstall Docker: https://docs.docker.com/get-docker/")
	}
	out, err := exec.Command("docker", "info").CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "Cannot connect") || strings.Contains(string(out), "Is the docker daemon running") {
			return fmt.Errorf("Docker is installed but not running\n\nStart Docker Desktop or run: sudo systemctl start docker")
		}
		return fmt.Errorf("Docker check failed: %w\n%s", err, string(out))
	}
	return nil
}

func isContainerRunning(name string) bool {
	out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func containerExists(name string) bool {
	err := exec.Command("docker", "inspect", name).Run()
	return err == nil
}

func removeStoppedContainer(name string) {
	if containerExists(name) && !isContainerRunning(name) {
		exec.Command("docker", "rm", name).Run()
	}
}

func init() {
	agentConfigCmd.Flags().StringVar(&agentConfig, "config", defaultConfigFile, "Output config file path")

	agentStartCmd.Flags().StringVar(&agentImage, "image", defaultImage, "Docker image (e.g., authzx/agent:v0.1.3)")
	agentStartCmd.Flags().StringVar(&agentName, "name", defaultContainerName, "Container name")
	agentStartCmd.Flags().StringVar(&agentPort, "port", defaultPort, "Host port to map to agent")
	agentStartCmd.Flags().StringVar(&agentConfig, "config", defaultConfigFile, "Agent config file to mount")
	agentStartCmd.Flags().BoolVar(&agentForeground, "foreground", false, "Run in foreground (not detached)")
	agentStartCmd.Flags().StringArrayVarP(&agentEnvVars, "env", "e", nil, "Environment variables (repeatable, e.g., -e AUTHZX_LOG_LEVEL=debug)")

	agentStatusCmd.Flags().StringVar(&agentPort, "port", defaultPort, "Agent port")

	agentStopCmd.Flags().StringVar(&agentName, "name", defaultContainerName, "Container name")

	agentLogsCmd.Flags().StringVar(&agentName, "name", defaultContainerName, "Container name")
	agentLogsCmd.Flags().StringVar(&agentLogsTail, "tail", "", "Number of lines to show from the end")

	agentCmd.AddCommand(agentConfigCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentLogsCmd)
}
