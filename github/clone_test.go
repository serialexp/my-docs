// ABOUTME: Tests for local git clone management.
// ABOUTME: Covers clone detection, local file reads, staleness checks, and path construction.

package github

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTestDirs(t *testing.T) (reposBase string, cleanup func()) {
	t.Helper()
	reposBase = t.TempDir()

	origRepos := reposDirFn
	reposDirFn = func() string { return reposBase }

	return reposBase, func() {
		reposDirFn = origRepos
	}
}

func TestClonePath(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
	defer cleanup()

	got := clonePath("grafana/alloy")
	want := filepath.Join(reposBase, "grafana/alloy")
	if got != want {
		t.Errorf("clonePath() = %q, want %q", got, want)
	}
}

func TestCloneTmpPath(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
	defer cleanup()

	got := cloneTmpPath("grafana/alloy")
	want := filepath.Join(reposBase, "grafana/alloy.tmp")
	if got != want {
		t.Errorf("cloneTmpPath() = %q, want %q", got, want)
	}
}

func TestCloneReady_NoDir(t *testing.T) {
	_, cleanup := withTestDirs(t)
	defer cleanup()

	if cloneReady("grafana/alloy") {
		t.Error("cloneReady() returned true for non-existent directory")
	}
}

func TestCloneReady_WithGitDir(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
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
	reposBase, cleanup := withTestDirs(t)
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
	_, cleanup := withTestDirs(t)
	defer cleanup()

	if cloneInProgress("grafana/alloy") {
		t.Error("cloneInProgress() returned true for non-existent .tmp directory")
	}
}

func TestReadLocalFile_Success(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
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
	_, cleanup := withTestDirs(t)
	defer cleanup()

	_, err := readLocalFile("owner/repo", "nonexistent.md")
	if err == nil {
		t.Error("readLocalFile() returned nil error for non-existent file")
	}
}

func TestIsCloneStale_NoMarker(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
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
	reposBase, cleanup := withTestDirs(t)
	defer cleanup()

	cloneDir := filepath.Join(reposBase, "owner", "repo")
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(cloneDir, ".last-pull")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if isCloneStale("owner/repo") {
		t.Error("isCloneStale() returned true for fresh .last-pull")
	}
}

func TestIsCloneStale_OldMarker(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
	defer cleanup()

	cloneDir := filepath.Join(reposBase, "owner", "repo")
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(cloneDir, ".last-pull")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(marker, old, old); err != nil {
		t.Fatal(err)
	}

	if !isCloneStale("owner/repo") {
		t.Error("isCloneStale() returned false for 48-hour-old .last-pull")
	}
}

func TestTouchLastPull(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
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

func TestCanSearchLocal(t *testing.T) {
	reposBase, cleanup := withTestDirs(t)
	defer cleanup()

	if CanSearchLocal("owner/repo") {
		t.Error("CanSearchLocal() returned true for non-existent clone")
	}

	gitDir := filepath.Join(reposBase, "owner", "repo", ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !CanSearchLocal("owner/repo") {
		t.Error("CanSearchLocal() returned false when clone exists")
	}
}

func TestParseRipgrepOutput(t *testing.T) {
	cloneDir := "/home/user/.cache/my-docs/repos/owner/repo"
	output := `/home/user/.cache/my-docs/repos/owner/repo/src/main.go:42:func handleRequest() {
/home/user/.cache/my-docs/repos/owner/repo/src/handler.go:10:import "net/http"
/home/user/.cache/my-docs/repos/owner/repo/docs/api.md:5:The request handler processes incoming calls
`
	matches := parseRipgrepOutput(output, cloneDir)

	if len(matches) != 3 {
		t.Fatalf("parseRipgrepOutput() returned %d matches, want 3", len(matches))
	}

	want := []SearchMatch{
		{Path: "src/main.go", Line: 42, Text: "func handleRequest() {"},
		{Path: "src/handler.go", Line: 10, Text: `import "net/http"`},
		{Path: "docs/api.md", Line: 5, Text: "The request handler processes incoming calls"},
	}

	for i, got := range matches {
		if got.Path != want[i].Path {
			t.Errorf("match[%d].Path = %q, want %q", i, got.Path, want[i].Path)
		}
		if got.Line != want[i].Line {
			t.Errorf("match[%d].Line = %d, want %d", i, got.Line, want[i].Line)
		}
		if got.Text != want[i].Text {
			t.Errorf("match[%d].Text = %q, want %q", i, got.Text, want[i].Text)
		}
	}
}

func TestParseRipgrepOutput_Empty(t *testing.T) {
	matches := parseRipgrepOutput("", "/some/dir")
	if len(matches) != 0 {
		t.Errorf("parseRipgrepOutput() returned %d matches for empty output, want 0", len(matches))
	}
}

func TestParseRipgrepOutput_ColonsInText(t *testing.T) {
	cloneDir := "/cache/repos/owner/repo"
	output := `/cache/repos/owner/repo/config.yaml:15:server: host:8080
`
	matches := parseRipgrepOutput(output, cloneDir)

	if len(matches) != 1 {
		t.Fatalf("parseRipgrepOutput() returned %d matches, want 1", len(matches))
	}
	if matches[0].Path != "config.yaml" {
		t.Errorf("Path = %q, want %q", matches[0].Path, "config.yaml")
	}
	if matches[0].Line != 15 {
		t.Errorf("Line = %d, want 15", matches[0].Line)
	}
	if matches[0].Text != "server: host:8080" {
		t.Errorf("Text = %q, want %q", matches[0].Text, "server: host:8080")
	}
}
