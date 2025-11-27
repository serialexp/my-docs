// ABOUTME: Tests for the install command.
// ABOUTME: Verifies CLAUDE.md section insertion and updating.

package cmd

import (
	"strings"
	"testing"
)

func TestUpdateClaudeMd_NoExistingSection(t *testing.T) {
	existing := `# My CLAUDE.md

Some existing content here.
`
	instructions := "Use my-docs to search repos."

	result := UpdateClaudeMdSection(existing, instructions)

	if !strings.Contains(result, "Some existing content here.") {
		t.Error("UpdateClaudeMdSection removed existing content")
	}
	if !strings.Contains(result, "<!-- my-docs:start -->") {
		t.Error("UpdateClaudeMdSection missing start marker")
	}
	if !strings.Contains(result, "<!-- my-docs:end -->") {
		t.Error("UpdateClaudeMdSection missing end marker")
	}
	if !strings.Contains(result, "Use my-docs to search repos.") {
		t.Error("UpdateClaudeMdSection missing instructions")
	}
}

func TestUpdateClaudeMd_ExistingSection(t *testing.T) {
	existing := `# My CLAUDE.md

Some existing content here.

<!-- my-docs:start -->
Old instructions that should be replaced.
<!-- my-docs:end -->

More content after.
`
	instructions := "New instructions here."

	result := UpdateClaudeMdSection(existing, instructions)

	if !strings.Contains(result, "Some existing content here.") {
		t.Error("UpdateClaudeMdSection removed content before section")
	}
	if !strings.Contains(result, "More content after.") {
		t.Error("UpdateClaudeMdSection removed content after section")
	}
	if strings.Contains(result, "Old instructions") {
		t.Error("UpdateClaudeMdSection did not replace old instructions")
	}
	if !strings.Contains(result, "New instructions here.") {
		t.Error("UpdateClaudeMdSection missing new instructions")
	}
	if strings.Count(result, "<!-- my-docs:start -->") != 1 {
		t.Error("UpdateClaudeMdSection has duplicate start markers")
	}
}

func TestUpdateClaudeMd_EmptyFile(t *testing.T) {
	existing := ""
	instructions := "Instructions for empty file."

	result := UpdateClaudeMdSection(existing, instructions)

	if !strings.Contains(result, "<!-- my-docs:start -->") {
		t.Error("UpdateClaudeMdSection missing start marker")
	}
	if !strings.Contains(result, "Instructions for empty file.") {
		t.Error("UpdateClaudeMdSection missing instructions")
	}
}
