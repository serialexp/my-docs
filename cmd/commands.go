// ABOUTME: Core command logic for the CLI.
// ABOUTME: Implements list, alias, remove, and resolve operations on config.

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bartriepe/my-docs/config"
)

type RepoEntry struct {
	Name string
	Repo string
}

func List(cfg *config.Config) []RepoEntry {
	entries := make([]RepoEntry, 0, len(cfg.Repos))
	for name, repo := range cfg.Repos {
		entries = append(entries, RepoEntry{Name: name, Repo: repo})
	}
	return entries
}

func Alias(cfg *config.Config, name, repo string) error {
	if !strings.Contains(repo, "/") {
		return fmt.Errorf("invalid repo format %q: must be owner/repo", repo)
	}
	cfg.Repos[name] = repo
	return nil
}

func Remove(cfg *config.Config, name string) error {
	if _, exists := cfg.Repos[name]; !exists {
		return fmt.Errorf("alias %q not found", name)
	}
	delete(cfg.Repos, name)
	return nil
}

func Resolve(cfg *config.Config, nameOrRepo string) (string, error) {
	if repo, exists := cfg.Repos[nameOrRepo]; exists {
		return repo, nil
	}
	if strings.Contains(nameOrRepo, "/") {
		return nameOrRepo, nil
	}
	return "", errors.New("unknown alias: " + nameOrRepo)
}
