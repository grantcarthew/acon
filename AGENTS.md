# AGENTS.md

## Project Overview

**acon** (Atlassian Confluence) is a CLI tool for managing Confluence pages and spaces from the terminal. It provides bidirectional Markdown conversion (Markdown ↔ Confluence storage format), enabling a local-first documentation workflow.

**Technology Stack:**

- Go 1.25.4
- Cobra (CLI framework)
- Goldmark (Markdown parser with GFM extension)
- html-to-Markdown v2 (Confluence storage to Markdown)
- Confluence REST API v2

## Setup Commands

```bash
# Install dependencies
go mod download

# Build the binary
go build -o acon

# Run directly without building
go run main.go [command]

# Install globally
go install
```

## Build and Test Commands

```bash
# Build
go build -o acon

# Build with version info
go build -ldflags "-X main.version=v1.0.0" -o acon

# Format code (required before commits)
gofmt -w .

# Run linter
golangci-lint run

# Run tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/api
```

## Code Style Guidelines

### Formatting

- **Always run `gofmt -w .` before committing**
- Use tabs for indentation (Go standard)
- Maximum line length: 100-120 characters (soft limit)

### Naming Conventions

- **Exported** (public): `CamelCase` - e.g., `CreatePage`, `Client`
- **Unexported** (private): `camelCase` - e.g., `doRequest`, `appVersion`
- **Package names**: Short, lowercase, singular - e.g., `api`, `config`, `converter`
- **Error variables**: `err` for standard, `ErrSomething` for sentinels
- **Acronyms**: All caps in names - e.g., `APIToken`, `BaseURL`, `PageID`

### Go Idioms

- Ensure structs work with zero values where possible
- Check every error - never use `_` to ignore errors
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use pointers for mutation or large structs, values for small immutable data
- Favour struct embedding over inheritance
- Pass `context.Context` as first parameter for I/O or long-running operations
- Follow "accept interfaces, return structs" principle

### Package Structure

```
acon/
├── cmd/                    # Cobra commands (UI layer)
│   ├── root.go             # Root command and version handling
│   ├── page.go             # Page subcommands
│   ├── space.go            # Space subcommands
│   └── debug.go            # Debug command for troubleshooting
├── internal/
│   ├── api/                # Confluence REST API client
│   │   └── client.go
│   ├── config/             # Environment variable configuration
│   │   └── config.go
│   └── converter/          # Bidirectional Markdown conversion
│       ├── markdown.go     # Markdown → Confluence storage
│       └── storage.go      # Confluence storage → Markdown
├── .ai/                    # AI agent working files (DDD)
│   ├── projects/           # Project documents
│   ├── design/             # Design records
│   └── tasks/              # Task documentation
├── docs/                   # Human-facing documentation
├── testdata/               # Test fixtures
│   ├── comprehensive-test.md
│   ├── roundtrip-test.sh
│   └── README.md           # Feature support matrix
└── main.go                 # Entry point (version injection)
```

### Architecture Principles

- **Separation of concerns**: `cmd/` handles CLI, `internal/api/` handles API, `internal/converter/` handles conversion
- **No circular dependencies**: `cmd/` → `internal/*`, never the reverse
- **Stateless API client**: `Client` struct holds credentials, methods are pure operations

## Development Workflow

### Environment Setup

Required environment variables:

```bash
export CONFLUENCE_BASE_URL="https://your-instance.atlassian.net"
export CONFLUENCE_EMAIL="your-email@example.com"
export CONFLUENCE_API_TOKEN="your-api-token"  # or ATLASSIAN_API_TOKEN or JIRA_API_TOKEN
export CONFLUENCE_SPACE_KEY="YOUR_SPACE"      # optional default
```

Get an API token: <https://id.atlassian.com/manage-profile/security/api-tokens>

### Testing the CLI

```bash
./acon space list
./acon page list -s MYSPACE
echo "# Test" | ./acon page create -t "Test Page"
./acon page view PAGE_ID
./acon page view PAGE_ID -j
```

### Branch Management

- Main branch: `main`
- Feature branches: `feature/description` or `fix/description`
- Create PRs against `main`

### Commit Message Format

Follow conventional commits:

- `feat: add page deletion command`
- `fix: handle empty space key correctly`
- `docs: update README with examples`
- `refactor: simplify error handling`
- `test: add table-driven tests for converter`
- `chore: update dependencies`

## Testing Instructions

### Running Tests

```bash
go test ./...                  # Run all tests
go test -race ./...            # Check for race conditions
go test -cover ./...           # Check coverage
go test -v ./internal/api      # Verbose output for specific package
```

### Round-Trip Testing

```bash
# Automated round-trip test
./testdata/roundtrip-test.sh PARENT_PAGE_ID

# Manual test
cat testdata/comprehensive-test.md | ./acon page create -t "Test Page" --parent PAGE_ID
./acon page view PAGE_ID
```

### Table-Driven Tests Pattern

```go
func TestConvertMarkdown(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"heading", "# Title", "<h1>Title</h1>"},
        {"bold", "**text**", "<strong>text</strong>"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ConvertMarkdown(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Coverage Priorities

- Error handling paths in `internal/api/client.go`
- Edge cases in `internal/converter/` (empty input, special characters, nested structures)
- Config validation in `internal/config/config.go`

## Security Considerations

### API Token Handling

- Never hardcode API tokens in code or commit them
- Use environment variables exclusively for credentials
- HTTP Basic Auth is used (email + API token)

### Input Validation

- All page IDs and space keys validated as non-empty before API calls
- User-provided Markdown converted to Confluence storage format (HTML-like)
- Be cautious with user input that could contain XSS vectors

### HTTP Client

- 30-second timeout on all HTTP requests
- Always check HTTP status codes (200-299 range)
- Response bodies always closed with `defer resp.Body.Close()`

### Error Messages

- API errors include status codes and response bodies
- Avoid leaking sensitive information in user-facing errors

## Common Patterns

### Adding a New API Method

```go
func (c *Client) MethodName(params) (*Result, error) {
    // Validate inputs
    if strings.TrimSpace(param) == "" {
        return nil, fmt.Errorf("param cannot be empty")
    }

    // Make request
    respBody, err := c.doRequest("METHOD", "/path", requestBody)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }

    // Parse response
    var result Result
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &result, nil
}
```

### Adding a New Command

1. Create command in appropriate file (`cmd/page.go` or `cmd/space.go`)
2. Follow existing patterns:
   - Load config with `config.Load()`
   - Create API client with `api.NewClient()`
   - Handle `-j/--json` flag for JSON output
   - Provide human-readable output by default
3. Add to parent command in `init()` function
4. Update README.md with usage examples

### Markdown Conversion

- **To Confluence**: Use `internal/converter/markdown.go`
- **From Confluence**: Use `internal/converter/storage.go`
- Both handle CommonMark + GFM features (tables, task lists, strikethrough)
- See `testdata/README.md` for feature support matrix and known limitations

## Troubleshooting

### "API token not set" Error

Set one of: `CONFLUENCE_API_TOKEN`, `ATLASSIAN_API_TOKEN`, or `JIRA_API_TOKEN`

### "space not found" Error

```bash
./acon space view YOUR_SPACE_KEY
```

### HTTP 401 Unauthorized

- Verify email and API token are correct
- Check if token has expired or been revoked
- Ensure correct Confluence instance URL

### HTTP 404 Not Found

- Verify page ID or space key is correct
- Check if resource still exists
- Ensure you have permission to access it

### Build Failures

```bash
go clean -modcache
go mod download
go mod verify
go build -o acon
```

## Release Process

See `.ai/tasks/release-process.md` for the complete release workflow including:

1. Pre-release validation
2. Version tagging
3. GitHub Release creation
4. Homebrew tap update

## Reference Documentation

- [Confluence REST API v2](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/)
- [Confluence Storage Format](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Goldmark Markdown Parser](https://github.com/yuin/goldmark)
- [html-to-Markdown](https://github.com/JohannesKaufmann/html-to-markdown)
- Code review checklist: `.ai/tasks/code-review.md`
- Feature support matrix: `testdata/README.md`
