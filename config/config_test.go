// ABOUTME: Tests for config loading, saving, and manipulation.
// ABOUTME: Verifies JSON config file operations for crate-to-repo mappings.

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
	if cfg.Crates == nil {
		t.Error("Load() Crates is nil, want empty map")
	}
	if len(cfg.Crates) != 0 {
		t.Errorf("Load() Crates has %d entries, want 0", len(cfg.Crates))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Crates: map[string]string{
			"alacritty_terminal": "alacritty/alacritty",
			"serde":              "serde-rs/serde",
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Crates) != 2 {
		t.Errorf("Load() Crates has %d entries, want 2", len(loaded.Crates))
	}
	if loaded.Crates["alacritty_terminal"] != "alacritty/alacritty" {
		t.Errorf("Load() Crates[alacritty_terminal] = %q, want %q", loaded.Crates["alacritty_terminal"], "alacritty/alacritty")
	}
	if loaded.Crates["serde"] != "serde-rs/serde" {
		t.Errorf("Load() Crates[serde] = %q, want %q", loaded.Crates["serde"], "serde-rs/serde")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	cfg := &Config{Crates: map[string]string{"test": "owner/repo"}}

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
