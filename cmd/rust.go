// ABOUTME: Logic for the rust command.
// ABOUTME: Resolves Rust crate names to GitHub repos and searches for symbols.

package cmd

import (
	"fmt"
	"strings"

	"github.com/bartriepe/my-docs/grepapp"
)

func CollectMatchingFiles(hits []grepapp.Hit) []string {
	seen := make(map[string]bool)
	var files []string

	for _, hit := range hits {
		if !seen[hit.Path] {
			seen[hit.Path] = true
			files = append(files, hit.Path)
		}
	}

	return files
}

func FormatMultipleMatches(symbol, repo string, files []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found '%s' in %d files:\n", symbol, len(files)))
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("  my-docs cat %s %s\n", repo, f))
	}
	return sb.String()
}

func FormatNoMatches(symbol, crate string) string {
	return fmt.Sprintf("No matches found for '%s' in crate '%s'\n", symbol, crate)
}
