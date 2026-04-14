package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "authzx",
	Short: "AuthzX CLI — authorization management and evaluation",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(policiesCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(versionCmd)
}
