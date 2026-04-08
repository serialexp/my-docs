// ABOUTME: Manages local git clones for repo search and file access.
// ABOUTME: Handles on-demand cloning, staleness checks, and local ripgrep search.

package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bartriepe/my-docs/ripgrep"
)

const pullStaleness = 24 * time.Hour

// reposDirFn is a package-level var to allow test injection.
var reposDirFn = reposDir

func reposDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "my-docs", "repos")
}

func clonePath(repo string) string {
	base := reposDirFn()
	if base == "" {
		return ""
	}
	return filepath.Join(base, repo)
}

func cloneTmpPath(repo string) string {
	base := reposDirFn()
	if base == "" {
		return ""
	}
	return filepath.Join(base, repo+".tmp")
}

func cloneReady(repo string) bool {
	dir := clonePath(repo)
	if dir == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

func cloneInProgress(repo string) bool {
	tmp := cloneTmpPath(repo)
	if tmp == "" {
		return false
	}
	_, err := os.Stat(tmp)
	return err == nil
}

func readLocalFile(repo, path string) (string, error) {
	dir := clonePath(repo)
	if dir == "" {
		return "", fmt.Errorf("no repos directory")
	}
	data, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func lastPullPath(repo string) string {
	dir := clonePath(repo)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, ".last-pull")
}

func touchLastPull(repo string) {
	path := lastPullPath(repo)
	if path == "" {
		return
	}
	_ = os.WriteFile(path, nil, 0o644)
}

func isCloneStale(repo string) bool {
	path := lastPullPath(repo)
	if path == "" {
		return true
	}
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > pullStaleness
}

// pullIfStale does a synchronous git pull if the clone hasn't been updated recently.
// If pull fails (e.g. network down), we still serve from the stale clone.
func pullIfStale(repo string) {
	if !isCloneStale(repo) {
		return
	}

	dir := clonePath(repo)
	if dir == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "[my-docs] updating local clone of %s...\n", repo)
	cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		// Pull failed (network issue, etc.) — serve stale content
		fmt.Fprintf(os.Stderr, "[my-docs] pull failed, using cached clone: %v\n", err)
	}

	// Touch marker regardless — avoid retrying pull on every call during an outage
	touchLastPull(repo)
}

// EnsureClone clones the repo if it hasn't been cloned yet, or pulls if stale.
// This blocks until the clone is ready. Returns an error only if the clone fails.
func EnsureClone(repo string) error {
	if cloneReady(repo) {
		pullIfStale(repo)
		return nil
	}
	fmt.Fprintf(os.Stderr, "[my-docs] cloning %s (shallow)...\n", repo)
	return RunClone(repo)
}

// RunClone performs the actual git clone. Called by EnsureClone and the hidden _clone subcommand.
func RunClone(repo string) error {
	dir := clonePath(repo)
	tmp := cloneTmpPath(repo)

	if dir == "" || tmp == "" {
		return fmt.Errorf("cannot determine cache directory")
	}

	if cloneReady(repo) {
		return nil
	}

	// Clean up stale tmp from a previous failed attempt
	if cloneInProgress(repo) {
		os.RemoveAll(tmp)
	}

	if err := os.MkdirAll(filepath.Dir(tmp), 0o755); err != nil {
		return err
	}

	url := fmt.Sprintf("https://github.com/%s.git", repo)
	cmd := exec.Command("git", "clone", "--depth", "1", url, tmp)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("git clone failed: %w", err)
	}

	if err := os.Rename(tmp, dir); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("rename failed: %w", err)
	}

	touchLastPull(repo)
	return nil
}

// SearchMatch represents a single search result from a local ripgrep search.
type SearchMatch struct {
	Path string
	Line int
	Text string
}

// CanSearchLocal returns true if a local clone is available for the repo.
func CanSearchLocal(repo string) bool {
	return cloneReady(repo)
}

// SearchLocal searches a local clone using ripgrep. It pulls if stale before searching.
// Returns an error if rg cannot be obtained.
func SearchLocal(repo, pattern string) ([]SearchMatch, error) {
	pullIfStale(repo)

	rgBin, err := ripgrep.BinPath()
	if err != nil {
		return nil, fmt.Errorf("ripgrep not available: %w", err)
	}

	dir := clonePath(repo)
	if dir == "" {
		return nil, fmt.Errorf("no repos directory")
	}

	cmd := exec.Command(rgBin, "--no-heading", "-n", pattern, dir)
	output, err := cmd.Output()
	if err != nil {
		// rg exits with code 1 when there are no matches — that's not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep failed: %w", err)
	}

	return parseRipgrepOutput(string(output), dir), nil
}

// parseRipgrepOutput parses `rg --no-heading -n` output into SearchMatch slices.
// Each line has the format: <abs-path>:<line>:<text>
// The cloneDir prefix is stripped from paths.
func parseRipgrepOutput(output, cloneDir string) []SearchMatch {
	var matches []SearchMatch
	prefix := cloneDir + "/"

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}

		// Strip clone dir prefix
		rel := strings.TrimPrefix(line, prefix)

		// Split on first ":" to get path
		pathEnd := strings.Index(rel, ":")
		if pathEnd == -1 {
			continue
		}
		path := rel[:pathEnd]
		rest := rel[pathEnd+1:]

		// Split on next ":" to get line number and text
		lineEnd := strings.Index(rest, ":")
		if lineEnd == -1 {
			continue
		}
		lineNum, err := strconv.Atoi(rest[:lineEnd])
		if err != nil {
			continue
		}
		text := rest[lineEnd+1:]

		matches = append(matches, SearchMatch{
			Path: path,
			Line: lineNum,
			Text: text,
		})
	}

	return matches
}
