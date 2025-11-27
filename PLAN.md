# my-docs

CLI tool for searching documentation across curated git repositories without cloning them locally.

## Problem

When working with AI agents, they need access to documentation to work effectively. Current options:
- Clone repos locally → clutters disk, manual updates
- Web requests → slow (20s latency), can't grep
- DevDocs/Dash → only mainstream frameworks, can't add custom repos

## Solution

A CLI that:
1. **Searches** via grep.app's undocumented API (fast, indexed, regex support)
2. **Reads files** via raw.githubusercontent.com
3. **Maps short names** to full repo paths via config

## Usage

```bash
# Search for a pattern in a repo
my-docs search alloy "prometheus.exporter"

# Read a specific file
my-docs cat alloy docs/sources/reference/components/prometheus/prometheus.exporter.cloudwatch.md

# List configured repos
my-docs list

# Find repos to alias
my-docs find alloy
# grafana/alloy (1423 matches)
# alloy-rs/alloy (892 matches)

# Add an alias for a repo
my-docs alias alloy grafana/alloy

# Show config location
my-docs config
```

## Architecture

```
┌─────────────┐     search      ┌─────────────────────────────────────┐
│   my-docs   │ ───────────────>│ grep.app/api/search?q=X&f.repo=Y   │
│    CLI      │                 └─────────────────────────────────────┘
│             │     cat         ┌─────────────────────────────────────────┐
│             │ ───────────────>│ raw.githubusercontent.com/owner/repo/… │
└─────────────┘                 └─────────────────────────────────────────┘
       │
       │ config
       ▼
 ~/.config/my-docs/config.json
```

## Config Format

```json
{
  "repos": {
    "alloy": "grafana/alloy",
    "otel": "open-telemetry/opentelemetry-collector",
    "prometheus": "prometheus/prometheus"
  }
}
```

## API Details

### grep.app Search API

**Request:**
```
GET https://grep.app/api/search?q=<query>&f.repo=<owner/repo>&f.lang=<language>&f.path=<path-prefix>
```

Query parameters:
- `q` - search query (required)
- `f.repo` - filter by repository (e.g., `grafana/alloy`)
- `f.lang` - filter by language (e.g., `Markdown`)
- `f.path` - filter by path prefix (e.g., `docs/`)

**Response:**
```json
{
  "time": 422,
  "facets": {
    "path": {
      "buckets": [{ "val": "docs/", "count": 399019 }, ...]
    },
    "repo": {
      "buckets": [{ "val": "grafana/alloy", "count": 1423, "owner_id": "..." }, ...]
    },
    "lang": {
      "buckets": [{ "val": "Markdown", "count": 1295744 }, ...]
    }
  },
  "hits": {
    "total": 11971000,
    "hits": [
      {
        "repo": "tensorflow/tensorflow",
        "branch": "master",
        "path": "path/to/file.cc",
        "content": { "snippet": "<html table with highlighted matches>" },
        "total_matches": "100+"
      }
    ]
  }
}
```

The `content.snippet` is HTML with `<mark>` tags for matches and line numbers in `data-line` attributes.

### GitHub Raw Files

```
GET https://raw.githubusercontent.com/<owner>/<repo>/<branch>/<path>
```

Returns raw file content. Default branch is `main`, fall back to `master`.

## Implementation

- **Language**: Go (cross-platform, single binary)
- **Config**: JSON in ~/.config/my-docs/
- **Dependencies**: Minimal - just HTTP client (JSON is stdlib)

## Commands

| Command | Description |
|---------|-------------|
| `search <repo> <pattern>` | Search repo via grep.app, show matches with context |
| `cat <repo> <path>` | Fetch and display file from GitHub |
| `find <query>` | Search for repos by name (for discovering repos to alias) |
| `list` | Show all configured repos |
| `alias <name> <owner/repo>` | Create alias for a repo |
| `remove <name>` | Remove repo from config |
| `config` | Show config file path |

## Output Format

Search results should be greppable and agent-friendly:

```
docs/reference/components/prometheus.exporter.cloudwatch.md:42: The `prometheus.exporter.cloudwatch` component...
docs/reference/components/prometheus.exporter.cloudwatch.md:58:   cloudwatch_exporter "example" {
```

## Future Ideas

- Version/tag support: `my-docs search alloy@v1.0 "pattern"`
- Cache frequently accessed files locally
- Support GitLab, Bitbucket (different raw URL format)
- Output as JSON for programmatic use
