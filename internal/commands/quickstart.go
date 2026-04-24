package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/authzx/authzx-cli/internal/credentials"
	"github.com/spf13/cobra"
)

// sampleData records the resources created by `quickstart` so they can be
// torn down later with `quickstart --cleanup`.
type sampleData struct {
	ApplicationID  string   `json:"application_id"`
	ResourceTypeID string   `json:"resource_type_id"`
	ResourceIDs    []string `json:"resource_ids"`
	SubjectIDs     []string `json:"subject_ids"`
	RoleIDs        []string `json:"role_ids"`
	PolicyIDs      []string `json:"policy_ids"`
}

func sampleDataPath() string {
	return filepath.Join(credentials.Dir(), "sample-data.json")
}

func saveSampleData(sd *sampleData) error {
	if err := os.MkdirAll(credentials.Dir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sampleDataPath(), data, 0600)
}

func loadSampleData() (*sampleData, error) {
	data, err := os.ReadFile(sampleDataPath())
	if err != nil {
		return nil, err
	}
	var sd sampleData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

var quickstartCleanup bool

var quickstartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "Create sample authorization model for testing",
	Long: `Creates a complete sample authorization model in your tenant:
  - 1 application (Sample App)
  - 1 resource type (Sample Document: read, write, delete)
  - 2 resources (Sample Engineering Wiki, Sample Product Roadmap)
  - 2 subjects (Sample Alice, Sample Bob)
  - 2 roles (Sample Editor, Sample Viewer)
  - 2 policies with assignments

After running, you can immediately test authorization checks.

Pass --cleanup to remove the sample data previously created by quickstart.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := credentials.Resolve(rootAPIKey)
		if err != nil {
			return err
		}

		c := &apiClient{
			apiKey:  apiKey,
			baseURL: credentials.Endpoint(),
			http:    &http.Client{Timeout: 30 * time.Second},
		}

		if quickstartCleanup {
			return runCleanup(c)
		}

		return runQuickstart(c)
	},
}

func runQuickstart(c *apiClient) error {
	fmt.Println("Creating sample authorization model...")
	fmt.Println()

	sd := &sampleData{}

	// 1. Create application
	app, err := c.post("/application-srv/v1/applications", map[string]interface{}{
		"name":        "Sample App",
		"description": "Sample application created by authzx quickstart",
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
			return fmt.Errorf("sample data already exists in this tenant.\n\nRun 'authzx quickstart --cleanup' to remove it, or delete the existing 'Sample App' and related resources from the console, then run 'authzx quickstart' again")
		}
		return fmt.Errorf("failed to create application: %w", err)
	}
	appID := app["id"].(string)
	sd.ApplicationID = appID
	_ = saveSampleData(sd)
	fmt.Printf("  Application:   Sample App (%s)\n", appID)

	// 2. Create resource type
	rt, err := c.post("/resource-srv/v1/resource-types", map[string]interface{}{
		"name":        "Sample Document",
		"description": "Sample resource type for documents",
		"default_actions": []map[string]string{
			{"name": "read", "description": "Read a document"},
			{"name": "write", "description": "Edit a document"},
			{"name": "delete", "description": "Delete a document"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create resource type: %w", err)
	}
	rtID := rt["id"].(string)
	sd.ResourceTypeID = rtID
	_ = saveSampleData(sd)
	fmt.Printf("  Resource Type: Sample Document — read, write, delete (%s)\n", rtID)

	// 3. Create resources
	r1, err := c.post("/resource-srv/v1/resources", map[string]interface{}{
		"name":           "Sample Engineering Wiki",
		"type":           rtID,
		"application_id": appID,
	})
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	r1ID := r1["id"].(string)
	sd.ResourceIDs = append(sd.ResourceIDs, r1ID)
	_ = saveSampleData(sd)

	r2, err := c.post("/resource-srv/v1/resources", map[string]interface{}{
		"name":           "Sample Product Roadmap",
		"type":           rtID,
		"application_id": appID,
	})
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	r2ID := r2["id"].(string)
	sd.ResourceIDs = append(sd.ResourceIDs, r2ID)
	_ = saveSampleData(sd)
	fmt.Printf("  Resources:     Sample Engineering Wiki (%s)\n", r1ID)
	fmt.Printf("                 Sample Product Roadmap (%s)\n", r2ID)

	// 4. Create subjects
	s1, err := c.post("/entity-srv/v1/entities", map[string]interface{}{
		"name":            "Sample Alice",
		"type":            "user",
		"external_id":     "sample-alice",
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create subject: %w", err)
	}
	s1ID := s1["id"].(string)
	sd.SubjectIDs = append(sd.SubjectIDs, s1ID)
	_ = saveSampleData(sd)

	s2, err := c.post("/entity-srv/v1/entities", map[string]interface{}{
		"name":            "Sample Bob",
		"type":            "user",
		"external_id":     "sample-bob",
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create subject: %w", err)
	}
	s2ID := s2["id"].(string)
	sd.SubjectIDs = append(sd.SubjectIDs, s2ID)
	_ = saveSampleData(sd)
	fmt.Printf("  Subjects:      Sample Alice (%s)\n", s1ID)
	fmt.Printf("                 Sample Bob (%s)\n", s2ID)

	// 5. Create roles
	role1, err := c.post("/policy-srv/v1/roles", map[string]interface{}{
		"name":            "Sample Editor",
		"description":     "Can read and write documents",
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	role1ID := role1["id"].(string)
	sd.RoleIDs = append(sd.RoleIDs, role1ID)
	_ = saveSampleData(sd)

	role2, err := c.post("/policy-srv/v1/roles", map[string]interface{}{
		"name":            "Sample Viewer",
		"description":     "Can only read documents",
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	role2ID := role2["id"].(string)
	sd.RoleIDs = append(sd.RoleIDs, role2ID)
	_ = saveSampleData(sd)
	fmt.Printf("  Roles:         Sample Editor (%s)\n", role1ID)
	fmt.Printf("                 Sample Viewer (%s)\n", role2ID)

	// 6. Create policies
	p1, err := c.post("/policy-srv/v1/policies", map[string]interface{}{
		"name":        "Sample Editors Read Write",
		"description": "Editors can read and write documents",
		"effect":      "ALLOW",
		"resources": []map[string]interface{}{
			{"resource_id": r1ID, "actions": []string{"read", "write"}},
			{"resource_id": r2ID, "actions": []string{"read", "write"}},
		},
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}
	p1ID := p1["id"].(string)
	sd.PolicyIDs = append(sd.PolicyIDs, p1ID)
	_ = saveSampleData(sd)

	p2, err := c.post("/policy-srv/v1/policies", map[string]interface{}{
		"name":        "Sample Viewers Read Only",
		"description": "Viewers can only read documents",
		"effect":      "ALLOW",
		"resources": []map[string]interface{}{
			{"resource_id": r1ID, "actions": []string{"read"}},
			{"resource_id": r2ID, "actions": []string{"read"}},
		},
		"application_ids": []string{appID},
	})
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}
	p2ID := p2["id"].(string)
	sd.PolicyIDs = append(sd.PolicyIDs, p2ID)
	_ = saveSampleData(sd)
	fmt.Printf("  Policies:      Sample Editors Read Write (%s)\n", p1ID)
	fmt.Printf("                 Sample Viewers Read Only (%s)\n", p2ID)

	// 7. Assign policies to roles
	_, err = c.put("/policy-srv/v1/policies/assign", map[string]interface{}{
		"policy_ids":  []string{p1ID},
		"entity_type": "role",
		"entity_id":   role1ID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign policy to role: %w", err)
	}

	_, err = c.put("/policy-srv/v1/policies/assign", map[string]interface{}{
		"policy_ids":  []string{p2ID},
		"entity_type": "role",
		"entity_id":   role2ID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign policy to role: %w", err)
	}

	// 8. Assign roles to subjects
	_, err = c.post("/entity-srv/v1/entities/"+s1ID+"/roles", map[string]interface{}{
		"role_id": role1ID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign role to subject: %w", err)
	}

	_, err = c.post("/entity-srv/v1/entities/"+s2ID+"/roles", map[string]interface{}{
		"role_id": role2ID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign role to subject: %w", err)
	}

	fmt.Println()
	fmt.Println("Sample data created successfully!")
	fmt.Println()
	fmt.Println("Try it:")
	fmt.Printf("  azx check --subject %s --action read --resource %s\n", s1ID, r1ID)
	fmt.Printf("    → Alice (editor) can read Engineering Wiki — \033[32mALLOWED\033[0m\n\n")
	fmt.Printf("  azx check --subject %s --action write --resource %s\n", s2ID, r1ID)
	fmt.Printf("    → Bob (viewer) cannot write Engineering Wiki — \033[31mDENIED\033[0m\n\n")
	fmt.Printf("  azx check --subject %s --action delete --resource %s\n", s1ID, r1ID)
	fmt.Printf("    → Alice (editor) cannot delete — \033[31mDENIED\033[0m\n\n")
	fmt.Println("When you're done, clean up with: azx quickstart --cleanup")

	return nil
}

func runCleanup(c *apiClient) error {
	sd, err := loadSampleData()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No sample data found to clean up. Run 'authzx quickstart' to create sample data first.")
			return nil
		}
		return fmt.Errorf("failed to read sample data: %w", err)
	}

	fmt.Println("Cleaning up sample data...")
	fmt.Println()

	// Deletion order (reverse of creation):
	// policies (removes assignments) → subjects → roles → resources → resource type → application
	for _, id := range sd.PolicyIDs {
		if err := c.delete("/policy-srv/v1/policies/" + id); err != nil {
			fmt.Printf("  Policy %s: failed — %v\n", id, err)
		} else {
			fmt.Printf("  Policy %s: deleted\n", id)
		}
	}

	for _, id := range sd.SubjectIDs {
		if err := c.delete("/entity-srv/v1/entities/" + id); err != nil {
			fmt.Printf("  Subject %s: failed — %v\n", id, err)
		} else {
			fmt.Printf("  Subject %s: deleted\n", id)
		}
	}

	for _, id := range sd.RoleIDs {
		if err := c.delete("/policy-srv/v1/roles/" + id); err != nil {
			fmt.Printf("  Role %s: failed — %v\n", id, err)
		} else {
			fmt.Printf("  Role %s: deleted\n", id)
		}
	}

	for _, id := range sd.ResourceIDs {
		if err := c.delete("/resource-srv/v1/resources/" + id); err != nil {
			fmt.Printf("  Resource %s: failed — %v\n", id, err)
		} else {
			fmt.Printf("  Resource %s: deleted\n", id)
		}
	}

	if sd.ResourceTypeID != "" {
		if err := c.delete("/resource-srv/v1/resource-types/" + sd.ResourceTypeID); err != nil {
			fmt.Printf("  Resource Type %s: failed — %v\n", sd.ResourceTypeID, err)
		} else {
			fmt.Printf("  Resource Type %s: deleted\n", sd.ResourceTypeID)
		}
	}

	if sd.ApplicationID != "" {
		if err := c.delete("/application-srv/v1/applications/" + sd.ApplicationID); err != nil {
			fmt.Printf("  Application %s: failed — %v\n", sd.ApplicationID, err)
		} else {
			fmt.Printf("  Application %s: deleted\n", sd.ApplicationID)
		}
	}

	if err := os.Remove(sampleDataPath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("\nWarning: failed to remove %s: %v\n", sampleDataPath(), err)
	} else {
		fmt.Printf("\nSample data cleaned up. Removed %s\n", sampleDataPath())
	}

	return nil
}

type apiClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func (c *apiClient) do(method, path string, body map[string]interface{}) (map[string]interface{}, error) {
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if len(respBody) == 0 {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func (c *apiClient) post(path string, body map[string]interface{}) (map[string]interface{}, error) {
	return c.do("POST", path, body)
}

func (c *apiClient) put(path string, body map[string]interface{}) (map[string]interface{}, error) {
	return c.do("PUT", path, body)
}

func (c *apiClient) delete(path string) error {
	_, err := c.do("DELETE", path, nil)
	return err
}

func init() {
	quickstartCmd.Flags().BoolVar(&quickstartCleanup, "cleanup", false, "Delete sample data previously created by quickstart")
	rootCmd.AddCommand(quickstartCmd)
}
