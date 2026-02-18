// ABOUTME: File cache for raw GitHub content to avoid redundant HTTP requests.
// ABOUTME: Uses ~/.cache/my-docs/files/ with 15-minute TTL based on file mtime.

package github

import (
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 15 * time.Minute

func cacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "my-docs", "files")
}

func cachePath(base, repo, branch, path string) string {
	return filepath.Join(base, repo, branch, path)
}

func readCache(path string) (string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if time.Since(info.ModTime()) > cacheTTL {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func writeCache(path, content string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(content), 0o644)
}
