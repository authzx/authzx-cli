package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
)

var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Manage policies",
}

var policiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List policies for your tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := credentials.Load()
		if err != nil {
			return err
		}

		url := creds.CloudURL + "/policy-srv/v1/policies"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+creds.APIKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var result interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println(string(body))
			return nil
		}

		pretty, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(pretty))
		return nil
	},
}

var policiesGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get policy details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := credentials.Load()
		if err != nil {
			return err
		}

		url := creds.CloudURL + "/policy-srv/v1/policies/" + args[0]
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+creds.APIKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		var result interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println(string(body))
			return nil
		}

		pretty, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(pretty))
		return nil
	},
}

func init() {
	policiesCmd.AddCommand(policiesListCmd)
	policiesCmd.AddCommand(policiesGetCmd)
}
