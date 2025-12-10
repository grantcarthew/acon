# Project: Add Search Command to acon

## Overview

Add a new `acon search` command to search Confluence content using the v1 REST API search endpoint with CQL (Confluence Query Language).

## Background

- Confluence REST API v2 does not have a search endpoint
- Search is only available via v1: `GET /wiki/rest/api/search?cql=<query>`
- The existing `acon` codebase uses v2 for other operations, but v1 can coexist (same auth)

## CLI Specification

```
acon search <query> [flags]
```

### Arguments

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `query` | string | Yes | Full-text search query |

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--space` | `-s` | []string | (all) | Filter to space(s), comma-separated |
| `--title` | `-t` | bool | false | Search titles only (instead of full-text) |
| `--label` | `-L` | []string | (none) | Filter by label(s), comma-separated |
| `--type` | `-T` | []string | (all) | Content type(s), comma-separated |
| `--limit` | `-l` | int | 25 | Maximum results to return |
| `--excerpt` | `-e` | bool | false | Show excerpt with results |
| `--json` | `-j` | bool | false | Output as JSON |

### Valid Content Types

`page`, `blogpost`, `attachment`, `comment`, `whiteboard`, `database`, `folder`

### Usage Examples

```bash
# Simple search
acon search "API documentation"

# Search in specific spaces
acon search "onboarding" -s HR,ENG

# Search pages only
acon search "release notes" -T page

# Search by label
acon search "quarterly" -L Q1,shipped

# Title-only search
acon search "roadmap" -t

# Show excerpts
acon search "meeting notes" -e

# JSON output
acon search "API" -j

# Combined filters
acon search "deployment" -s DEV,OPS -T page,blogpost -l 50
```

## API Specification

### Endpoint

```
GET /wiki/rest/api/search
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `cql` | string | CQL query (required) |
| `limit` | int | Results per page (max ~25) |
| `cursor` | string | Pagination cursor |

### CQL Query Building

Build CQL from CLI flags:

| Flag | CQL Clause | Example |
|------|------------|---------|
| `query` (default) | `text ~ "query"` | `text ~ "meeting"` |
| `query` + `--title` | `title ~ "query"` | `title ~ "roadmap"` |
| `--space` | `space IN (A,B)` | `space IN (DEV,PROD)` |
| `--type` | `type IN (a,b)` | `type IN (page,blogpost)` |
| `--label` | `label IN (a,b)` | `label IN (Q1,shipped)` |

Clauses are joined with `AND`.

**Example built query:**
```
text ~ "deployment" AND space IN (DEV,OPS) AND type IN (page,blogpost)
```

### Query Escaping

User input must be escaped for CQL:
- Double quotes inside query: `"` → `\"`
- Backslashes: `\` → `\\`

### Response Structure

```json
{
  "results": [
    {
      "content": {
        "id": "123456",
        "type": "page",
        "status": "current",
        "title": "Meeting Notes",
        "space": {
          "key": "DEV",
          "name": "Development"
        }
      },
      "excerpt": "...highlighted <em>search</em> terms...",
      "url": "/wiki/spaces/DEV/pages/123456",
      "lastModified": "2025-01-15T10:30:00Z",
      "friendlyLastModified": "yesterday"
    }
  ],
  "limit": 25,
  "size": 25,
  "totalSize": 100,
  "_links": {
    "next": "/rest/api/search?cql=...&cursor=abc123"
  }
}
```

### Pagination

- API returns max ~25 results per request
- Use `_links.next` cursor for subsequent pages
- Auto-paginate until `--limit` is reached (matching `ListPages` behaviour)

## Output Formats

### Default Table Output

```
ID        TITLE                          SPACE  TYPE
123456    Meeting Notes                  DEV    page
789012    API Documentation              ENG    page
345678    Q1 Release Blog                HR     blogpost
```

### With `--excerpt`

```
123456  Meeting Notes (DEV, page)
        ...highlighted search terms in context...

789012  API Documentation (ENG, page)
        ...this document describes the API...
```

- Strip HTML tags from excerpt (e.g., `<em>` highlighting)
- Truncate long excerpts to reasonable terminal width

### With `--json`

Output clean struct array:

```json
[
  {
    "id": "123456",
    "title": "Meeting Notes",
    "space": "DEV",
    "type": "page",
    "excerpt": "...search terms in context...",
    "url": "/wiki/spaces/DEV/pages/123456",
    "lastModified": "2025-01-15T10:30:00Z"
  }
]
```

### No Results

```
No results found
```

## Implementation

### New Files

- `cmd/search.go` - Search command implementation

### Modified Files

- `internal/api/client.go` - Add search types and method

### Go Structs

```go
// In internal/api/client.go

// SearchResult represents a single search result
type SearchResult struct {
    ID           string `json:"id"`
    Title        string `json:"title"`
    Type         string `json:"type"`
    SpaceKey     string `json:"space"`
    Excerpt      string `json:"excerpt,omitempty"`
    URL          string `json:"url"`
    LastModified string `json:"lastModified"`
}

// SearchResponse represents the v1 search API response
type SearchResponse struct {
    Results []struct {
        Content struct {
            ID    string `json:"id"`
            Type  string `json:"type"`
            Title string `json:"title"`
            Space struct {
                Key string `json:"key"`
            } `json:"space"`
        } `json:"content"`
        Excerpt      string `json:"excerpt"`
        URL          string `json:"url"`
        LastModified string `json:"lastModified"`
    } `json:"results"`
    Links struct {
        Next string `json:"next"`
    } `json:"_links"`
}

// Search performs a CQL search and returns results
func (c *Client) Search(ctx context.Context, cql string, limit int) ([]SearchResult, error)
```

### CQL Builder Function

```go
// In cmd/search.go

func buildCQL(query string, titleOnly bool, spaces, types, labels []string) string
```

### Escaping Function

```go
// In cmd/search.go or internal/api/client.go

func escapeCQL(s string) string
```

## Error Handling

| Scenario | Behaviour |
|----------|-----------|
| Empty query | Error: "query is required" |
| Invalid CQL (400) | Display API error message |
| Auth failure (401/403) | Display auth error |
| No results | Display "No results found" |
| Network error | Display connection error |

## Testing

### Unit Tests

**CQL Building** (`cmd/search_test.go`):
- Basic query: `"meeting"` → `text ~ "meeting"`
- Title only: `"meeting"` + `-t` → `title ~ "meeting"`
- With space: `"meeting"` + `-s DEV` → `text ~ "meeting" AND space = DEV`
- Multiple spaces: `-s DEV,PROD` → `space IN (DEV,PROD)`
- Multiple types: `-T page,blogpost` → `type IN (page,blogpost)`
- Multiple labels: `-L Q1,Q2` → `label IN (Q1,Q2)`
- Combined filters: all flags together
- Escaping quotes: `"test \"quoted\""` → `text ~ "test \"quoted\""`

**Excerpt Stripping** (`cmd/search_test.go`):
- Remove `<em>` tags
- Handle nested tags
- Preserve text content

### Integration Tests

**API Client** (`internal/api/client_test.go`):
- Mock HTTP responses for search endpoint
- Pagination handling (multi-page results)
- Error responses (400, 401, 403)

### Manual Testing

```bash
# Basic search
./acon search "test"

# With filters
./acon search "test" -s MYSPACE -T page -l 10

# Excerpt output
./acon search "test" -e

# JSON output
./acon search "test" -j

# No results
./acon search "xyznonexistent12345"
```

## Implementation Order

1. Add `SearchResult` and `SearchResponse` structs to `internal/api/client.go`
2. Add `Search` method to `Client` with pagination
3. Create `cmd/search.go` with command definition
4. Implement `buildCQL` function
5. Implement `escapeCQL` function
6. Implement table output
7. Implement excerpt output (with HTML stripping)
8. Implement JSON output
9. Write unit tests for CQL building
10. Write unit tests for excerpt stripping
11. Write integration tests for API client
12. Manual testing
13. Update README.md with search examples

## References

- [Confluence REST API v1 Search](https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-search/)
- [CQL Reference](https://developer.atlassian.com/cloud/confluence/advanced-searching-using-cql/)
- [CQL Fields](https://developer.atlassian.com/cloud/confluence/cql-fields/)
