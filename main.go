// ABOUTME: CLI entry point for my-docs.
// ABOUTME: Searches documentation across git repositories using local shallow clones and ripgrep.

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
	"github.com/bartriepe/my-docs/githubsearch"
)

var version = "dev"

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
	case "info":
		runInfo(args)
	case "rust":
		runRust(args)
	case "install":
		runInstall()
	case "_clone":
		runCloneCmd(args)
	case "version", "-v", "--version":
		fmt.Println(version)
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
  search [owner/repo] <pattern>  Search repo locally with ripgrep (clones on first use)
                                 Omit repo to search across GitHub with code search
    --limit N                    Max results to show (default: 15)
    --offset N                   Skip first N results (for pagination)
  cat <owner/repo> <path>        Fetch and display file from GitHub
    --lines N or N-M             Show only specific lines (e.g., --lines 10-50)
  info <owner/repo>              Show repo info (llms.txt or README.md)
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
	resp, err := githubsearch.SearchRepos(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Items) == 0 {
		fmt.Println("No repositories found")
		return
	}
	for _, r := range resp.Items {
		if r.Description != "" {
			fmt.Printf("%s (★ %d) - %s\n", r.FullName, r.Stars, r.Description)
		} else {
			fmt.Printf("%s (★ %d)\n", r.FullName, r.Stars)
		}
	}
}

func runSearch(args []string) {
	// Default pagination
	limit := 15
	offset := 0

	// Parse flags manually (before positional args)
	var positionalArgs []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--limit" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%d", &limit)
			i++
		case args[i] == "--offset" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%d", &offset)
			i++
		case strings.HasPrefix(args[i], "--limit="):
			fmt.Sscanf(args[i], "--limit=%d", &limit)
		case strings.HasPrefix(args[i], "--offset="):
			fmt.Sscanf(args[i], "--offset=%d", &offset)
		default:
			positionalArgs = append(positionalArgs, args[i])
		}
	}

	if len(positionalArgs) < 1 || len(positionalArgs) > 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs search [owner/repo] <pattern> [--limit N] [--offset N]")
		os.Exit(1)
	}

	var repo, pattern string
	if len(positionalArgs) == 1 {
		// No repo specified, search across all repos
		pattern = positionalArgs[0]
		repo = ""
	} else {
		// Repo specified in owner/repo format
		repo = positionalArgs[0]
		if !strings.Contains(repo, "/") {
			fmt.Fprintf(os.Stderr, "error: invalid repo format %q: must be owner/repo\n", repo)
			os.Exit(1)
		}
		pattern = positionalArgs[1]
	}

	if repo != "" {
		// Specific repo: clone locally and search with ripgrep (full regex support)
		if err := github.EnsureClone(repo); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		allMatches, err := github.SearchLocal(repo, pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(allMatches) == 0 {
			fmt.Println("No matches found")
			return
		}

		// Apply offset and limit
		total := len(allMatches)
		if offset >= total {
			fmt.Println("No more results")
			return
		}
		end := offset + limit
		if end > total {
			end = total
		}

		for _, m := range allMatches[offset:end] {
			fmt.Printf("%s:%d: %s\n", m.Path, m.Line, m.Text)
		}

		remaining := total - end
		if remaining > 0 {
			fmt.Printf("\n... and %d more results (use --offset %d to see next page)\n", remaining, end)
		}
	} else {
		// No repo specified: search across all repos via GitHub code search
		resp, err := githubsearch.Search(pattern, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		results := githubsearch.ExtractResults(resp)
		if len(results) == 0 {
			fmt.Println("No matches found")
			return
		}

		// Apply offset and limit
		total := len(results)
		if offset >= total {
			fmt.Println("No more results")
			return
		}
		end := offset + limit
		if end > total {
			end = total
		}

		for _, r := range results[offset:end] {
			if r.Fragment != "" {
				fmt.Printf("[%s] %s:\n", r.Repo, r.Path)
				for _, line := range strings.Split(r.Fragment, "\n") {
					fmt.Printf("  %s\n", line)
				}
			} else {
				fmt.Printf("[%s] %s\n", r.Repo, r.Path)
			}
		}

		remaining := total - end
		if remaining > 0 {
			fmt.Printf("\n... and %d more results (use --offset %d to see next page)\n", remaining, end)
		}
	}
}

func runCat(args []string) {
	var lineRange string
	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--lines" && i+1 < len(args):
			lineRange = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--lines="):
			lineRange = strings.TrimPrefix(args[i], "--lines=")
		default:
			positionalArgs = append(positionalArgs, args[i])
		}
	}

	if len(positionalArgs) != 2 {
		fmt.Fprintln(os.Stderr, "usage: my-docs cat <owner/repo> <path> [--lines START-END]")
		os.Exit(1)
	}
	repo := positionalArgs[0]
	if !strings.Contains(repo, "/") {
		fmt.Fprintf(os.Stderr, "error: invalid repo format %q: must be owner/repo\n", repo)
		os.Exit(1)
	}
	content, err := github.FetchFile(repo, positionalArgs[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if lineRange != "" {
		content, err = extractLines(content, lineRange)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(content)
}

func extractLines(content, lineRange string) (string, error) {
	var start, end int
	if n, _ := fmt.Sscanf(lineRange, "%d-%d", &start, &end); n == 2 {
		// range: 10-20
	} else if n, _ := fmt.Sscanf(lineRange, "%d", &start); n == 1 {
		// single line
		end = start
	} else {
		return "", fmt.Errorf("invalid line range %q: use N or N-M", lineRange)
	}

	if start < 1 {
		return "", fmt.Errorf("line numbers must be >= 1")
	}
	if end < start {
		return "", fmt.Errorf("end line %d must be >= start line %d", end, start)
	}

	lines := strings.Split(content, "\n")
	if start > len(lines) {
		return "", fmt.Errorf("start line %d exceeds file length (%d lines)", start, len(lines))
	}
	if end > len(lines) {
		end = len(lines)
	}

	selected := lines[start-1 : end]
	result := strings.Join(selected, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result, nil
}

func runInfo(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: my-docs info <owner/repo>")
		os.Exit(1)
	}
	repo := args[0]
	if !strings.Contains(repo, "/") {
		fmt.Fprintf(os.Stderr, "error: invalid repo format %q: must be owner/repo\n", repo)
		os.Exit(1)
	}
	content, err := github.FetchRepoInfo(repo)
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

	// Clone the repo and search locally for the symbol
	if err := github.EnsureClone(repo); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	matches, err := github.SearchLocal(repo, symbol)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Collect unique file paths from matches
	seen := make(map[string]bool)
	var files []string
	for _, m := range matches {
		if !seen[m.Path] {
			seen[m.Path] = true
			files = append(files, m.Path)
		}
	}

	if len(files) == 0 {
		fmt.Print(cmd.FormatNoMatches(symbol, crateName))
		os.Exit(1)
	}

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

func runCloneCmd(args []string) {
	if len(args) != 1 {
		os.Exit(1)
	}
	repo := args[0]
	if !strings.Contains(repo, "/") {
		os.Exit(1)
	}
	if err := github.RunClone(repo); err != nil {
		fmt.Fprintf(os.Stderr, "clone failed: %v\n", err)
		os.Exit(1)
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
