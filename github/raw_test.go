// ABOUTME: Tests for GitHub raw file fetching.
// ABOUTME: Verifies URL construction for raw.githubusercontent.com.

package github

import "testing"

func TestBuildRawURL(t *testing.T) {
	tests := []struct {
		name   string
		repo   string
		branch string
		path   string
		want   string
	}{
		{
			name:   "main branch",
			repo:   "grafana/alloy",
			branch: "main",
			path:   "docs/intro.md",
			want:   "https://raw.githubusercontent.com/grafana/alloy/main/docs/intro.md",
		},
		{
			name:   "master branch",
			repo:   "torvalds/linux",
			branch: "master",
			path:   "README",
			want:   "https://raw.githubusercontent.com/torvalds/linux/master/README",
		},
		{
			name:   "path with spaces",
			repo:   "owner/repo",
			branch: "main",
			path:   "docs/my file.md",
			want:   "https://raw.githubusercontent.com/owner/repo/main/docs/my%20file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildRawURL(tt.repo, tt.branch, tt.path)
			if got != tt.want {
				t.Errorf("BuildRawURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
