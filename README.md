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
sudo cp my-docs /usr/local/bin/
```

## Usage

```bash
# Find repos to work with
my-docs find alloy
# grafana/alloy (2233 matches)
# alloy-rs/alloy (430 matches)

# Search for patterns (use owner/repo format)
my-docs search grafana/alloy "prometheus.exporter"
# internal/component/prometheus/exporter/self/self.go:14: Name: "prometheus.exporter.self",

# Read specific files
my-docs cat grafana/alloy README.md

# Look up Rust crate symbols
my-docs rust alacritty_terminal KeyboardModes
# (outputs the file containing KeyboardModes, or lists files if multiple matches)

# Install instructions into ~/.claude/CLAUDE.md for AI agents
my-docs install
```

## Commands

| Command | Description |
|---------|-------------|
| `find <query>` | Search for repos by name |
| `search [owner/repo] <pattern>` | Search repo via grep.app (omit repo to search all) |
| `cat <owner/repo> <path>` | Fetch and display file from GitHub |
| `rust <crate> <symbol>` | Look up a Rust crate symbol and show its source |
| `install` | Install instructions into ~/.claude/CLAUDE.md |

## For AI Agents

Run `my-docs install` to add usage instructions to your `~/.claude/CLAUDE.md`. This helps AI agents understand how to use the tool.
