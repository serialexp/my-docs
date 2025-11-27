// ABOUTME: Tests for config loading, saving, and manipulation.
// ABOUTME: Verifies JSON config file operations and alias management.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Repos == nil {
		t.Error("Load() Repos is nil, want empty map")
	}
	if len(cfg.Repos) != 0 {
		t.Errorf("Load() Repos has %d entries, want 0", len(cfg.Repos))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Repos: map[string]string{
			"alloy": "grafana/alloy",
			"otel":  "open-telemetry/opentelemetry-collector",
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Repos) != 2 {
		t.Errorf("Load() Repos has %d entries, want 2", len(loaded.Repos))
	}
	if loaded.Repos["alloy"] != "grafana/alloy" {
		t.Errorf("Load() Repos[alloy] = %q, want %q", loaded.Repos["alloy"], "grafana/alloy")
	}
	if loaded.Repos["otel"] != "open-telemetry/opentelemetry-collector" {
		t.Errorf("Load() Repos[otel] = %q, want %q", loaded.Repos["otel"], "open-telemetry/opentelemetry-collector")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	cfg := &Config{Repos: map[string]string{"test": "owner/repo"}}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create config file")
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if path == "" {
		t.Error("DefaultPath() returned empty string")
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("DefaultPath() = %q, want config.json filename", path)
	}
}
