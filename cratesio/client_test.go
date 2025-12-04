// ABOUTME: Tests for crates.io API client.
// ABOUTME: Verifies crate info lookup and repository extraction.

package cratesio

import (
	"encoding/json"
	"testing"
)

func TestParseResponse_WithRepository(t *testing.T) {
	jsonData := `{
		"crate": {
			"name": "alacritty_terminal",
			"repository": "https://github.com/alacritty/alacritty"
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	repo, err := ExtractGitHubRepo(&resp)
	if err != nil {
		t.Fatalf("ExtractGitHubRepo() error = %v", err)
	}
	if repo != "alacritty/alacritty" {
		t.Errorf("ExtractGitHubRepo() = %q, want %q", repo, "alacritty/alacritty")
	}
}

func TestParseResponse_WithTrailingSlash(t *testing.T) {
	jsonData := `{
		"crate": {
			"name": "serde",
			"repository": "https://github.com/serde-rs/serde/"
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	repo, err := ExtractGitHubRepo(&resp)
	if err != nil {
		t.Fatalf("ExtractGitHubRepo() error = %v", err)
	}
	if repo != "serde-rs/serde" {
		t.Errorf("ExtractGitHubRepo() = %q, want %q", repo, "serde-rs/serde")
	}
}

func TestParseResponse_WithSubdirectory(t *testing.T) {
	// Some crates point to a subdirectory in a monorepo
	jsonData := `{
		"crate": {
			"name": "tokio-macros",
			"repository": "https://github.com/tokio-rs/tokio/tree/master/tokio-macros"
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	repo, err := ExtractGitHubRepo(&resp)
	if err != nil {
		t.Fatalf("ExtractGitHubRepo() error = %v", err)
	}
	if repo != "tokio-rs/tokio" {
		t.Errorf("ExtractGitHubRepo() = %q, want %q", repo, "tokio-rs/tokio")
	}
}

func TestParseResponse_NoRepository(t *testing.T) {
	jsonData := `{
		"crate": {
			"name": "some-crate",
			"repository": null
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	_, err := ExtractGitHubRepo(&resp)
	if err == nil {
		t.Error("ExtractGitHubRepo() error = nil, want error for missing repository")
	}
}

func TestParseResponse_NonGitHubRepository(t *testing.T) {
	jsonData := `{
		"crate": {
			"name": "some-crate",
			"repository": "https://gitlab.com/foo/bar"
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	_, err := ExtractGitHubRepo(&resp)
	if err == nil {
		t.Error("ExtractGitHubRepo() error = nil, want error for non-GitHub repository")
	}
}

func TestBuildURL(t *testing.T) {
	url := BuildURL("alacritty_terminal")
	expected := "https://crates.io/api/v1/crates/alacritty_terminal"
	if url != expected {
		t.Errorf("BuildURL() = %q, want %q", url, expected)
	}
}
