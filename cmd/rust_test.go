// ABOUTME: Tests for the rust command logic.
// ABOUTME: Verifies crate lookup, search result handling, and output formatting.

package cmd

import (
	"strings"
	"testing"

	"github.com/bartriepe/my-docs/grepapp"
)

func TestCollectMatchingFiles_SingleFile(t *testing.T) {
	hits := []grepapp.Hit{
		{Repo: "alacritty/alacritty", Path: "alacritty_terminal/src/term/mod.rs"},
		{Repo: "alacritty/alacritty", Path: "alacritty_terminal/src/term/mod.rs"},
	}

	files := CollectMatchingFiles(hits)

	if len(files) != 1 {
		t.Errorf("CollectMatchingFiles() returned %d files, want 1", len(files))
	}
	if files[0] != "alacritty_terminal/src/term/mod.rs" {
		t.Errorf("CollectMatchingFiles()[0] = %q, want %q", files[0], "alacritty_terminal/src/term/mod.rs")
	}
}

func TestCollectMatchingFiles_MultipleFiles(t *testing.T) {
	hits := []grepapp.Hit{
		{Repo: "alacritty/alacritty", Path: "alacritty_terminal/src/term/mod.rs"},
		{Repo: "alacritty/alacritty", Path: "alacritty_terminal/src/config.rs"},
		{Repo: "alacritty/alacritty", Path: "alacritty_terminal/src/term/mod.rs"},
	}

	files := CollectMatchingFiles(hits)

	if len(files) != 2 {
		t.Errorf("CollectMatchingFiles() returned %d files, want 2", len(files))
	}
}

func TestCollectMatchingFiles_Empty(t *testing.T) {
	hits := []grepapp.Hit{}

	files := CollectMatchingFiles(hits)

	if len(files) != 0 {
		t.Errorf("CollectMatchingFiles() returned %d files, want 0", len(files))
	}
}

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
