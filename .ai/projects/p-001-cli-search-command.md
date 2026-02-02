# P-001: Search Command Implementation

- Status: Pending
- Started: (not yet started)
- Completed: (not yet completed)

## Overview

Implement a search command for acon that enables users to search Confluence content using CQL (Confluence Query Language). The implementation provides a user-friendly CLI interface with simple flags for common searches, while exposing raw CQL for advanced queries.

This project addresses GitHub Issue #1 and implements the design specified in DR-001.

## Goals

1. Add v1 API support to the client (search only available in v1)
2. Implement CQL query builder that constructs queries from CLI flags
3. Create search command with simple flags for common use cases
4. Support raw CQL for advanced queries
5. Implement cursor-based pagination for search results
6. Provide comprehensive test coverage (cannot test on real instance)

## Scope

In Scope:

- v1 API client method (doRequestV1) for search endpoint
- Search command with positional text query
- Simple flags: --title, --label, --creator (with 'me' alias)
- Standard flags: -s/--space, -l/--limit, --type, -j/--json
- Raw CQL support via --cql flag
- CQL query construction and URL encoding
- Cursor-based pagination
- Standard output format (Title, Space, URL, Excerpt, Modified)
- JSON output format
- Comprehensive unit and integration tests
- Error handling including rate limiting (429)

Out of Scope:

- Date filters (too complex, use raw CQL)
- Interactive search refinement
- Search result caching
- Search history
- Automatic retry on rate limiting (fail fast)
- Body-only search (not supported by CQL)
- OR logic combinations (use raw CQL)

## Success Criteria

- [ ] doRequestV1 method added to internal/api/client.go
- [ ] Search types and methods in internal/api/search.go
- [ ] Search command in cmd/search.go following acon patterns
- [ ] Simple flags work: text, title, label, creator
- [ ] 'me' alias resolves to currentUser() in CQL
- [ ] Raw CQL flag works for advanced queries
- [ ] All search criteria combine with AND logic
- [ ] Default behavior searches pages only (type=page)
- [ ] Cursor-based pagination implemented
- [ ] Standard output format matches design
- [ ] JSON output includes all search response fields
- [ ] Comprehensive tests (15-20 cases minimum)
- [ ] Tests cover CQL construction, URL encoding, pagination, errors
- [ ] Rate limiting fails fast with informative error message
- [ ] All tests pass: go test ./...
- [ ] Documentation updated in README.md
- [ ] GitHub Issue #1 closed

## Deliverables

Code:

- internal/api/search.go - Search types, methods, CQL builder
- internal/api/search_test.go - Comprehensive search tests
- cmd/search.go - Search command implementation
- cmd/search_test.go - Command tests

Documentation:

- README.md - Updated with search command usage
- Closes GitHub Issue #1

Design:

- dr-001-cli-search-command.md - Already created

## Current State

Relevant codebase context:

API Client:

- internal/api/client.go has doRequest method for v2 API
- Uses http.Client with 30-second timeout
- Basic auth with email + API token
- Context-aware requests
- Returns response body as []byte
- v1 API will need separate doRequestV1 method with path: /wiki/rest/api/ (not /wiki/api/v2/)

Testing Patterns:

- Table-driven tests using httptest.NewServer
- Tests cover success cases, error cases (400, 401, 404, 500), edge cases
- Context cancellation tested
- Pagination tested with multiple pages
- Example: client_test.go has 983 lines of comprehensive tests
- All existing tests pass with Go 1.25.6

Command Structure:

- Commands use Cobra framework
- Flags defined in init() function
- Standard pattern: initClient(), call API method, handle JSON or human output
- Flags: -j/--json for JSON output, -s/--space for space, -l/--limit for limits
- Human output uses fmt.Printf with labeled fields

Error Handling:

- Errors wrapped with context using fmt.Errorf
- API errors include status code and response body
- Input validation before API calls (empty string checks)

v1 Search API Response Structure:

- Response has nested structure: results[].content contains page details
- Key fields: title, excerpt, url, lastModified (top level of each result)
- content object has: id, type, status, space.key, space.name
- Pagination via _links.next (cursor-based, not offset)
- Response includes: start, limit, size, totalSize, cqlQuery, searchDuration
- Excerpt can be controlled via query param: highlight, indexed, none

Dependencies (current with updates available):

- cobra v1.10.1 (v1.10.2 available)
- goldmark v1.7.13 (v1.7.16 available)
- html-to-markdown v2.5.0 (current)

Rate Limiting (Confluence v1 API):

- Returns 429 with Retry-After header when rate limited
- Recommended approach: fail fast, display error with Retry-After value
- Do not implement automatic retry to avoid masking rate limit issues
- External REST API requests are rate limited, internal Confluence actions are not

## Decision Points

All decisions confirmed and locked down:

1. Excerpt Strategy: APPROVED
   - Use excerpt=indexed parameter in API request
   - Shows excerpt text without HTML highlighting markup
   - Clean text output for human-readable format

2. Pagination Behavior: APPROVED
   - Display "Showing X of Y results" message using totalSize from response
   - Respect --limit flag (fetch up to limit, then stop)
   - One API call, no automatic fetching of additional pages
   - Message format: "Showing 25 of 150 results" or omit if all results returned

3. URL Output Format: APPROVED
   - Construct full URL by prepending baseURL to relative URL
   - Example: baseURL + "/wiki/spaces/DEV/pages/123456/Page+Title"
   - Clickable in modern terminals
   - Consistent with other acon commands

4. Space Key in Output: APPROVED
   - Show space key in parentheses after title: "Page Title (DEV)"
   - Compact single-line format
   - Matches design record example
   - Extract from response: results[].content.space.key

5. 'me' Alias Implementation: APPROVED
   - Convert "me" string to currentUser() CQL function
   - Native CQL function, no email lookup needed
   - Example: --creator me generates "creator = currentUser()"
   - Simple string replacement in CQL builder

## Technical Approach

Phase 1: v1 API Support

- Add doRequestV1 method to internal/api/client.go
- Use path prefix: /wiki/rest/api/ (v1) instead of /wiki/api/v2/ (v2)
- Mirror doRequest pattern: same auth, timeout, error handling
- Share authentication (Basic Auth: email + API token) and HTTP client with v2
- Return []byte response body like doRequest
- Handle same status codes: 200-299 success, others error
- Note: No changes to existing doRequest or v2 endpoints

Phase 2: Search Types and CQL Builder

- Create internal/api/search.go with:
  - SearchResult struct (top-level fields: title, excerpt, url, lastModified)
  - SearchContent struct (nested: id, type, status, space with key/name)
  - SearchResponse struct (results, start, limit, size, totalSize, cqlQuery, searchDuration, _links)
  - SearchPaginationLinks struct (_links.next with cursor URL)
  - BuildCQL function that constructs CQL from search parameters
  - Search method on Client that calls doRequestV1 with GET /wiki/rest/api/search
- CQL builder logic:
  - Default to type=page (unless --type specified)
  - Combine all conditions with AND
  - Use url.QueryEscape for complete CQL string after building
  - Handle 'me' alias by converting to currentUser() in CQL
  - Text search: text ~ "query" (searches title, body, labels)
  - Title search: title ~ "query" (title only)
  - Label search: label = "value" (exact match, not fuzzy)
  - Space: space = KEY (exact space key match)
  - Creator: creator = currentUser() or creator = email
- URL construction: /wiki/rest/api/search?cql={encoded_query}&limit={limit}
- Response parsing: handle nested content object, extract space.key for output

Phase 3: Search Command

- Create cmd/search.go following page.go patterns
- Define flags in init() function
- Implement RunE function:
  - Parse flags and positional argument
  - Build CQL query or use raw --cql
  - Add excerpt=indexed to API request
  - Call client.Search with context
  - Format output (JSON or human-readable)
  - Display "Showing X of Y results" message if totalSize > size
- Standard output format:
  - Title (SPACE_KEY)
  - Full URL (baseURL + relative URL)
  - Excerpt text (no HTML highlighting)
  - Modified: date
  - ---
- Validate mutually exclusive options (positional text with --cql)

Phase 4: Comprehensive Testing

- internal/api/search_test.go:
  - Test CQL construction for all flag combinations
  - Test URL encoding of special characters
  - Test pagination with cursors
  - Test all error responses (400, 401, 404, 429, 500)
  - Test empty results
  - Test response parsing
  - Test 'me' alias conversion
- cmd/search_test.go:
  - Test flag parsing
  - Test output formatting
  - Test error handling

Phase 5: Documentation

- Update README.md with search command examples
- Include simple and advanced usage
- Document --cql flag for complex queries

## Testing Strategy

Unit Tests:

- CQL builder function with table-driven tests
- Test all flag combinations produce correct CQL
- Test URL encoding edge cases (quotes, spaces, special chars)
- Test 'me' alias conversion to currentUser()
- Mock v1 API responses using httptest.NewServer
- Test response parsing with various field combinations

Integration Tests:

- Full command execution with mocked server
- Test flag combinations end-to-end
- Test output formatting (both JSON and human)
- Test error messages

Edge Cases:

- Empty search results
- Malformed JSON responses
- Missing optional fields in response
- Rate limiting (429) with and without Retry-After header
- Context cancellation
- Invalid CQL syntax (400 error)

Test Coverage Target:

- Minimum 15-20 test cases
- Cover all success paths
- Cover all error paths
- Cover pagination scenarios
- Match quality of existing client_test.go

Cannot Test (no real Confluence instance):

- Actual v1 API responses
- Real CQL query execution
- Actual pagination cursors
- Real rate limiting behavior

Mitigation: Comprehensive mocking ensures correct implementation.

## Related DRs

Implements:

- dr-001-cli-search-command.md - Search command CLI interface design

## Notes

Implementation Considerations:

- v1 API uses different path prefix than v2 (/wiki/rest/api/ vs /wiki/api/v2/)
- v1 response structure differs from v2 (nested content object with search-specific fields)
- Response parsing: results[].content.space.key needed for output format
- Cursor pagination is opaque string from _links.next, not numeric offset
- CQL query must be URL-encoded after construction using url.QueryEscape
- Rate limiting should fail fast with clear error message including Retry-After value
- No automatic retry to avoid masking rate limit issues
- Full URL construction: baseURL + relative url (url field starts with /wiki/spaces/)
- Excerpt strategy: use excerpt=indexed parameter for clean text without HTML markup
- Pagination message: show "Showing X of Y results" when size < totalSize
- 'me' alias: convert string "me" to CQL function currentUser() in query builder

CQL Special Characters:

- Must URL-encode: spaces, quotes, ~, !, =, etc.
- Use url.QueryEscape for complete CQL string after building query
- Example: "text ~ \"api docs\"" becomes "text+%7E+%22api+docs%22"
- Reserved characters in CQL: " ", -, ~, +, ., ,, ;, ?, |, *, /, %, ^, $, #, @, [, ]

Testing Without Real Instance:

- Mock all API responses using httptest.NewServer (same as existing tests)
- Use example responses from Confluence v1 API docs for realistic structure
- Test CQL construction logic independently from API calls
- Verify URL encoding with known test cases (special chars, quotes, spaces)
- Test nested response parsing (content.space.key extraction)
- Comprehensive error handling tests (400, 401, 404, 429, 500)
- Test cursor pagination by mocking _links.next responses

Response Structure Example (v1 API):

```json
{
  "results": [{
    "title": "Page Title",
    "excerpt": "...matching text...",
    "url": "/wiki/spaces/DEV/pages/123456/Page+Title",
    "lastModified": "2024-01-15T10:30:00.000Z",
    "content": {
      "id": "123456",
      "type": "page",
      "status": "current",
      "space": {
        "key": "DEV",
        "name": "Development"
      }
    }
  }],
  "start": 0,
  "limit": 25,
  "size": 1,
  "totalSize": 1,
  "cqlQuery": "type=page and space=DEV",
  "searchDuration": 45,
  "_links": {
    "next": "/rest/api/search?cql=type=page&limit=25&cursor=raNDoMsTRiNg"
  }
}
```

Dependency Updates Available (non-blocking):

- cobra v1.10.1 → v1.10.2 (minor update)
- goldmark v1.7.13 → v1.7.16 (minor update)
- Can be updated in separate PR after search implementation
