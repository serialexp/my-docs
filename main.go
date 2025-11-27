// ABOUTME: CLI entry point for my-docs.
// ABOUTME: Searches documentation across git repositories without cloning them locally.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/bartriepe/my-docs/cmd"
	"github.com/bartriepe/my-docs/config"
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
	case "list":
		runList()
	case "alias":
		runAlias(args)
	case "remove":
		runRemove(args)
	case "config":
		runConfig()
	case "find":
		runFind(args)
	case "search":
		runSearch(args)
	case "cat":
		runCat(args)
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
  search <repo> <pattern>    Search repo via grep.app
  cat <repo> <path>          Fetch and display file from GitHub
  find <query>               Search for repos by name
  list                       Show all configured repo aliases
  alias <name> <owner/repo>  Create alias for a repo
  remove <name>              Remove a repo alias
  config                     Show config file path
  install                    Install instructions into ~/.claude/CLAUDE.md`)
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

func runList() {
	cfg := loadConfig()
	entries := cmd.List(cfg)
	if len(entries) == 0 {
		fmt.Println("No repos configured. Use 'my-docs alias <name> <owner/repo>' to add one.")
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	for _, e := range entries {
		fmt.Printf("%s\t%s\n", e.Name, e.Repo)
	}
}

func runAlias(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs alias <name> <owner/repo>")
		os.Exit(1)
	}
	cfg := loadConfig()
	if err := cmd.Alias(cfg, args[0], args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	saveConfig(cfg)
	fmt.Printf("Aliased %s -> %s\n", args[0], args[1])
}

func runRemove(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: my-docs remove <name>")
		os.Exit(1)
	}
	cfg := loadConfig()
	if err := cmd.Remove(cfg, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	saveConfig(cfg)
	fmt.Printf("Removed alias %s\n", args[0])
}

func runConfig() {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(path)
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
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs search <repo> <pattern>")
		os.Exit(1)
	}
	cfg := loadConfig()
	repo, err := cmd.Resolve(cfg, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	resp, err := grepapp.Search(args[1], repo)
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
		fmt.Fprintln(os.Stderr, "usage: my-docs cat <repo> <path>")
		os.Exit(1)
	}
	cfg := loadConfig()
	repo, err := cmd.Resolve(cfg, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	content, err := github.FetchFile(repo, args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(content)
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
