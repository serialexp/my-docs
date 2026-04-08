// ABOUTME: Tests for the rust command logic.
// ABOUTME: Verifies output formatting for symbol lookup results.

package cmd

import (
	"strings"
	"testing"
)

func TestFormatMultipleMatches(t *testing.T) {
	files := []string{
		"alacritty_terminal/src/term/mod.rs",
		"alacritty_terminal/src/config.rs",
	}
	repo := "alacritty/alacritty"
	symbol := "KeyboardModes"

	output := FormatMultipleMatches(symbol, repo, files)

	if !strings.Contains(output, "KeyboardModes") {
		t.Error("FormatMultipleMatches() should mention the symbol")
	}
	if !strings.Contains(output, "my-docs cat alacritty/alacritty alacritty_terminal/src/term/mod.rs") {
		t.Error("FormatMultipleMatches() should contain cat command for first file")
	}
	if !strings.Contains(output, "my-docs cat alacritty/alacritty alacritty_terminal/src/config.rs") {
		t.Error("FormatMultipleMatches() should contain cat command for second file")
	}
}

func TestFormatNoMatches(t *testing.T) {
	output := FormatNoMatches("KeyboardModes", "alacritty_terminal")

	if !strings.Contains(output, "KeyboardModes") {
		t.Error("FormatNoMatches() should mention the symbol")
	}
	if !strings.Contains(output, "alacritty_terminal") {
		t.Error("FormatNoMatches() should mention the crate")
	}
}
