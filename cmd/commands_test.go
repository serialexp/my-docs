// ABOUTME: Tests for CLI command logic.
// ABOUTME: Verifies list, alias, remove, and config commands work correctly.

package cmd

import (
	"testing"

	"github.com/bartriepe/my-docs/config"
)

func TestList_Empty(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	result := List(cfg)

	if len(result) != 0 {
		t.Errorf("List() returned %d items, want 0", len(result))
	}
}

func TestList_WithRepos(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{
		"alloy": "grafana/alloy",
		"otel":  "open-telemetry/opentelemetry-collector",
	}}

	result := List(cfg)

	if len(result) != 2 {
		t.Errorf("List() returned %d items, want 2", len(result))
	}
}

func TestAlias_New(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	err := Alias(cfg, "alloy", "grafana/alloy")

	if err != nil {
		t.Fatalf("Alias() error = %v", err)
	}
	if cfg.Repos["alloy"] != "grafana/alloy" {
		t.Errorf("Alias() did not set repo, got %q", cfg.Repos["alloy"])
	}
}

func TestAlias_Overwrite(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{
		"alloy": "old/repo",
	}}

	err := Alias(cfg, "alloy", "grafana/alloy")

	if err != nil {
		t.Fatalf("Alias() error = %v", err)
	}
	if cfg.Repos["alloy"] != "grafana/alloy" {
		t.Errorf("Alias() did not update repo, got %q", cfg.Repos["alloy"])
	}
}

func TestAlias_InvalidRepo(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	err := Alias(cfg, "test", "invalid-no-slash")

	if err == nil {
		t.Error("Alias() error = nil, want error for invalid repo format")
	}
}

func TestRemove_Exists(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{
		"alloy": "grafana/alloy",
	}}

	err := Remove(cfg, "alloy")

	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, exists := cfg.Repos["alloy"]; exists {
		t.Error("Remove() did not delete the alias")
	}
}

func TestRemove_NotExists(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	err := Remove(cfg, "nonexistent")

	if err == nil {
		t.Error("Remove() error = nil, want error for nonexistent alias")
	}
}

func TestResolve_Alias(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{
		"alloy": "grafana/alloy",
	}}

	repo, err := Resolve(cfg, "alloy")

	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if repo != "grafana/alloy" {
		t.Errorf("Resolve() = %q, want %q", repo, "grafana/alloy")
	}
}

func TestResolve_FullRepo(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	repo, err := Resolve(cfg, "grafana/alloy")

	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if repo != "grafana/alloy" {
		t.Errorf("Resolve() = %q, want %q", repo, "grafana/alloy")
	}
}

func TestResolve_NotFound(t *testing.T) {
	cfg := &config.Config{Repos: map[string]string{}}

	_, err := Resolve(cfg, "nonexistent")

	if err == nil {
		t.Error("Resolve() error = nil, want error for unknown alias")
	}
}
