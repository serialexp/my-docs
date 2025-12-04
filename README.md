# my-docs

CLI tool for searching documentation across git repositories without cloning them locally.

## Problem

When working with AI agents, they need access to documentation to work effectively. Current options:
- Clone repos locally → clutters disk, manual updates
- Web requests → slow, can't grep
- DevDocs/Dash → only mainstream frameworks, can't add custom repos

## Solution

A CLI that:
1. **Searches** via grep.app's API (fast, indexed)
2. **Reads files** via raw.githubusercontent.com
3. **Maps short names** to full repo paths via config

## Installation

### Download binary (easiest)

Download the latest release for your platform from [GitHub Releases](https://github.com/serialexp/my-docs/releases), then:

```bash
# macOS/Linux
chmod +x my-docs-*
sudo mv my-docs-* /usr/local/bin/my-docs

# Verify
my-docs help
```

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
sudo cp my-docs /usr/local/bin/
```

## Usage

```bash
# Find repos to work with
my-docs find alloy
# grafana/alloy (2233 matches)
# alloy-rs/alloy (430 matches)

# Create an alias
my-docs alias alloy grafana/alloy

# Search for patterns
my-docs search alloy "prometheus.exporter"
# internal/component/prometheus/exporter/self/self.go:14: Name: "prometheus.exporter.self",

# Read specific files
my-docs cat alloy README.md

# Look up Rust crate symbols
my-docs rust alacritty_terminal KeyboardModes
# (outputs the file containing KeyboardModes, or lists files if multiple matches)

# List configured aliases
my-docs list

# Remove an alias
my-docs remove alloy

# Show config file location
my-docs config

# Install instructions into ~/.claude/CLAUDE.md for AI agents
my-docs install
```

## Commands

| Command | Description |
|---------|-------------|
| `find <query>` | Search for repos by name |
| `alias <name> <owner/repo>` | Create alias for a repo |
| `search <repo> <pattern>` | Search repo via grep.app |
| `cat <repo> <path>` | Fetch and display file from GitHub |
| `rust <crate> <symbol>` | Look up a Rust crate symbol and show its source |
| `list` | Show all configured repo aliases |
| `remove <name>` | Remove a repo alias |
| `config` | Show config file path |
| `install` | Install instructions into ~/.claude/CLAUDE.md |

## Config

Aliases are stored in `~/.config/my-docs/config.json`:

```json
{
  "repos": {
    "alloy": "grafana/alloy",
    "otel": "open-telemetry/opentelemetry-collector"
  }
}
```

## For AI Agents

Run `my-docs install` to add usage instructions to your `~/.claude/CLAUDE.md`. This helps AI agents understand how to use the tool.
