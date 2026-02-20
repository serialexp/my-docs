// ABOUTME: Manages automatic local git clones for frequently accessed repos.
// ABOUTME: Handles access counting, clone detection, background clone spawning, and local file reads.

package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const cloneThreshold = 5
const pullStaleness = 24 * time.Hour

// reposDirFn and countsDirFn are package-level vars to allow test injection.
var reposDirFn = reposDir
var countsDirFn = countsDir

func reposDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "my-docs", "repos")
}

func countsDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "my-docs", "counts")
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

func countPath(repo string) string {
	base := countsDirFn()
	if base == "" {
		return ""
	}
	return filepath.Join(base, repo)
}

func incrementCount(repo string) int {
	path := countPath(repo)
	if path == "" {
		return 0
	}

	count := 0
	data, err := os.ReadFile(path)
	if err == nil {
		count, _ = strconv.Atoi(string(data))
	}

	count++

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0
	}
	_ = os.WriteFile(path, []byte(strconv.Itoa(count)), 0o644)

	return count
}

func maybeTriggerClone(repo string) {
	if cloneReady(repo) || cloneInProgress(repo) {
		return
	}
	count := incrementCount(repo)
	if count >= cloneThreshold {
		fmt.Fprintf(os.Stderr, "[my-docs] cloning %s locally for faster access...\n", repo)
		spawnClone(repo)
	}
}

func spawnClone(repo string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	cmd := exec.Command(exe, "_clone", repo)
	cmd.SysProcAttr = detachedProcAttr()
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	_ = cmd.Start()
}

// RunClone performs the actual git clone. Called by the hidden _clone subcommand.
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
