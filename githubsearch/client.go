// ABOUTME: HTTP client for GitHub search APIs.
// ABOUTME: Code search for cross-repo discovery, repository search for finding repos by name.

package githubsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const codeSearchURL = "https://api.github.com/search/code"
const repoSearchURL = "https://api.github.com/search/repositories"

// --- Code search types (for cross-repo search without a specific repo) ---

// CodeSearchResponse is the top-level GitHub code search response.
type CodeSearchResponse struct {
	TotalCount        int        `json:"total_count"`
	IncompleteResults bool       `json:"incomplete_results"`
	Items             []CodeItem `json:"items"`
}

// CodeItem represents a single file that matched the code search query.
type CodeItem struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Repository  Repository  `json:"repository"`
	TextMatches []TextMatch `json:"text_matches"`
}

// Repository contains the repo metadata for a search result.
type Repository struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

// TextMatch is a fragment of code from the file containing a match.
// Requires Accept: application/vnd.github.text-match+json header.
type TextMatch struct {
	Fragment string  `json:"fragment"`
	Matches  []Match `json:"matches"`
}

// Match describes a single matched string within a text fragment.
type Match struct {
	Text    string `json:"text"`
	Indices []int  `json:"indices"`
}

// ResultLine represents a single line of search output for display.
type ResultLine struct {
	Path     string
	Fragment string
	Repo     string
}

// --- Repository search types (for finding repos by name) ---

// RepoSearchResponse is the top-level GitHub repository search response.
type RepoSearchResponse struct {
	TotalCount        int        `json:"total_count"`
	IncompleteResults bool       `json:"incomplete_results"`
	Items             []RepoItem `json:"items"`
}

// RepoItem represents a single repository from a repository search.
type RepoItem struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
}

// --- Shared HTTP helper ---

func doGitHubGet(reqURL string, accept string) (*http.Response, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "my-docs/1.0")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return http.DefaultClient.Do(req)
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		return fmt.Errorf("GitHub API rate limit exceeded (status %d); try again in a minute", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub search query invalid (status 422): %s", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// --- Code search (cross-repo) ---

// BuildCodeSearchURL constructs the full GitHub code search API URL.
func BuildCodeSearchURL(query string, perPage int) string {
	params := url.Values{}
	params.Set("q", query)
	if perPage > 0 {
		params.Set("per_page", fmt.Sprintf("%d", perPage))
	}
	return codeSearchURL + "?" + params.Encode()
}

// Search performs a GitHub code search and returns the parsed response.
// It requests text match metadata to get code fragments.
// This is used only for cross-repo search (when no specific repo is specified).
func Search(pattern, repo string) (*CodeSearchResponse, error) {
	q := pattern
	if repo != "" {
		q += " repo:" + repo
	}
	searchURL := BuildCodeSearchURL(q, 100)

	resp, err := doGitHubGet(searchURL, "application/vnd.github.text-match+json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result CodeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ExtractResults converts a GitHub code search response into display-friendly result lines.
// Each text match fragment becomes one result line.
func ExtractResults(resp *CodeSearchResponse) []ResultLine {
	var results []ResultLine
	for _, item := range resp.Items {
		if len(item.TextMatches) == 0 {
			results = append(results, ResultLine{
				Path: item.Path,
				Repo: item.Repository.FullName,
			})
			continue
		}
		for _, tm := range item.TextMatches {
			fragment := strings.TrimRight(tm.Fragment, "\n")
			results = append(results, ResultLine{
				Path:     item.Path,
				Fragment: fragment,
				Repo:     item.Repository.FullName,
			})
		}
	}
	return results
}

// --- Repository search (for find command) ---

// BuildRepoSearchURL constructs the full GitHub repository search API URL.
func BuildRepoSearchURL(query string, perPage int) string {
	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	if perPage > 0 {
		params.Set("per_page", fmt.Sprintf("%d", perPage))
	}
	return repoSearchURL + "?" + params.Encode()
}

// SearchRepos searches for GitHub repositories matching the query, sorted by stars.
func SearchRepos(query string) (*RepoSearchResponse, error) {
	searchURL := BuildRepoSearchURL(query, 20)

	resp, err := doGitHubGet(searchURL, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result RepoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
