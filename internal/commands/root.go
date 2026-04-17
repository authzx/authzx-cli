package commands

import (
	"github.com/spf13/cobra"
)

// rootAPIKey is populated from the persistent --api-key flag and consumed by
// commands via credentials.Resolve(). It is package-level state because
// cobra binds persistent flags at init time, before commands run.
var rootAPIKey string

var rootCmd = &cobra.Command{
	Use:   "authzx",
	Short: "AuthzX CLI — authorization management and evaluation",
	Long: `AuthzX command-line interface.

Get started:
  authzx configure         # paste your API key
  authzx check ...         # run an authorization check`,
	SilenceErrors: true,
}

// Execute runs the CLI; main() wraps it and prints errors to stderr.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootAPIKey, "api-key", "", "AuthzX API key (overrides AUTHZX_API_KEY and ~/.authzx/config.yaml)")

	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(checkCmd)
	// TODO: add `agent` command back once scope is locked. Planned v2 surface:
	//   - authzx agent config   → write a starter YAML with API key prefilled
	//   - authzx agent status   → hit /healthz on the running agent
	// Intentionally omitted: start/stop/pull — Docker does those.
	// rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(versionCmd)
}
