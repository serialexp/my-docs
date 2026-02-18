// ABOUTME: Fetches raw files from GitHub via raw.githubusercontent.com.
// ABOUTME: Handles URL construction and branch fallback (main -> master).

package github

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const rawBaseURL = "https://raw.githubusercontent.com"

func BuildRawURL(repo, branch, path string) string {
	encodedPath := url.PathEscape(path)
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")
	return fmt.Sprintf("%s/%s/%s/%s", rawBaseURL, repo, branch, encodedPath)
}

func FetchFile(repo, path string) (string, error) {
	branches := []string{"main", "master"}
	base := cacheDir()

	// Check cache for each branch before making any HTTP requests
	if base != "" {
		for _, branch := range branches {
			cp := cachePath(base, repo, branch, path)
			if content, ok := readCache(cp); ok {
				return content, nil
			}
		}
	}

	var lastErr error
	for _, branch := range branches {
		rawURL := BuildRawURL(repo, branch, path)

		resp, err := http.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("not found on branch %s", branch)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		content := string(body)
		if base != "" {
			writeCache(cachePath(base, repo, branch, path), content)
		}
		return content, nil
	}

	return "", fmt.Errorf("could not fetch %s/%s: %v", repo, path, lastErr)
}

// FetchRepoInfo fetches llms.txt and README.md concurrently, preferring llms.txt.
func FetchRepoInfo(repo string) (string, error) {
	type result struct {
		content string
		err     error
	}

	var (
		llms, readme result
		wg           sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		llms.content, llms.err = FetchFile(repo, "llms.txt")
	}()
	go func() {
		defer wg.Done()
		readme.content, readme.err = FetchFile(repo, "README.md")
	}()
	wg.Wait()

	if llms.err == nil {
		return llms.content, nil
	}
	if readme.err == nil {
		return readme.content, nil
	}
	return "", fmt.Errorf("could not fetch repo info for %s: no llms.txt or README.md found", repo)
}
