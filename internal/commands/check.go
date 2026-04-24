package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/authzx/authzx-cli/internal/credentials"
	authzx "github.com/authzx/authzx-go"
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
	Long: `Run an authorization check against the AuthzX cloud (or a local agent).

Examples:
  azx check --subject user-123 --resource doc-456 --action read
  azx check --subject user:123 --action read --resource document:456
  azx check --subject user:123 --action read --resource document:456 \
               --roles editor,viewer --context '{"ip":"10.0.0.1"}'
  azx check --subject user:123 --action read --resource document:456 --local`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse subject — accepts "id" or "type:id".
		var subject authzx.Subject
		if parts := strings.SplitN(checkSubject, ":", 2); len(parts) == 2 {
			subject = authzx.Subject{ID: parts[1], Type: parts[0]}
		} else {
			subject = authzx.Subject{ID: checkSubject}
		}
		if checkRoles != "" {
			subject.Roles = strings.Split(checkRoles, ",")
		}

		// Parse resource — accepts "id" or "type:id".
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
			apiKey, err := credentials.Resolve(rootAPIKey)
			if err != nil {
				return err
			}
			client = authzx.NewClient(apiKey, authzx.WithBaseURL(credentials.Endpoint()))
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

		out := cmd.OutOrStdout()
		verdict := "denied"
		if resp.Allowed {
			verdict = "allowed"
		}
		// Short form: single line that's grep/awk friendly.
		if resp.AccessPath != "" {
			fmt.Fprintf(out, "%s: %t  (%s)\n", verdict, resp.Allowed, resp.AccessPath)
		} else {
			fmt.Fprintf(out, "%s: %t\n", verdict, resp.Allowed)
		}
		if resp.Reason != "" {
			fmt.Fprintf(out, "  reason: %s\n", resp.Reason)
		}
		if resp.PolicyID != "" {
			fmt.Fprintf(out, "  policy: %s\n", resp.PolicyID)
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
	_ = checkCmd.MarkFlagRequired("subject")
	_ = checkCmd.MarkFlagRequired("action")
	_ = checkCmd.MarkFlagRequired("resource")
}
