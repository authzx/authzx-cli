package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// readSecret prompts with the given label and reads a line from stdin with
// echo suppressed when stdin is a terminal. Falls back to a plain line read
// if stdin is piped (e.g. `echo $KEY | authzx configure`).
func readSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // newline after masked input
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Non-TTY: read one line, strip trailing newline.
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Store your AuthzX API key",
	Long: `Interactively store an AuthzX API key at ~/.authzx/config.yaml.

The prompt masks your input. Get a key from the AuthzX console
(https://console.authzx.com) under Settings → API Keys.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := readSecret("AuthzX API Key: ")
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey := strings.TrimSpace(raw)

		if err := credentials.ValidateAPIKey(apiKey); err != nil {
			return err
		}

		if err := credentials.Save(&credentials.Config{APIKey: apiKey}); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Saved to %s\n", credentials.Path())
		return nil
	},
}

