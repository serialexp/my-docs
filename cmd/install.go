// ABOUTME: Installs my-docs instructions into ~/.claude/CLAUDE.md.
// ABOUTME: Manages a marked section that can be updated by future versions.

package cmd

import (
	"strings"
)

const startMarker = "<!-- my-docs:start -->"
const endMarker = "<!-- my-docs:end -->"

const Instructions = `## my-docs

You have access to the ` + "`my-docs`" + ` CLI tool for searching documentation across git repositories without cloning them locally.

### When to use my-docs

Use ` + "`my-docs`" + ` when you need to:
- **Understand how a library or framework works** - Search for specific APIs, patterns, or examples in the official repo
- **Find implementation details** - Look at actual source code to understand behavior beyond what docs describe
- **Discover available features** - Search for keywords to see what's possible (e.g., search "exporter" to find all exporters)
- **Reference configuration options** - Search directly for config values (e.g., search "timeout" to find all timeout settings and their documentation). GitHub repos often contain full documentation, so searching for a config option will return both code examples and explanatory docs.
- **Check latest behavior** - Access current documentation without relying on potentially outdated training data

### Workflow

1. **First time using a library?** Set up an alias:
   ` + "`my-docs find opentelemetry`" + ` → discover available repos
   ` + "`my-docs alias otel open-telemetry/opentelemetry-collector`" + ` → save for quick access

2. **Need to find something?** Search the repo:
   ` + "`my-docs search otel \"processor.*metrics\"`" + ` → find metrics processor code
   Returns: file paths with line numbers showing matches

3. **Want to read a specific file?** Fetch it directly:
   ` + "`my-docs cat otel docs/configuration.md`" + ` → read configuration docs
   ` + "`my-docs cat otel processor/metrics/factory.go`" + ` → read source code

### Available commands

- ` + "`my-docs find <query>`" + ` - Search GitHub for repos matching query
- ` + "`my-docs alias <name> <owner/repo>`" + ` - Save repo with a short alias
- ` + "`my-docs search <repo> <pattern>`" + ` - Search repo contents (supports regex)
- ` + "`my-docs cat <repo> <path>`" + ` - Fetch and display file contents
- ` + "`my-docs list`" + ` - Show all configured aliases
- ` + "`my-docs remove <name>`" + ` - Remove an alias

### Tips

- Use search to find examples: ` + "`my-docs search otel \"prometheusreceiver\"`" + `
- Use cat to read docs: ` + "`my-docs cat otel README.md`" + `
- Regex patterns work: ` + "`my-docs search otel \"func.*Start\"`" + `
- Check what's aliased: ` + "`my-docs list`" + `
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
