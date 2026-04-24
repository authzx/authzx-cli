package credentials

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrNotAuthenticated is returned when no API key can be resolved from any
// source. Commands should surface its message verbatim.
var ErrNotAuthenticated = errors.New("not authenticated. Run 'azx configure' or set AUTHZX_API_KEY")

// Resolve returns the API key to use for the current command, applying the
// v1 precedence order:
//
//  1. explicit --api-key flag (flagValue)
//  2. AUTHZX_API_KEY environment variable
//  3. api_key in ~/.authzx/config.yaml
//
// Returns ErrNotAuthenticated if no source yields a key.
func Resolve(flagValue string) (string, error) {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(os.Getenv("AUTHZX_API_KEY")); v != "" {
		return v, nil
	}
	cfg, err := Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNotAuthenticated
		}
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return "", ErrNotAuthenticated
	}
	return cfg.APIKey, nil
}
