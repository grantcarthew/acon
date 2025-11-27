# AGENTS.md

## Project Overview

**acon** (Atlassian Confluence) is a CLI tool written in Go for managing Atlassian Confluence pages and spaces. It provides bidirectional Markdown conversion (Markdown ↔ Confluence storage format) and supports create, read, update, delete, and list operations for both pages and spaces.

**Technology Stack:**
- Go 1.25.4
- Cobra (CLI framework)
- html-to-markdown v2 (Confluence storage to Markdown)
- Confluence REST API v2

## Setup Commands

```bash
# Install dependencies
go mod download

# Build the binary
go build -o acon

# Install globally (optional)
go install

# Run directly without building
go run main.go [command]
```

## Build and Test Commands

```bash
# Build
go build -o acon

# Build with version info
go build -ldflags "-X main.version=v1.0.0" -o acon

# Format code (required before commits)
gofmt -w .

# Run linter (recommended)
golangci-lint run

# Run with race detector (when tests exist)
go test -race ./...

# Run all tests (when tests exist)
go test ./...

# Generate shell completions
./acon completion bash > /usr/local/etc/bash_completion.d/acon
./acon completion zsh > "${fpath[1]}/_acon"
./acon completion fish > ~/.config/fish/completions/acon.fish
```

## Code Style Guidelines

### Formatting
- **Always run `gofmt -w .` before committing** (or use `goimports`)
- Use tabs for indentation (Go standard)
- Maximum line length: 100-120 characters (soft limit)

### Naming Conventions
- **Exported** (public): `CamelCase` - e.g., `CreatePage`, `Client`
- **Unexported** (private): `camelCase` - e.g., `doRequest`, `appVersion`
- **Package names**: Short, lowercase, singular - e.g., `api`, `config`, `converter`
- **Error variables**: `err` for standard errors, `ErrSomething` for sentinel errors
- **Interfaces**: Follow "accept interfaces, return structs" principle
- **Acronyms**: Use all caps in names - e.g., `APIToken`, `BaseURL`, `PageID`

### Go Idioms
- **Zero value**: Ensure structs work correctly with zero values where possible
- **Error handling**: Check every error, never use `_` to ignore errors
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` to preserve error chain
- **Pointers vs values**: Use pointers for mutation or large structs, values for small immutable data
- **Composition**: Favor struct embedding over inheritance
- **Context**: Pass `context.Context` as first parameter for I/O or long-running operations

### Package Structure

```
acon/
├── cmd/                    # Cobra commands (UI layer)
│   ├── root.go            # Root command and version handling
│   ├── page.go            # Page subcommands (create, view, update, delete, list)
│   └── space.go           # Space subcommands (view, list)
├── internal/
│   ├── api/               # Confluence REST API client (business logic)
│   │   └── client.go      # HTTP client, structs, API methods
│   ├── config/            # Environment variable configuration
│   │   └── config.go      # Config struct and validation
│   └── converter/         # Bidirectional Markdown conversion
│       ├── markdown.go    # Markdown → Confluence storage format
│       └── storage.go     # Confluence storage → Markdown
└── main.go                # Entry point (version injection)
```

### Architecture Principles
- **Separation of concerns**: `cmd/` handles CLI parsing and user interaction, `internal/api/` handles API communication, `internal/converter/` handles format conversion
- **No circular dependencies**: `cmd/` → `internal/*`, never the reverse
- **Keep `internal/` packages focused**: Each package has a single responsibility
- **API client is stateless**: `Client` struct holds credentials, methods are pure operations

## Development Workflow

### Environment Setup
Required environment variables for testing:
```bash
export CONFLUENCE_BASE_URL="https://your-instance.atlassian.net"
export CONFLUENCE_EMAIL="your-email@example.com"
export CONFLUENCE_API_TOKEN="your-api-token"  # or ATLASSIAN_API_TOKEN or JIRA_API_TOKEN
export CONFLUENCE_SPACE_KEY="YOUR_SPACE"      # optional default space
```

Get an API token: https://id.atlassian.com/manage-profile/security/api-tokens

### Testing the CLI
```bash
# Test page listing
./acon page list

# Test page creation with Markdown
echo "# Test Page" | ./acon page create -t "Test Page Title"

# Test page viewing
./acon page view PAGE_ID

# Test with JSON output
./acon page list -j
```

### Branch Management
- Main branch: `main`
- Feature branches: `feature/description` or `fix/description`
- Always create PRs against `main`

### Commit Message Format
Follow conventional commits style:
- `feat: add page deletion command`
- `fix: handle empty space key correctly`
- `docs: update README with examples`
- `refactor: simplify error handling in client`
- `test: add table-driven tests for converter`

## Testing Instructions

### Current State
The project currently has no test files. When adding tests:

1. **Create test files** alongside implementation files:
   - `internal/api/client_test.go`
   - `internal/config/config_test.go`
   - `internal/converter/markdown_test.go`
   - `internal/converter/storage_test.go`

2. **Use table-driven tests** for functions with multiple inputs:
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

3. **Run tests before committing**:
```bash
go test ./...                  # Run all tests
go test -race ./...           # Check for race conditions
go test -cover ./...          # Check coverage
go test -v ./internal/api     # Verbose output for specific package
```

4. **Test coverage priorities**:
   - Error handling paths in `internal/api/client.go`
   - Edge cases in `internal/converter/` (empty input, special characters, nested structures)
   - Config validation in `internal/config/config.go`

## Security Considerations

### API Token Handling
- **Never hardcode API tokens** in code or commit them to git
- Use environment variables exclusively for credentials
- The config package validates tokens are present but never logs them
- HTTP Basic Auth is used (email + API token)

### Input Validation
- All page IDs and space keys are validated as non-empty before API calls
- User-provided Markdown is converted to Confluence storage format (HTML-like)
- Be cautious with user input that could contain XSS vectors when converting to HTML

### HTTP Client
- 30-second timeout on all HTTP requests to prevent hanging
- Always check HTTP status codes (200-299 range)
- Response bodies are always closed with `defer resp.Body.Close()`

### Error Messages
- API errors include status codes and response bodies for debugging
- Be careful not to leak sensitive information in error messages shown to users

## Common Patterns

### Adding a New API Method
1. Define request/response structs in `internal/api/client.go`
2. Add method to `Client` struct following pattern:
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
- Both converters handle CommonMark features (headings, lists, code blocks, links, etc.)

## Troubleshooting

### "API token not set" Error
Ensure one of these environment variables is set:
- `CONFLUENCE_API_TOKEN` (highest priority)
- `ATLASSIAN_API_TOKEN`
- `JIRA_API_TOKEN`

### "space not found" Error
Verify the space key exists and you have access:
```bash
./acon space view YOUR_SPACE_KEY
```

### HTTP 401 Unauthorized
- Verify your email and API token are correct
- Check if the API token has expired
- Ensure you're using the correct Confluence instance URL

### HTTP 404 Not Found
- Verify the page ID or space key is correct
- Check if the page/space has been deleted
- Ensure you have permission to access the resource

### Build Failures
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify dependencies
go mod verify
```

## Reference Documentation

- [Confluence REST API v2](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Code review checklist: `docs/tasks/code-review.md`
