// ABOUTME: CLI entry point for my-docs.
// ABOUTME: Searches documentation across git repositories without cloning them locally.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartriepe/my-docs/cmd"
	"github.com/bartriepe/my-docs/config"
	"github.com/bartriepe/my-docs/cratesio"
	"github.com/bartriepe/my-docs/github"
	"github.com/bartriepe/my-docs/grepapp"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "find":
		runFind(args)
	case "search":
		runSearch(args)
	case "cat":
		runCat(args)
	case "rust":
		runRust(args)
	case "install":
		runInstall()
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`my-docs - Search documentation across git repositories

Usage:
  my-docs <command> [arguments]

Commands:
  search [owner/repo] <pattern>  Search repo via grep.app (omit repo to search all)
  cat <owner/repo> <path>        Fetch and display file from GitHub
  find <query>                   Search for repos by name
  rust <crate> <symbol>          Look up a Rust crate symbol and show its source
  install                        Install instructions into ~/.claude/CLAUDE.md`)
}

func loadConfig() *config.Config {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func saveConfig(cfg *config.Config) {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := config.Save(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
		os.Exit(1)
	}
}

func runFind(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: my-docs find <query>")
		os.Exit(1)
	}
	resp, err := grepapp.Search(args[0], "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Facets.Repo.Buckets) == 0 {
		fmt.Println("No repositories found")
		return
	}
	for _, bucket := range resp.Facets.Repo.Buckets {
		fmt.Printf("%s (%d matches)\n", bucket.Val, bucket.Count)
	}
}

func runSearch(args []string) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs search [owner/repo] <pattern>")
		os.Exit(1)
	}

	var repo, pattern string
	if len(args) == 1 {
		// No repo specified, search across all repos
		pattern = args[0]
		repo = ""
	} else {
		// Repo specified in owner/repo format
		repo = args[0]
		if !strings.Contains(repo, "/") {
			fmt.Fprintf(os.Stderr, "error: invalid repo format %q: must be owner/repo\n", repo)
			os.Exit(1)
		}
		pattern = args[1]
	}

	resp, err := grepapp.Search(pattern, repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Hits.Hits) == 0 {
		fmt.Println("No matches found")
		return
	}
	for _, hit := range resp.Hits.Hits {
		matches := grepapp.ExtractText(hit.Content.Snippet)
		for _, m := range matches {
			fmt.Printf("%s:%d: %s\n", hit.Path, m.Line, m.Text)
		}
	}
}

func runCat(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs cat <owner/repo> <path>")
		os.Exit(1)
	}
	repo := args[0]
	if !strings.Contains(repo, "/") {
		fmt.Fprintf(os.Stderr, "error: invalid repo format %q: must be owner/repo\n", repo)
		os.Exit(1)
	}
	content, err := github.FetchFile(repo, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(content)
}

func runRust(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs rust <crate> <symbol>")
		os.Exit(1)
	}
	crateName := args[0]
	symbol := args[1]

	cfg := loadConfig()

	// Check cache first
	repo, cached := cfg.Crates[crateName]
	if !cached {
		// Look up on crates.io
		resp, err := cratesio.Lookup(crateName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		repo, err = cratesio.ExtractGitHubRepo(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		// Cache the result
		cfg.Crates[crateName] = repo
		saveConfig(cfg)
	}

	// Search for the symbol in the repo
	resp, err := grepapp.Search(symbol, repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Hits.Hits) == 0 {
		fmt.Print(cmd.FormatNoMatches(symbol, crateName))
		os.Exit(1)
	}

	files := cmd.CollectMatchingFiles(resp.Hits.Hits)

	if len(files) == 1 {
		// Single file - fetch and output it
		content, err := github.FetchFile(repo, files[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(content)
	} else {
		// Multiple files - show cat commands
		fmt.Print(cmd.FormatMultipleMatches(symbol, repo, files))
	}
}

func runInstall() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	claudeMdPath := filepath.Join(home, ".claude", "CLAUDE.md")

	existing, err := os.ReadFile(claudeMdPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", claudeMdPath, err)
		os.Exit(1)
	}

	updated := cmd.UpdateClaudeMdSection(string(existing), cmd.Instructions)

	dir := filepath.Dir(claudeMdPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(claudeMdPath, []byte(updated), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", claudeMdPath, err)
		os.Exit(1)
	}

	fmt.Printf("Installed my-docs instructions to %s\n", claudeMdPath)
}
