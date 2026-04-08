// ABOUTME: Logic for the rust command.
// ABOUTME: Formats output for Rust crate symbol lookups.

package cmd

import (
	"fmt"
	"strings"
)

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
