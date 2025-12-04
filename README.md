# acon

**A fast, simple command-line tool for Atlassian Confluence** - because the web editor is not it.

Manage Confluence pages and spaces from your terminal with full Markdown support. Create, edit, and publish documentation without ever opening a browser.

```bash
# Write documentation in Markdown, publish to Confluence
echo "# API Docs\n\nYour content here" | acon page create -t "API Documentation"

# Edit pages with your favorite editor
acon page view 123456 > docs.md
vim docs.md
acon page update 123456 -f docs.md
```

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/grantcarthew/acon)](https://goreportcard.com/report/github.com/grantcarthew/acon)

## Features

- **Bidirectional Markdown conversion** - Write in Markdown, view in Markdown, never touch HTML
- **Full page management** - Create, view, update, delete, and list pages
- **Space operations** - View and list Confluence spaces
- **JSON output** - Perfect for scripting and automation
- **Environment-based config** - No config files, works with existing Atlassian tokens
- **Shell completion** - Bash, Zsh, and Fish support

## Quick Start

### Installation

**Homebrew** (macOS/Linux):
```bash
# Add the tap first
brew tap grantcarthew/tap

# Then install
brew install acon

# Or do it in one command
brew install grantcarthew/tap/acon
```

*Check out [my other Homebrew packages](https://github.com/grantcarthew/homebrew-tap) in the tap!*

**Go install**:
```bash
go install github.com/grantcarthew/acon@latest
```

**From source**:
```bash
git clone https://github.com/grantcarthew/acon.git
cd acon
go build -o acon
sudo mv acon /usr/local/bin/
```

### Configuration

Set the following environment variables:

```bash
export CONFLUENCE_BASE_URL="https://your-instance.atlassian.net"
export CONFLUENCE_EMAIL="your-email@example.com"
export CONFLUENCE_API_TOKEN="your-api-token"
export CONFLUENCE_SPACE_KEY="YOUR_SPACE"  # (Optional)
```

Get an API token at: https://id.atlassian.com/manage-profile/security/api-tokens

**Note**: The same API token works for Confluence and Jira. You can use `CONFLUENCE_API_TOKEN`, `ATLASSIAN_API_TOKEN`, or `JIRA_API_TOKEN`.

### First Commands

```bash
# List spaces you have access to
acon space list

# List pages in a space
acon page list -s MYSPACE

# Create a page from Markdown
cat README.md | acon page create -t "My Documentation"

# View a page (outputs Markdown)
acon page view 123456789
```

## Usage

### Commands Overview

```
acon [command]

Available Commands:
  page        Manage Confluence pages
  space       Manage Confluence spaces
  completion  Generate shell completion
  help        Help about any command

Flags:
  -h, --help      help for acon
  -v, --version   Print version
```

### Page Commands

#### `acon page create`

Create a new Confluence page from Markdown.

```bash
acon page create -t TITLE [flags]

Flags:
  -f, --file string     Markdown file to read (default: stdin)
  -j, --json           Output JSON instead of human-readable format
  -m, --message string Version message
  -p, --parent string  Parent page ID
  -s, --space string   Space key (uses CONFLUENCE_SPACE_KEY if not set)
  -t, --title string   Page title (required)
```

**Examples**:

```bash
# Create from file
acon page create -t "API Documentation" -f api-docs.md

# Create from stdin
echo "# Hello World" | acon page create -t "Hello"

# Create in specific space with parent
acon page create -t "Child Page" -f content.md -s MYSPACE -p 123456

# JSON output for scripting
acon page create -t "Title" -f content.md -j
```

#### `acon page view`

View a Confluence page (outputs Markdown).

```bash
acon page view PAGE_ID [flags]

Arguments:
  PAGE_ID   Confluence page ID (required)

Flags:
  -j, --json   Output JSON instead of Markdown
```

**Examples**:

```bash
# View as Markdown
acon page view 123456789

# View as JSON
acon page view 123456789 -j

# Save to file
acon page view 123456789 > local-copy.md

# Edit and update workflow
acon page view 123456789 > docs.md
vim docs.md
acon page update 123456789 -f docs.md
```

#### `acon page update`

Update an existing Confluence page.

```bash
acon page update PAGE_ID [flags]

Arguments:
  PAGE_ID   Confluence page ID (required)

Flags:
  -f, --file string     Markdown file to read (default: stdin)
  -j, --json           Output JSON instead of human-readable format
  -m, --message string Version message (appears in page history)
  -t, --title string   New page title (optional, keeps existing if not set)
```

**Examples**:

```bash
# Update content from file
acon page update 123456789 -f updated-docs.md

# Update content from stdin
cat updated-docs.md | acon page update 123456789

# Update title and content
acon page update 123456789 -t "New Title" -f updated-docs.md

# Add version message
acon page update 123456789 -f docs.md -m "Updated API endpoints"
```

#### `acon page delete`

Delete a Confluence page.

```bash
acon page delete PAGE_ID

Arguments:
  PAGE_ID   Confluence page ID (required)
```

**Example**:

```bash
acon page delete 123456789
```

#### `acon page list`

List pages in a Confluence space or children of a specific page.

```bash
acon page list [flags]

Flags:
      --desc            Sort in descending order
  -j, --json            Output JSON instead of human-readable format
  -l, --limit int       Maximum number of pages to return (default: 25)
  -p, --parent string   Parent page ID (list children of this page)
      --sort string     Sort order (see below)
  -s, --space string    Space key (uses CONFLUENCE_SPACE_KEY if not set)
```

**Sort options**:
- With `--parent`: `web` (default), `title`, `created`, `modified`, `id`
- Without `--parent`: `title`, `created`, `modified`, `id`

The `web` sort matches the manual page order in Confluence's web interface.

**Examples**:

```bash
# List pages in default space
acon page list

# List pages in specific space
acon page list -s MYSPACE

# List child pages of a parent
acon page list --parent 123456789

# Sort by creation date (newest first)
acon page list --parent 123456789 --sort created --desc

# Sort by title
acon page list -s MYSPACE --sort title

# Reverse default order
acon page list --parent 123456789 --desc

# Limit results
acon page list -l 10

# JSON output
acon page list -j
```

### Space Commands

#### `acon space view`

View details about a Confluence space.

```bash
acon space view SPACE_KEY [flags]

Arguments:
  SPACE_KEY   Confluence space key (required)

Flags:
  -j, --json   Output JSON instead of human-readable format
```

**Examples**:

```bash
# View space details
acon space view MYSPACE

# JSON output
acon space view MYSPACE -j
```

#### `acon space list`

List Confluence spaces you have access to.

```bash
acon space list [flags]

Flags:
  -j, --json       Output JSON instead of human-readable format
  -l, --limit int  Maximum number of spaces to return (default: 25)
```

**Examples**:

```bash
# List all accessible spaces
acon space list

# Limit results
acon space list -l 10

# JSON output for scripting
acon space list -j
```

### Shell Completion

Generate shell completion scripts for Bash, Zsh, or Fish.

```bash
acon completion [bash|zsh|fish]
```

**Installation**:

```bash
# Bash (add to ~/.bashrc or /etc/bash_completion.d/)
acon completion bash > /usr/local/etc/bash_completion.d/acon

# Zsh (add to ~/.zshrc or fpath)
acon completion zsh > "${fpath[1]}/_acon"

# Fish
acon completion fish > ~/.config/fish/completions/acon.fish
```

## Markdown Support

acon provides seamless bidirectional Markdown conversion:

### When Creating/Updating Pages (Markdown → Confluence)

Your Markdown is automatically converted to Confluence storage format:

| Markdown | Result |
|----------|--------|
| `# Heading 1` | Heading 1 |
| `## Heading 2` | Heading 2 |
| `**bold**` | Bold text |
| `*italic*` | Italic text |
| `` `code` `` | Inline code |
| ` ```language ` | Code block |
| `[text](url)` | Hyperlink |
| `- item` or `* item` | Unordered list |
| `1. item` | Ordered list |
| `> quote` | Blockquote |

### When Viewing Pages (Confluence → Markdown)

Confluence storage format is converted back to clean, readable Markdown - perfect for editing locally.

**Supported conversions**:
- Tables
- Nested lists
- Code blocks with syntax highlighting
- Links (internal and external)
- Strikethrough
- All CommonMark features

## Examples

### Create Documentation from Local Files

```bash
# Single file
acon page create -t "Installation Guide" -f docs/install.md

# Batch create multiple pages
for file in docs/*.md; do
  title=$(basename "$file" .md)
  acon page create -t "$title" -f "$file" -s DOCS
done
```

### Edit-Update Workflow

```bash
# Download page to local file
acon page view 123456789 > docs.md

# Edit with your favourite editor
vim docs.md
# or code docs.md
# or emacs docs.md

# Upload changes
acon page update 123456789 -f docs.md -m "Updated via acon"
```

### Scripting with JSON Output

```bash
# Get page ID by title (requires jq)
PAGE_ID=$(acon page list -j | jq -r '.results[] | select(.title=="My Page") | .id')

# Bulk operations
acon space list -j | jq -r '.results[].key' | while read space; do
  echo "Pages in $space:"
  acon page list -s "$space" -l 5
done
```

### CI/CD Integration

```yaml
# Deploy docs to Confluence from GitHub Actions
name: Deploy Docs
on:
  push:
    branches: [main]
    paths: [docs/**]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/grantcarthew/acon@latest
      - name: Update Confluence
        env:
          CONFLUENCE_BASE_URL: ${{ secrets.CONFLUENCE_BASE_URL }}
          CONFLUENCE_EMAIL: ${{ secrets.CONFLUENCE_EMAIL }}
          CONFLUENCE_API_TOKEN: ${{ secrets.CONFLUENCE_API_TOKEN }}
        run: |
          acon page update ${{ vars.DOCS_PAGE_ID }} -f docs/README.md -m "Deploy from ${{ github.sha }}"
```

### Backup Confluence Pages

```bash
#!/bin/bash
# backup-confluence.sh - Backup all pages in a space

SPACE="MYSPACE"
BACKUP_DIR="confluence-backup-$(date +%Y%m%d)"

mkdir -p "$BACKUP_DIR"

# Get all page IDs
PAGE_IDS=$(acon page list -s "$SPACE" -j | jq -r '.results[].id')

# Download each page
for id in $PAGE_IDS; do
  title=$(acon page view "$id" -j | jq -r '.title' | tr ' ' '-')
  echo "Backing up: $title"
  acon page view "$id" > "$BACKUP_DIR/${id}-${title}.md"
done

echo "Backup complete: $BACKUP_DIR"
```

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CONFLUENCE_BASE_URL` | Yes | Your Confluence instance URL | `https://company.atlassian.net` |
| `CONFLUENCE_EMAIL` | Yes | Your email address | `user@example.com` |
| `CONFLUENCE_API_TOKEN` | Yes* | Atlassian API token | Get from [API tokens](https://id.atlassian.com/manage-profile/security/api-tokens) |
| `ATLASSIAN_API_TOKEN` | Yes* | Alternative to CONFLUENCE_API_TOKEN | Same token works for all Atlassian products |
| `JIRA_API_TOKEN` | Yes* | Alternative (if you already have Jira token) | Same token works for Confluence |
| `CONFLUENCE_SPACE_KEY` | No | Default space key (avoids `-s` flag) | `DOCS`, `TEAM`, etc. |

**Note**: Only one API token variable is required. acon checks in order: `CONFLUENCE_API_TOKEN` → `ATLASSIAN_API_TOKEN` → `JIRA_API_TOKEN`.

## Development

### Building from Source

```bash
git clone https://github.com/grantcarthew/acon.git
cd acon
go mod download
go build -o acon
```

### Project Structure

```
acon/
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command and version
│   ├── page.go            # Page subcommands
│   └── space.go           # Space subcommands
├── internal/
│   ├── api/               # Confluence REST API client
│   │   └── client.go
│   ├── config/            # Environment variable loader
│   │   └── config.go
│   └── converter/         # Bidirectional Markdown conversion
│       ├── markdown.go    # Markdown → Confluence storage
│       └── storage.go     # Confluence storage → Markdown
├── docs/                  # Documentation and processes
│   └── tasks/
│       ├── code-review.md
│       └── release-process.md
├── main.go                # Entry point
├── AGENTS.md              # Agent development guidelines
└── README.md
```

### Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - Confluence storage to Markdown converter

### Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run `gofmt -w .` before committing
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

See [AGENTS.md](AGENTS.md) for detailed development guidelines.

## Troubleshooting

### "API token not set" Error

Ensure one of these environment variables is set:
- `CONFLUENCE_API_TOKEN`
- `ATLASSIAN_API_TOKEN`
- `JIRA_API_TOKEN`

```bash
# Check if set
echo $CONFLUENCE_API_TOKEN

# Set it
export CONFLUENCE_API_TOKEN="your-token-here"
```

### "space not found" Error

Verify the space exists and you have access:

```bash
# List all spaces you can access
acon space list

# Try viewing the specific space
acon space view YOUR_SPACE_KEY
```

### HTTP 401 Unauthorized

- Verify your email is correct
- Check if API token is valid (they don't expire but can be revoked)
- Ensure you're using the correct Confluence instance URL
- Create a new API token if needed

### HTTP 404 Not Found

- Verify page ID or space key is correct
- Check if the page/space still exists (might have been deleted)
- Ensure you have permission to access the resource

### Build Errors

```bash
# Clean and rebuild
go clean -modcache
go mod download
go mod verify
go build -o acon
```

## FAQ

**Q: Can I use this with Confluence Server/Data Center?**

A: Currently acon targets Confluence Cloud (REST API v2). Server/Data Center support could be added - PRs welcome!

**Q: Does this support attachments?**

A: Not yet. File attachment support is planned for a future release.

**Q: Can I migrate content from other systems?**

A: Yes! If you can convert your content to Markdown, acon can publish it to Confluence. Great for migrating from GitHub wikis, GitBook, MkDocs, etc.

**Q: Is there a way to search pages?**

A: Not directly, but you can use JSON output with `jq` to filter results from `acon page list`.

**Q: Can I use this in CI/CD?**

A: Absolutely! See the [CI/CD Integration example](#cicd-integration) above.

## Related Projects

- [jira-cli](https://github.com/ankitpokhrel/jira-cli) - CLI for Jira (pairs great with acon)
- [pandoc](https://pandoc.org/) - Universal document converter
- [glab](https://gitlab.com/gitlab-org/cli) - GitLab CLI

## License

This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

See [LICENSE](LICENSE) for full details.

## Author

**Grant Carthew** - [GitHub](https://github.com/grantcarthew)

---

**Made with ❤️ for everyone who hates editing docs in a web browser.**

*Star this repo if acon makes your documentation workflow better!*
