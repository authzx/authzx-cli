package credentials

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAPIKey(t *testing.T) {
	cases := []struct {
		name    string
		key     string
		wantErr bool
		errHas  string
	}{
		{"empty", "", true, "empty"},
		{"whitespace", "   ", true, "empty"},
		{"valid", "azx_abc123def456", false, ""},
		{"wrong prefix", "sk_abc123", true, "expected prefix"},
		{"client secret rejected", "azx_cs_abc123", true, "OAuth client secret"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAPIKey(tc.key)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.errHas != "" && err != nil && !strings.Contains(err.Error(), tc.errHas) {
				t.Fatalf("error %q missing substring %q", err.Error(), tc.errHas)
			}
		})
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &Config{APIKey: "azx_test_roundtrip_key"}
	if err := Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify permissions on file and dir.
	dInfo, err := os.Stat(filepath.Join(tmp, ".authzx"))
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if dInfo.Mode().Perm() != 0700 {
		t.Errorf("dir perm = %o, want 0700", dInfo.Mode().Perm())
	}
	fInfo, err := os.Stat(filepath.Join(tmp, ".authzx", "config.yaml"))
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if fInfo.Mode().Perm() != 0600 {
		t.Errorf("file perm = %o, want 0600", fInfo.Mode().Perm())
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", got.APIKey, cfg.APIKey)
	}
}

func TestRemoveIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Remove when file doesn't exist — should be nil.
	if err := Remove(); err != nil {
		t.Fatalf("remove (absent): %v", err)
	}

	// Create, then remove.
	if err := Save(&Config{APIKey: "azx_remove_test"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := Remove(); err != nil {
		t.Fatalf("remove (present): %v", err)
	}

	// Remove again — still nil.
	if err := Remove(); err != nil {
		t.Fatalf("remove (second time): %v", err)
	}
}

func TestResolvePrecedence(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("AUTHZX_API_KEY", "")

	// No key anywhere.
	if _, err := Resolve(""); err == nil {
		t.Fatal("expected ErrNotAuthenticated, got nil")
	}

	// Config file only.
	if err := Save(&Config{APIKey: "azx_from_file"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Resolve("")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "azx_from_file" {
		t.Errorf("got %q, want from file", got)
	}

	// Env var takes precedence over file.
	t.Setenv("AUTHZX_API_KEY", "azx_from_env")
	got, err = Resolve("")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "azx_from_env" {
		t.Errorf("got %q, want from env", got)
	}

	// Flag beats everything.
	got, err = Resolve("azx_from_flag")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "azx_from_flag" {
		t.Errorf("got %q, want from flag", got)
	}
}

func TestEndpoint(t *testing.T) {
	t.Setenv("AUTHZX_ENDPOINT", "")
	if got := Endpoint(); got != DefaultEndpoint {
		t.Errorf("default endpoint: got %q, want %q", got, DefaultEndpoint)
	}

	t.Setenv("AUTHZX_ENDPOINT", "https://staging.authzx.com/")
	if got := Endpoint(); got != "https://staging.authzx.com" {
		t.Errorf("override endpoint: got %q", got)
	}
}
