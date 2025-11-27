// ABOUTME: Tests for grep.app API client.
// ABOUTME: Verifies response parsing and URL construction.

package grepapp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseResponse(t *testing.T) {
	raw := `{
		"time": 422,
		"facets": {
			"repo": {
				"buckets": [
					{"val": "grafana/alloy", "count": 1423},
					{"val": "alloy-rs/alloy", "count": 892}
				]
			},
			"lang": {
				"buckets": [
					{"val": "Markdown", "count": 500}
				]
			},
			"path": {
				"buckets": [
					{"val": "docs/", "count": 300}
				]
			}
		},
		"hits": {
			"total": 2315,
			"hits": [
				{
					"repo": "grafana/alloy",
					"branch": "main",
					"path": "docs/intro.md",
					"content": {"snippet": "<table><tr data-line=\"10\"><td>some <mark>match</mark></td></tr></table>"},
					"total_matches": "5"
				}
			]
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.Time != 422 {
		t.Errorf("Time = %d, want 422", resp.Time)
	}

	if len(resp.Facets.Repo.Buckets) != 2 {
		t.Errorf("Repo buckets = %d, want 2", len(resp.Facets.Repo.Buckets))
	}

	if resp.Facets.Repo.Buckets[0].Val != "grafana/alloy" {
		t.Errorf("First repo = %q, want grafana/alloy", resp.Facets.Repo.Buckets[0].Val)
	}

	if resp.Facets.Repo.Buckets[0].Count != 1423 {
		t.Errorf("First repo count = %d, want 1423", resp.Facets.Repo.Buckets[0].Count)
	}

	if resp.Hits.Total != 2315 {
		t.Errorf("Hits.Total = %d, want 2315", resp.Hits.Total)
	}

	if len(resp.Hits.Hits) != 1 {
		t.Errorf("Hits.Hits length = %d, want 1", len(resp.Hits.Hits))
	}

	hit := resp.Hits.Hits[0]
	if hit.Repo != "grafana/alloy" {
		t.Errorf("Hit repo = %q, want grafana/alloy", hit.Repo)
	}
	if hit.Path != "docs/intro.md" {
		t.Errorf("Hit path = %q, want docs/intro.md", hit.Path)
	}
}

func TestBuildURL(t *testing.T) {
	t.Run("simple query", func(t *testing.T) {
		got := BuildURL("test", "")
		if got != "https://grep.app/api/search?q=test" {
			t.Errorf("BuildURL() = %q", got)
		}
	})

	t.Run("with repo filter", func(t *testing.T) {
		got := BuildURL("prometheus", "grafana/alloy")
		if !strings.Contains(got, "q=prometheus") {
			t.Errorf("BuildURL() missing q param: %q", got)
		}
		if !strings.Contains(got, "f.repo=grafana%2Falloy") {
			t.Errorf("BuildURL() missing f.repo param: %q", got)
		}
	})

	t.Run("query with spaces", func(t *testing.T) {
		got := BuildURL("hello world", "")
		if !strings.Contains(got, "q=hello+world") {
			t.Errorf("BuildURL() = %q, want q=hello+world", got)
		}
	})
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name    string
		snippet string
		want    []Match
	}{
		{
			name:    "single line",
			snippet: `<table class="highlight-table"><tr data-line="42"><td><div class="lineno">42</div></td><td><div class="highlight"><pre>some <mark>match</mark> here</pre></div></td></tr></table>`,
			want:    []Match{{Line: 42, Text: "some match here"}},
		},
		{
			name:    "multiple lines",
			snippet: `<table class="highlight-table"><tr data-line="10"><td><div class="lineno">10</div></td><td><div class="highlight"><pre>first</pre></div></td></tr><tr data-line="11"><td><div class="lineno">11</div></td><td><div class="highlight"><pre>second</pre></div></td></tr></table>`,
			want:    []Match{{Line: 10, Text: "first"}, {Line: 11, Text: "second"}},
		},
		{
			name:    "html entities",
			snippet: `<table class="highlight-table"><tr data-line="5"><td><div class="lineno">5</div></td><td><div class="highlight"><pre>&quot;hello&quot; &amp; &lt;world&gt;</pre></div></td></tr></table>`,
			want:    []Match{{Line: 5, Text: `"hello" & <world>`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractText(tt.snippet)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractText() returned %d matches, want %d", len(got), len(tt.want))
				return
			}
			for i, m := range got {
				if m.Line != tt.want[i].Line {
					t.Errorf("Match[%d].Line = %d, want %d", i, m.Line, tt.want[i].Line)
				}
				if m.Text != tt.want[i].Text {
					t.Errorf("Match[%d].Text = %q, want %q", i, m.Text, tt.want[i].Text)
				}
			}
		})
	}
}
