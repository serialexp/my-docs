# my-docs

CLI tool for searching documentation across git repositories without cloning them locally. Designed to give AI coding agents fast, searchable access to any GitHub repository's documentation and source code.

## Problem

When working with AI agents, they need access to documentation to work effectively. Current options:
- Clone repos locally → clutters disk, manual updates
- Web requests → slow, can't grep
- DevDocs/Dash → only mainstream frameworks, can't add custom repos

## Solution

A CLI that:
1. **Searches** via [grep.app](https://grep.app)'s API (fast, indexed across all public GitHub repos)
2. **Reads files** via raw.githubusercontent.com (with local caching)
3. **Auto-clones** frequently accessed repos locally for instant reads
4. **Integrates** with Claude Code via a single `install` command

## Installation

### Quick install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/serialexp/my-docs/main/install.sh | bash
```

This automatically detects your platform, downloads the latest release, and installs it.

### Using Go

```bash
go install github.com/serialexp/my-docs@latest
```

Note: Requires `~/go/bin` in your PATH.

### Build from source

```bash
git clone https://github.com/serialexp/my-docs
cd my-docs
go build
cp my-docs ~/.local/bin/
```

## Usage

```bash
# Find repos to work with
my-docs find alloy
# grafana/alloy (2233 matches)
# alloy-rs/alloy (430 matches)

# Get a quick overview of a repo (fetches llms.txt, falls back to README)
my-docs info grafana/alloy

# Search for patterns (use owner/repo format)
my-docs search grafana/alloy "prometheus.exporter"
# internal/component/prometheus/exporter/self/self.go:14: Name: "prometheus.exporter.self",

# Search across all repos
my-docs search "specific_function_name"

# Read specific files
my-docs cat grafana/alloy README.md

# Read specific lines from a file
my-docs cat grafana/alloy internal/config/config.go --lines 10-50

# Look up Rust crate symbols
my-docs rust alacritty_terminal KeyboardModes

# Install instructions into ~/.claude/CLAUDE.md for AI agents
my-docs install
```

## Commands

| Command | Description |
|---------|-------------|
| `find <query>` | Search for repos by name via grep.app |
| `search [owner/repo] <pattern>` | Search repo contents (supports regex). Omit repo to search all. Supports `--limit N` and `--offset N` for pagination |
| `cat <owner/repo> <path>` | Fetch and display file from GitHub. Supports `--lines N-M` for line ranges |
| `info <owner/repo>` | Show repo overview (prefers `llms.txt`, falls back to `README.md`) |
| `rust <crate> <symbol>` | Look up a Rust crate symbol and show its source |
| `install` | Install usage instructions into `~/.claude/CLAUDE.md` |
| `version` | Show version |

## How it works

### Searching

`find` and `search` use the [grep.app](https://grep.app) API, which indexes all public GitHub repositories. This means you can search any public repo without cloning it first. Search patterns support regex.

### File fetching

`cat` and `info` fetch files via `raw.githubusercontent.com` with automatic branch detection (tries `main`, then `master`). Fetched files are cached locally for 15 minutes to avoid redundant HTTP requests.

### Auto-cloning

After a repo has been accessed 5 or more times, `my-docs` automatically clones it locally (shallow, in the background) for instant file reads. Cloned repos are stored in your system cache directory (`~/Library/Caches/my-docs/repos/` on macOS) and kept up to date with periodic `git pull` (every 24 hours). If the clone is unavailable or a file isn't found locally, it falls back to HTTP transparently.

### Rust crate resolution

The `rust` command resolves crate names to their GitHub repository via crates.io, caches the mapping in `~/.config/my-docs/config.json`, then searches for the symbol. If the symbol is found in a single file, it outputs the full file. If found in multiple files, it lists `my-docs cat` commands you can run.

## For AI Agents

Run `my-docs install` to add usage instructions to your `~/.claude/CLAUDE.md`. This teaches Claude Code how and when to use the tool. The instructions are managed in a marked section that gets updated cleanly when you run `install` again.

Here's what gets added to your CLAUDE.md (this is the exact prompt Claude Code sees):

<details>
<summary>Claude Code instructions (click to expand)</summary>

````markdown
## my-docs

You have access to the `my-docs` CLI tool for searching documentation across git repositories without cloning them locally.

### When to use my-docs

Use `my-docs` when you need to:
- **Understand how a library or framework works** - Search for specific APIs, patterns, or examples in the official repo
- **Find implementation details** - Look at actual source code to understand behavior beyond what docs describe
- **Discover available features** - Search for keywords to see what's possible (e.g., search "exporter" to find all exporters)
- **Reference configuration options** - Search directly for config values (e.g., search "timeout" to find all timeout settings and their documentation). GitHub repos often contain full documentation, so searching for a config option will return both code examples and explanatory docs.
- **Check latest behavior** - Access current documentation without relying on potentially outdated training data

### Workflow

1. **Not sure about the exact repo name?** Find it first:
   `my-docs find opentelemetry` → discover available repos (returns owner/repo names)

2. **First time using a library?** Get a quick overview:
   `my-docs info grafana/alloy` → fetch the repo's llms.txt (or README if unavailable)

3. **Need to find something?** Search the repo:
   `my-docs search open-telemetry/opentelemetry-collector "processor.*metrics"` → find metrics processor code
   Returns: file paths with line numbers showing matches

4. **Want to read a specific file?** Fetch it directly:
   `my-docs cat open-telemetry/opentelemetry-collector docs/configuration.md` → read configuration docs
   `my-docs cat open-telemetry/opentelemetry-collector processor/metrics/factory.go` → read source code
   `my-docs cat open-telemetry/opentelemetry-collector processor/metrics/factory.go --lines 10-50` → read specific lines

### Available commands

- `my-docs find <query>` - Search GitHub for repos matching query
- `my-docs search [owner/repo] <pattern>` - Search repo contents (supports regex). Repo should be in owner/repo format, or omitted to search all repos
- `my-docs cat <owner/repo> <path> [--lines N-M]` - Fetch and display file contents (optionally a line range)
- `my-docs info <owner/repo>` - Show repo overview (prefers llms.txt, falls back to README.md)
- `my-docs rust <crate> <symbol>` - Look up a Rust crate symbol and show its source

### Rust Crates

For Rust crates, use the `rust` command to look up symbols directly:
```
my-docs rust alacritty_terminal KeyboardModes
```
This automatically resolves the crate to its GitHub repository (cached for future use).
If the symbol appears in multiple files, you'll get a list of cat commands to run.

### Important

NEVER use `2>/dev/null` or `2>&1` when running `my-docs` commands. If the command fails or returns empty output, you need to see the error message to understand why. Empty output with suppressed stderr likely means the tool is not installed.

### Tips

- Search specific repo: `my-docs search grafana/alloy "exporter"`
- Search all repos: `my-docs search "specific_function_name"`
- Use cat to read docs: `my-docs cat grafana/alloy README.md`
- Regex patterns work: `my-docs search grafana/alloy "func.*Start"`
````

</details>
