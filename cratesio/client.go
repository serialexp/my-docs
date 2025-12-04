// ABOUTME: HTTP client for crates.io API.
// ABOUTME: Looks up Rust crates and extracts their GitHub repository URLs.

package cratesio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const baseURL = "https://crates.io/api/v1/crates"

type Response struct {
	Crate Crate `json:"crate"`
}

type Crate struct {
	Name       string  `json:"name"`
	Repository *string `json:"repository"`
}

func BuildURL(crateName string) string {
	return fmt.Sprintf("%s/%s", baseURL, crateName)
}

func Lookup(crateName string) (*Response, error) {
	url := BuildURL(crateName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "my-docs/1.0 (https://github.com/serialexp/my-docs)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("crate %q not found on crates.io", crateName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crates.io returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func ExtractGitHubRepo(resp *Response) (string, error) {
	if resp.Crate.Repository == nil || *resp.Crate.Repository == "" {
		return "", errors.New("crate has no repository URL")
	}

	repoURL := *resp.Crate.Repository

	if !strings.Contains(repoURL, "github.com") {
		return "", fmt.Errorf("repository %q is not on GitHub", repoURL)
	}

	// Extract owner/repo from URLs like:
	// - https://github.com/owner/repo
	// - https://github.com/owner/repo/
	// - https://github.com/owner/repo/tree/master/subdir
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "http://")
	repoURL = strings.TrimPrefix(repoURL, "github.com/")

	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid GitHub URL: %q", *resp.Crate.Repository)
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")

	return owner + "/" + repo, nil
}
