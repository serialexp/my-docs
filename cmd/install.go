// ABOUTME: Installs my-docs instructions into ~/.claude/CLAUDE.md.
// ABOUTME: Manages a marked section that can be updated by future versions.

package cmd

import (
	"strings"
)

const startMarker = "<!-- my-docs:start -->"
const endMarker = "<!-- my-docs:end -->"

const Instructions = `## my-docs

CLI tool for searching documentation across git repositories without cloning them locally.

### Commands

- ` + "`my-docs search <repo> <pattern>`" + ` - Search repo via grep.app, show matches with file:line: format
- ` + "`my-docs cat <repo> <path>`" + ` - Fetch and display raw file from GitHub
- ` + "`my-docs find <query>`" + ` - Search for repos by name (to discover repos to alias)
- ` + "`my-docs alias <name> <owner/repo>`" + ` - Create a short alias for a repository
- ` + "`my-docs list`" + ` - Show all configured repo aliases
- ` + "`my-docs remove <name>`" + ` - Remove a repo alias
- ` + "`my-docs config`" + ` - Show config file path

### Usage

When you need documentation for a library or tool:
1. Use ` + "`my-docs find <name>`" + ` to discover the repository
2. Use ` + "`my-docs alias <short> <owner/repo>`" + ` to save it
3. Use ` + "`my-docs search <short> <pattern>`" + ` to find relevant docs
4. Use ` + "`my-docs cat <short> <path>`" + ` to read specific files
`

func UpdateClaudeMdSection(content, instructions string) string {
	section := startMarker + "\n" + instructions + "\n" + endMarker

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		before := content[:startIdx]
		after := content[endIdx+len(endMarker):]
		return before + section + after
	}

	if content == "" {
		return section + "\n"
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + section + "\n"
}
