package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	authzx "github.com/authzx/authzx-go"
	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
)

var (
	checkSubject  string
	checkAction   string
	checkResource string
	checkRoles    string
	checkContext  string
	checkLocal    bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run an authorization check",
	Long: `Run an authorization check against cloud or local agent.

Examples:
  authzx check --subject 72a9b754-5034-45f7-bfca-5e6b47943a23 --action read --resource 2abebabe-74f8-4bb7-b14b-f9500b7c496d
  authzx check --subject user:123 --action read --resource document:456
  authzx check --subject user:123 --action read --resource document:456 --roles editor,viewer --local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse subject — accepts "id" or "type:id"
		var subject authzx.Subject
		if parts := strings.SplitN(checkSubject, ":", 2); len(parts) == 2 {
			subject = authzx.Subject{ID: parts[1], Type: parts[0]}
		} else {
			subject = authzx.Subject{ID: checkSubject}
		}
		if checkRoles != "" {
			subject.Roles = strings.Split(checkRoles, ",")
		}

		// Parse resource — accepts "id" or "type:id"
		var resource authzx.Resource
		if parts := strings.SplitN(checkResource, ":", 2); len(parts) == 2 {
			resource = authzx.Resource{ID: parts[1], Type: parts[0]}
		} else {
			resource = authzx.Resource{ID: checkResource}
		}

		var ctxMap map[string]interface{}
		if checkContext != "" {
			if err := json.Unmarshal([]byte(checkContext), &ctxMap); err != nil {
				return fmt.Errorf("invalid --context JSON: %w", err)
			}
		}

		var client *authzx.Client
		if checkLocal {
			client = authzx.NewClient("", authzx.WithBaseURL("http://localhost:8181"))
		} else {
			creds, err := credentials.Load()
			if err != nil {
				return err
			}
			client = authzx.NewClient(creds.APIKey, authzx.WithBaseURL(creds.CloudURL))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.Authorize(ctx, &authzx.AuthorizeRequest{
			Subject:  subject,
			Resource: resource,
			Action:   checkAction,
			Context:  ctxMap,
		})
		if err != nil {
			return fmt.Errorf("authorization failed: %w", err)
		}

		if resp.Allowed {
			fmt.Println("\033[32mALLOWED\033[0m")
		} else {
			fmt.Println("\033[31mDENIED\033[0m")
		}
		fmt.Printf("  Reason: %s\n", resp.Reason)
		if resp.PolicyID != "" {
			fmt.Printf("  Policy: %s\n", resp.PolicyID)
		}
		if resp.AccessPath != "" {
			fmt.Printf("  Path:   %s\n", resp.AccessPath)
		}

		return nil
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkSubject, "subject", "", "Subject ID or type:id")
	checkCmd.Flags().StringVar(&checkAction, "action", "", "Action to check")
	checkCmd.Flags().StringVar(&checkResource, "resource", "", "Resource ID or type:id")
	checkCmd.Flags().StringVar(&checkRoles, "roles", "", "Comma-separated roles")
	checkCmd.Flags().StringVar(&checkContext, "context", "", "JSON context object")
	checkCmd.Flags().BoolVar(&checkLocal, "local", false, "Use local agent (localhost:8181)")
	checkCmd.MarkFlagRequired("subject")
	checkCmd.MarkFlagRequired("action")
	checkCmd.MarkFlagRequired("resource")
}
