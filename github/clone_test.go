// ABOUTME: Tests for automatic local git clone management.
// ABOUTME: Covers access counting, clone detection, local file reads, and path construction.

package github

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTestDirs(t *testing.T) (reposBase, countsBase string, cleanup func()) {
	t.Helper()
	reposBase = t.TempDir()
	countsBase = t.TempDir()

	origRepos := reposDirFn
	origCounts := countsDirFn
	reposDirFn = func() string { return reposBase }
	countsDirFn = func() string { return countsBase }

	return reposBase, countsBase, func() {
		reposDirFn = origRepos
		countsDirFn = origCounts
	}
}

func TestClonePath(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	got := clonePath("grafana/alloy")
	want := filepath.Join(reposBase, "grafana/alloy")
	if got != want {
		t.Errorf("clonePath() = %q, want %q", got, want)
	}
}

func TestCloneTmpPath(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	got := cloneTmpPath("grafana/alloy")
	want := filepath.Join(reposBase, "grafana/alloy.tmp")
	if got != want {
		t.Errorf("cloneTmpPath() = %q, want %q", got, want)
	}
}

func TestCloneReady_NoDir(t *testing.T) {
	_, _, cleanup := withTestDirs(t)
	defer cleanup()

	if cloneReady("grafana/alloy") {
		t.Error("cloneReady() returned true for non-existent directory")
	}
}

func TestCloneReady_WithGitDir(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	gitDir := filepath.Join(reposBase, "grafana", "alloy", ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !cloneReady("grafana/alloy") {
		t.Error("cloneReady() returned false when .git directory exists")
	}
}

func TestCloneInProgress_TmpExists(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	tmpDir := filepath.Join(reposBase, "grafana", "alloy.tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !cloneInProgress("grafana/alloy") {
		t.Error("cloneInProgress() returned false when .tmp directory exists")
	}
}

func TestCloneInProgress_NoTmp(t *testing.T) {
	_, _, cleanup := withTestDirs(t)
	defer cleanup()

	if cloneInProgress("grafana/alloy") {
		t.Error("cloneInProgress() returned true for non-existent .tmp directory")
	}
}

func TestIncrementCount_NewRepo(t *testing.T) {
	_, countsBase, cleanup := withTestDirs(t)
	defer cleanup()

	got := incrementCount("grafana/alloy")
	if got != 1 {
		t.Errorf("incrementCount() = %d, want 1", got)
	}

	data, err := os.ReadFile(filepath.Join(countsBase, "grafana", "alloy"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1" {
		t.Errorf("count file contains %q, want %q", string(data), "1")
	}
}

func TestIncrementCount_ExistingCount(t *testing.T) {
	_, countsBase, cleanup := withTestDirs(t)
	defer cleanup()

	path := filepath.Join(countsBase, "grafana", "alloy")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("3"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := incrementCount("grafana/alloy")
	if got != 4 {
		t.Errorf("incrementCount() = %d, want 4", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "4" {
		t.Errorf("count file contains %q, want %q", string(data), "4")
	}
}

func TestIncrementCount_CorruptFile(t *testing.T) {
	_, countsBase, cleanup := withTestDirs(t)
	defer cleanup()

	path := filepath.Join(countsBase, "grafana", "alloy")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := incrementCount("grafana/alloy")
	if got != 1 {
		t.Errorf("incrementCount() = %d, want 1 (corrupt file treated as 0)", got)
	}
}

func TestReadLocalFile_Success(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	filePath := filepath.Join(reposBase, "owner", "repo", "docs", "readme.md")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Hello\nThis is a test."
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readLocalFile("owner/repo", "docs/readme.md")
	if err != nil {
		t.Fatalf("readLocalFile() error = %v", err)
	}
	if got != content {
		t.Errorf("readLocalFile() = %q, want %q", got, content)
	}
}

func TestReadLocalFile_NotFound(t *testing.T) {
	_, _, cleanup := withTestDirs(t)
	defer cleanup()

	_, err := readLocalFile("owner/repo", "nonexistent.md")
	if err == nil {
		t.Error("readLocalFile() returned nil error for non-existent file")
	}
}

func TestIsCloneStale_NoMarker(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	// Create a clone dir without .last-pull
	gitDir := filepath.Join(reposBase, "owner", "repo", ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !isCloneStale("owner/repo") {
		t.Error("isCloneStale() returned false when .last-pull does not exist")
	}
}

func TestIsCloneStale_FreshMarker(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	cloneDir := filepath.Join(reposBase, "owner", "repo")
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Touch .last-pull with current time
	marker := filepath.Join(cloneDir, ".last-pull")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if isCloneStale("owner/repo") {
		t.Error("isCloneStale() returned true for fresh .last-pull")
	}
}

func TestIsCloneStale_OldMarker(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	cloneDir := filepath.Join(reposBase, "owner", "repo")
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(cloneDir, ".last-pull")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// Set mtime to 48 hours ago
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(marker, old, old); err != nil {
		t.Fatal(err)
	}

	if !isCloneStale("owner/repo") {
		t.Error("isCloneStale() returned false for 48-hour-old .last-pull")
	}
}

func TestTouchLastPull(t *testing.T) {
	reposBase, _, cleanup := withTestDirs(t)
	defer cleanup()

	cloneDir := filepath.Join(reposBase, "owner", "repo")
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}

	touchLastPull("owner/repo")

	marker := filepath.Join(cloneDir, ".last-pull")
	info, err := os.Stat(marker)
	if err != nil {
		t.Fatalf("touchLastPull() did not create .last-pull: %v", err)
	}
	if time.Since(info.ModTime()) > time.Second {
		t.Error("touchLastPull() did not set mtime to recent time")
	}
}

func TestCountPath(t *testing.T) {
	_, countsBase, cleanup := withTestDirs(t)
	defer cleanup()

	got := countPath("grafana/alloy")
	want := filepath.Join(countsBase, "grafana/alloy")
	if got != want {
		t.Errorf("countPath() = %q, want %q", got, want)
	}
}
