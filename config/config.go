// ABOUTME: Manages the JSON config file for crate-to-repo mappings.
// ABOUTME: Handles loading, saving, and locating the config file.

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Crates map[string]string `json:"crates,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{
			Crates: make(map[string]string),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Crates == nil {
		cfg.Crates = make(map[string]string)
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "my-docs", "config.json"), nil
}
