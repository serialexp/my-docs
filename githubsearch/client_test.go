// ABOUTME: Tests for GitHub search API clients.
// ABOUTME: Verifies code search, repository search, URL building, and result extraction.

package githubsearch

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildCodeSearchURL(t *testing.T) {
	t.Run("basic query", func(t *testing.T) {
		got := BuildCodeSearchURL("test", 30)
		if !strings.Contains(got, "q=test") {
			t.Errorf("missing q param: %q", got)
		}
		if !strings.Contains(got, "per_page=30") {
			t.Errorf("missing per_page param: %q", got)
		}
		if !strings.HasPrefix(got, codeSearchURL) {
			t.Errorf("should start with %q, got %q", codeSearchURL, got)
		}
	})

	t.Run("zero perPage omitted", func(t *testing.T) {
		got := BuildCodeSearchURL("test", 0)
		if strings.Contains(got, "per_page=") {
			t.Errorf("should omit per_page when 0: %q", got)
		}
	})
}

func TestBuildRepoSearchURL(t *testing.T) {
	got := BuildRepoSearchURL("opentelemetry", 20)
	if !strings.Contains(got, "q=opentelemetry") {
		t.Errorf("missing q param: %q", got)
	}
	if !strings.Contains(got, "sort=stars") {
		t.Errorf("missing sort param: %q", got)
	}
	if !strings.Contains(got, "order=desc") {
		t.Errorf("missing order param: %q", got)
	}
	if !strings.Contains(got, "per_page=20") {
		t.Errorf("missing per_page param: %q", got)
	}
	if !strings.HasPrefix(got, repoSearchURL) {
		t.Errorf("should start with %q, got %q", repoSearchURL, got)
	}
}

func TestParseCodeSearchResponse(t *testing.T) {
	raw := `{
		"total_count": 42,
		"incomplete_results": false,
		"items": [
			{
				"name": "main.go",
				"path": "cmd/main.go",
				"repository": {
					"full_name": "owner/repo"
				},
				"text_matches": [
					{
						"fragment": "func main() {\n\tfmt.Println(\"hello\")\n}",
						"matches": [
							{
								"text": "main",
								"indices": [5, 9]
							}
						]
					}
				]
			},
			{
				"name": "server.go",
				"path": "internal/server.go",
				"repository": {
					"full_name": "owner/repo"
				},
				"text_matches": []
			}
		]
	}`

	var resp CodeSearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.TotalCount != 42 {
		t.Errorf("TotalCount = %d, want 42", resp.TotalCount)
	}

	if resp.IncompleteResults {
		t.Error("IncompleteResults should be false")
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Items length = %d, want 2", len(resp.Items))
	}

	item := resp.Items[0]
	if item.Path != "cmd/main.go" {
		t.Errorf("Item.Path = %q, want %q", item.Path, "cmd/main.go")
	}
	if item.Repository.FullName != "owner/repo" {
		t.Errorf("Item.Repository.FullName = %q, want %q", item.Repository.FullName, "owner/repo")
	}

	if len(item.TextMatches) != 1 {
		t.Fatalf("TextMatches length = %d, want 1", len(item.TextMatches))
	}

	tm := item.TextMatches[0]
	if !strings.Contains(tm.Fragment, "func main()") {
		t.Errorf("Fragment = %q, want to contain 'func main()'", tm.Fragment)
	}
	if len(tm.Matches) != 1 {
		t.Fatalf("Matches length = %d, want 1", len(tm.Matches))
	}
	if tm.Matches[0].Text != "main" {
		t.Errorf("Match.Text = %q, want %q", tm.Matches[0].Text, "main")
	}
}

func TestParseRepoSearchResponse(t *testing.T) {
	raw := `{
		"total_count": 5,
		"incomplete_results": false,
		"items": [
			{
				"full_name": "open-telemetry/opentelemetry-go",
				"description": "OpenTelemetry Go API and SDK",
				"stargazers_count": 5000
			},
			{
				"full_name": "open-telemetry/opentelemetry-collector",
				"description": "OpenTelemetry Collector",
				"stargazers_count": 4000
			}
		]
	}`

	var resp RepoSearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", resp.TotalCount)
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Items length = %d, want 2", len(resp.Items))
	}

	if resp.Items[0].FullName != "open-telemetry/opentelemetry-go" {
		t.Errorf("Items[0].FullName = %q", resp.Items[0].FullName)
	}
	if resp.Items[0].Stars != 5000 {
		t.Errorf("Items[0].Stars = %d, want 5000", resp.Items[0].Stars)
	}
	if resp.Items[0].Description != "OpenTelemetry Go API and SDK" {
		t.Errorf("Items[0].Description = %q", resp.Items[0].Description)
	}
}

func TestExtractResults(t *testing.T) {
	resp := &CodeSearchResponse{
		TotalCount: 3,
		Items: []CodeItem{
			{
				Path:       "cmd/main.go",
				Repository: Repository{FullName: "owner/repo"},
				TextMatches: []TextMatch{
					{
						Fragment: "func main() {\n\tfmt.Println(\"hello\")\n}\n",
						Matches:  []Match{{Text: "main", Indices: []int{5, 9}}},
					},
					{
						Fragment: "// main package\n",
						Matches:  []Match{{Text: "main", Indices: []int{3, 7}}},
					},
				},
			},
			{
				Path:        "internal/server.go",
				Repository:  Repository{FullName: "owner/repo"},
				TextMatches: []TextMatch{},
			},
		},
	}

	results := ExtractResults(resp)

	// Two fragments from first item + one entry for second item (no fragments)
	if len(results) != 3 {
		t.Fatalf("ExtractResults() returned %d results, want 3", len(results))
	}

	if results[0].Path != "cmd/main.go" {
		t.Errorf("results[0].Path = %q, want %q", results[0].Path, "cmd/main.go")
	}
	if !strings.Contains(results[0].Fragment, "func main()") {
		t.Errorf("results[0].Fragment should contain 'func main()', got %q", results[0].Fragment)
	}
	// Trailing newlines should be trimmed
	if strings.HasSuffix(results[0].Fragment, "\n") {
		t.Errorf("results[0].Fragment should not have trailing newline, got %q", results[0].Fragment)
	}

	if results[2].Path != "internal/server.go" {
		t.Errorf("results[2].Path = %q, want %q", results[2].Path, "internal/server.go")
	}
	if results[2].Fragment != "" {
		t.Errorf("results[2].Fragment = %q, want empty string", results[2].Fragment)
	}
}
