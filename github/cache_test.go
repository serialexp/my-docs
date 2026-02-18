// ABOUTME: Tests for the file cache used by FetchFile.
// ABOUTME: Covers path construction, round-trip, TTL expiry, and cache misses.

package github

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCachePath(t *testing.T) {
	got := cachePath("/tmp/cache", "grafana/alloy", "main", "docs/intro.md")
	want := filepath.Join("/tmp/cache", "grafana/alloy", "main", "docs/intro.md")
	if got != want {
		t.Errorf("cachePath() = %q, want %q", got, want)
	}
}

func TestReadCacheMiss(t *testing.T) {
	_, ok := readCache("/nonexistent/path/that/does/not/exist")
	if ok {
		t.Error("readCache() returned true for non-existent file")
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "owner", "repo", "main", "file.md")
	content := "# Hello\nThis is cached content."

	writeCache(path, content)

	got, ok := readCache(path)
	if !ok {
		t.Fatal("readCache() returned false after writeCache()")
	}
	if got != content {
		t.Errorf("readCache() = %q, want %q", got, content)
	}
}

func TestExpiredCacheMiss(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "expired.txt")
	content := "old content"

	writeCache(path, content)

	// Set mtime to 20 minutes ago
	old := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	_, ok := readCache(path)
	if ok {
		t.Error("readCache() returned true for expired file")
	}
}

func TestFreshCacheHit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fresh.txt")
	content := "fresh content"

	writeCache(path, content)

	// Set mtime to 5 minutes ago (well within TTL)
	recent := time.Now().Add(-5 * time.Minute)
	if err := os.Chtimes(path, recent, recent); err != nil {
		t.Fatal(err)
	}

	got, ok := readCache(path)
	if !ok {
		t.Fatal("readCache() returned false for fresh file")
	}
	if got != content {
		t.Errorf("readCache() = %q, want %q", got, content)
	}
}
