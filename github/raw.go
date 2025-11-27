// ABOUTME: Fetches raw files from GitHub via raw.githubusercontent.com.
// ABOUTME: Handles URL construction and branch fallback (main -> master).

package github

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const rawBaseURL = "https://raw.githubusercontent.com"

func BuildRawURL(repo, branch, path string) string {
	encodedPath := url.PathEscape(path)
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")
	return fmt.Sprintf("%s/%s/%s/%s", rawBaseURL, repo, branch, encodedPath)
}

func FetchFile(repo, path string) (string, error) {
	branches := []string{"main", "master"}

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
		return string(body), nil
	}

	return "", fmt.Errorf("could not fetch %s/%s: %v", repo, path, lastErr)
}
