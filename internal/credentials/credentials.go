package credentials

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Credentials struct {
	APIKey   string `json:"api_key"`
	CloudURL string `json:"cloud_url"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".authzx")
}

func Path() string {
	return filepath.Join(Dir(), "credentials")
}

func Load() (*Credentials, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("not logged in — run 'authzx login' first")
		}
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func Save(creds *Credentials) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

func Remove() error {
	err := os.Remove(Path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
