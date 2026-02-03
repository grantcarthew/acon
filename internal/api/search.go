package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// DefaultSearchLimit is the default maximum number of search results per request
const DefaultSearchLimit = 25

// spaceKeyRegex validates space keys (compiled once at package level for performance)
var spaceKeyRegex = regexp.MustCompile(`^~?[A-Za-z0-9_]+$`)

// SearchSpace represents space information in search results
type SearchSpace struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// SearchContent represents the nested content object in search results
type SearchContent struct {
	ID     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Space  SearchSpace `json:"space"`
}

// SearchResult represents a single search result from the v1 API
type SearchResult struct {
	Title        string        `json:"title"`
	Excerpt      string        `json:"excerpt"`
	URL          string        `json:"url"`
	LastModified string        `json:"lastModified"`
	Content      SearchContent `json:"content"`
}

// SearchPaginationLinks represents the _links field in search response
type SearchPaginationLinks struct {
	Next string `json:"next,omitempty"`
}

// SearchResponse represents the v1 search API response
type SearchResponse struct {
	Results        []SearchResult        `json:"results"`
	Start          int                   `json:"start"`
	Limit          int                   `json:"limit"`
	Size           int                   `json:"size"`
	TotalSize      int                   `json:"totalSize"`
	CQLQuery       string                `json:"cqlQuery"`
	SearchDuration int                   `json:"searchDuration"`
	Links          SearchPaginationLinks `json:"_links,omitempty"`
}

// SearchParams holds the parameters for building a CQL query
type SearchParams struct {
	Text    string
	Title   string
	Label   string
	Creator string
	Space   string
	Type    string
}

// escapeCQLString escapes special characters in CQL string values.
// CQL uses Lucene syntax which requires escaping these special characters with backslash:
// + - & | ! ( ) { } [ ] ^ " ~ * ? : \ /
// Reference: https://developer.atlassian.com/server/confluence/performing-text-searches-using-cql
//
// This function processes the input string in a single pass, checking each character
// and prefixing special characters with a backslash. The order of characters in the
// switch statement does not affect the result since each character is processed independently.
func escapeCQLString(s string) string {
	// Use strings.Builder for efficient single-pass string construction
	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate to avoid reallocations

	for _, ch := range s {
		// Check if character needs escaping and add backslash prefix
		switch ch {
		case '\\', '+', '-', '&', '|', '!', '(', ')', '{', '}', '[', ']', '^', '"', '~', '*', '?', ':', '/':
			result.WriteRune('\\')
		}
		result.WriteRune(ch)
	}

	return result.String()
}

// validateSpaceKey checks if a space key has a valid format.
// Space keys must be 1-255 characters, alphanumeric with underscores.
func validateSpaceKey(key string) error {
	if key == "" {
		return nil // Empty is allowed (means no space filter)
	}
	if len(key) > 255 {
		return fmt.Errorf("space key too long (max 255 characters)")
	}
	// Space keys can be alphanumeric (upper or lowercase) with underscores
	// Allow tilde prefix for personal spaces (~username)
	if !spaceKeyRegex.MatchString(key) {
		return fmt.Errorf("invalid space key format (must be alphanumeric with underscores, optionally prefixed with ~)")
	}
	return nil
}

// validContentTypes defines the allowlist of valid Confluence content types.
// These are CQL keywords and should not be quoted or escaped.
var validContentTypes = map[string]bool{
	"page":       true,
	"blogpost":   true,
	"attachment": true,
	"comment":    true,
	"whiteboard": true,
	"database":   true,
	"embed":      true,
	"folder":     true,
}

// validateContentType checks if a content type is valid.
// Content types are CQL keywords and must match the allowlist.
func validateContentType(contentType string) error {
	if contentType == "" {
		return nil // Empty is allowed (will default to "page")
	}
	if !validContentTypes[contentType] {
		// Generate valid types list dynamically from the map
		validTypes := make([]string, 0, len(validContentTypes))
		for t := range validContentTypes {
			validTypes = append(validTypes, t)
		}
		sort.Strings(validTypes) // Sort for consistent error messages
		return fmt.Errorf("invalid content type: %s (valid types: %s)", contentType, strings.Join(validTypes, ", "))
	}
	return nil
}

// BuildCQL constructs a CQL query from search parameters.
// All conditions are combined with AND logic.
// Returns the CQL string (not URL-encoded) or an error if validation fails.
func BuildCQL(params SearchParams) (string, error) {
	var conditions []string

	// Validate space key if provided
	if err := validateSpaceKey(params.Space); err != nil {
		return "", fmt.Errorf("invalid space key: %w", err)
	}

	// Default to type=page unless specified
	contentType := params.Type
	if contentType == "" {
		contentType = "page"
	}

	// Validate content type
	if err := validateContentType(contentType); err != nil {
		return "", err
	}

	// SECURITY NOTE: Content type is a CQL keyword (not user string data) and MUST NOT be quoted/escaped.
	// This is ONLY SAFE because validateContentType() enforces an allowlist above.
	// The allowlist prevents CQL injection by rejecting any value not in validContentTypes map.
	// DO NOT modify this code to bypass validation - doing so creates a critical injection vulnerability.
	//
	// Defense-in-depth: Assert that contentType is in the allowlist before using it unescaped
	if !validContentTypes[contentType] {
		// This should never happen if validation above worked correctly
		return "", fmt.Errorf("internal error: contentType '%s' not in allowlist (validation was bypassed)", contentType)
	}
	conditions = append(conditions, fmt.Sprintf("type=%s", contentType))

	// Text search (full-text: searches title, body, labels)
	if params.Text != "" {
		conditions = append(conditions, fmt.Sprintf("text ~ \"%s\"", escapeCQLString(params.Text)))
	}

	// Title-only search
	if params.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title ~ \"%s\"", escapeCQLString(params.Title)))
	}

	// Label search (exact match)
	if params.Label != "" {
		conditions = append(conditions, fmt.Sprintf("label = \"%s\"", escapeCQLString(params.Label)))
	}

	// Creator search (handle 'me' alias - case insensitive for better UX)
	if params.Creator != "" {
		if strings.EqualFold(params.Creator, "me") {
			conditions = append(conditions, "creator = currentUser()")
		} else {
			conditions = append(conditions, fmt.Sprintf("creator = \"%s\"", escapeCQLString(params.Creator)))
		}
	}

	// Space filter (space keys must be quoted in CQL syntax)
	// Reference: https://developer.atlassian.com/server/confluence/advanced-searching-using-cql
	// Example: space = "TEST" or space = "~username" for personal spaces
	//
	// SECURITY NOTE: Space keys are NOT escaped because validateSpaceKey() above enforces
	// a strict regex pattern (^~?[A-Za-z0-9_]+$) that only allows safe characters.
	// The tilde prefix for personal spaces (~username) is part of the CQL identifier syntax,
	// not a Lucene special character that needs escaping in this context.
	// DO NOT add escaping here - it will break personal space searches.
	if params.Space != "" {
		// Defense-in-depth: Assert that space key matches expected format before using it unescaped
		if !spaceKeyRegex.MatchString(params.Space) {
			// This should never happen if validation above worked correctly
			return "", fmt.Errorf("internal error: space key '%s' failed regex validation (validation was bypassed)", params.Space)
		}
		conditions = append(conditions, fmt.Sprintf("space = \"%s\"", params.Space))
	}

	return strings.Join(conditions, " and "), nil
}

// Search performs a CQL search using the v1 API.
// The cql parameter should be the complete CQL query string (not URL-encoded).
// The limit parameter controls the maximum number of results per page.
// The start parameter specifies the starting index for pagination (0-based).
// Returns the search response, a boolean indicating if more results are available, and an error.
func (c *Client) Search(ctx context.Context, cql string, limit, start int) (*SearchResponse, bool, error) {
	if strings.TrimSpace(cql) == "" {
		return nil, false, fmt.Errorf("cql query cannot be empty")
	}

	if limit <= 0 {
		return nil, false, fmt.Errorf("limit must be greater than 0")
	}

	if start < 0 {
		return nil, false, fmt.Errorf("start must be greater than or equal to 0")
	}

	// Construct query parameters using url.Values for safe encoding
	params := url.Values{}
	params.Set("cql", cql)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("excerpt", "indexed")

	path := "/wiki/rest/api/search?" + params.Encode()

	respBody, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, false, fmt.Errorf("search request failed: %w", err)
	}

	var result SearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Validate response values before arithmetic to prevent integer overflow
	// and handle malicious/buggy API responses gracefully
	if result.Start < 0 || result.Size < 0 || result.TotalSize < 0 {
		return nil, false, fmt.Errorf("invalid search response: negative values (start=%d, size=%d, totalSize=%d)",
			result.Start, result.Size, result.TotalSize)
	}

	// Sanity check: current position shouldn't exceed total
	if result.Start > result.TotalSize {
		return nil, false, fmt.Errorf("invalid search response: start position %d exceeds total size %d",
			result.Start, result.TotalSize)
	}

	// Calculate if more results are available
	// Safe to do arithmetic now that we've validated non-negative values
	// Check for overflow before addition (Go silently wraps on overflow)
	if result.Start > 0 && result.Size > 0 && result.Start > (1<<31-result.Size) {
		// Extremely unlikely in practice, but handle gracefully
		return nil, false, fmt.Errorf("invalid search response: pagination overflow (start=%d, size=%d)", result.Start, result.Size)
	}
	hasMore := result.Start+result.Size < result.TotalSize

	return &result, hasMore, nil
}
