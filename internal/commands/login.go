package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store API key for AuthzX cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("API Key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		cloudURL := "https://api.authzx.com"
		fmt.Printf("Cloud URL [%s]: ", cloudURL)
		urlInput, _ := reader.ReadString('\n')
		urlInput = strings.TrimSpace(urlInput)
		if urlInput != "" {
			cloudURL = urlInput
		}

		creds := &credentials.Credentials{
			APIKey:   apiKey,
			CloudURL: cloudURL,
		}

		if err := credentials.Save(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Println("Credentials saved.")
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := credentials.Remove(); err != nil {
			return fmt.Errorf("failed to remove credentials: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}
