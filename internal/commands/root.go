package commands

import (
	"github.com/spf13/cobra"
)

// rootAPIKey is populated from the persistent --api-key flag and consumed by
// commands via credentials.Resolve(). It is package-level state because
// cobra binds persistent flags at init time, before commands run.
var rootAPIKey string

var rootCmd = &cobra.Command{
	Use:   "azx",
	Short: "AuthzX CLI — authorization management and evaluation",
	Long: `AuthzX command-line interface.

Get started:
  azx configure         # paste your API key
  azx check ...         # run an authorization check`,
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
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(versionCmd)
}
