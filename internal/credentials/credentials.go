// Package credentials manages the AuthzX CLI's local configuration file.
//
// The config lives at ~/.authzx/config.yaml with mode 0600. The directory is
// created (if missing) with mode 0700. Only the API key is stored; the
// endpoint is resolved at call time from the AUTHZX_ENDPOINT env var (falling
// back to the hardcoded production URL).
package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultEndpoint is the production AuthzX API base URL. Override at runtime
// with the AUTHZX_ENDPOINT env var (undocumented escape hatch for staging).
const DefaultEndpoint = "https://api.authzx.com"

// APIKeyPrefix is the expected prefix of a valid AuthzX CLI API key.
const APIKeyPrefix = "azx_"

// ClientSecretPrefix identifies OAuth client secrets, which are NOT valid
// CLI API keys — callers who paste one should get a targeted error.
const ClientSecretPrefix = "azx_cs_"

// Config is the on-disk shape of ~/.authzx/config.yaml.
type Config struct {
	APIKey string `yaml:"api_key"`
}

// Dir returns the absolute path to ~/.authzx.
func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".authzx")
}

// Path returns the absolute path to ~/.authzx/config.yaml.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Endpoint resolves the API base URL, honoring the AUTHZX_ENDPOINT env var.
func Endpoint() string {
	if v := strings.TrimSpace(os.Getenv("AUTHZX_ENDPOINT")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultEndpoint
}

// ValidateAPIKey enforces the shape rules for v1: must start with "azx_" and
// must NOT be an OAuth client secret ("azx_cs_"). Returns a customer-facing
// error message on failure.
func ValidateAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("API key cannot be empty")
	}
	if strings.HasPrefix(key, ClientSecretPrefix) {
		return errors.New("that looks like an OAuth client secret (azx_cs_...), not a CLI API key — generate an API key from the AuthzX console instead")
	}
	if !strings.HasPrefix(key, APIKeyPrefix) {
		return fmt.Errorf("invalid API key: expected prefix %q", APIKeyPrefix)
	}
	return nil
}

// Load reads ~/.authzx/config.yaml. Returns os.ErrNotExist (unwrapped via
// errors.Is) when the file is absent so callers can distinguish "no config"
// from "broken config".
func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", Path(), err)
	}
	return &cfg, nil
}

// Save writes cfg to ~/.authzx/config.yaml, creating the directory if needed.
// Applies 0700 to the directory and 0600 to the file.
func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

// Remove deletes ~/.authzx/config.yaml. Idempotent — returns nil if the file
// is already absent.
func Remove() error {
	err := os.Remove(Path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// MaskKey returns a user-safe preview of an API key, e.g. "azx_teTF...".
func MaskKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}
