package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Email      string
	APIToken   string
	client     *http.Client
	VerboseLog io.Writer // Writer for verbose logging (typically os.Stderr or nil)
}

type Page struct {
	ID       string       `json:"id,omitempty"`
	SpaceID  string       `json:"spaceId,omitempty"`
	Status   string       `json:"status,omitempty"`
	Title    string       `json:"title"`
	Body     *PageBodyGet `json:"body,omitempty"`
	ParentID string       `json:"parentId,omitempty"`
	Version  *Version     `json:"version,omitempty"`
}

type PageBodyGet struct {
	Storage        *BodyContent `json:"storage,omitempty"`
	AtlasDocFormat *BodyContent `json:"atlas_doc_format,omitempty"`
}

type BodyContent struct {
	Representation string `json:"representation,omitempty"`
	Value          string `json:"value,omitempty"`
}

type PageBodyWrite struct {
	Representation string `json:"representation"`
	Value          string `json:"value"`
}

type Version struct {
	Number  int    `json:"number"`
	Message string `json:"message,omitempty"`
}

type Space struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PaginationLinks represents the _links field in paginated API responses
type PaginationLinks struct {
	Next string `json:"next,omitempty"`
}

type PageListResponse struct {
	Results []Page          `json:"results"`
	Links   PaginationLinks `json:"_links,omitempty"`
}

type SpaceListResponse struct {
	Results []Space         `json:"results"`
	Links   PaginationLinks `json:"_links,omitempty"`
}

func NewClient(baseURL, email, apiToken string) (*Client, error) {
	// Validate required parameters to fail fast
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if strings.TrimSpace(apiToken) == "" {
		return nil, fmt.Errorf("apiToken cannot be empty")
	}

	return &Client{
		BaseURL:    baseURL,
		Email:      email,
		APIToken:   apiToken,
		VerboseLog: nil, // Set by caller if verbose mode enabled
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// logVerbose writes to VerboseLog if it's set
func (c *Client) logVerbose(format string, args ...interface{}) {
	if c.VerboseLog != nil {
		fmt.Fprintf(c.VerboseLog, format, args...)
	}
}

// truncateStringUTF8Safe safely truncates a string to maxRunes runes,
// ensuring we don't split multi-byte UTF-8 characters.
func truncateStringUTF8Safe(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Only track timing if verbose logging is enabled
	var start time.Time
	if c.VerboseLog != nil {
		start = time.Now()
	}

	var reqBody io.Reader
	var reqBodyBytes []byte
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBodyBytes = jsonData
		reqBody = bytes.NewBuffer(reqBodyBytes)
	}

	url := strings.TrimRight(c.BaseURL, "/") + path
	c.logVerbose("[API] %s %s\n", method, url)

	if c.VerboseLog != nil && len(reqBodyBytes) > 0 {
		// Truncate large bodies (UTF-8 safe to avoid splitting multi-byte characters)
		preview := truncateStringUTF8Safe(string(reqBodyBytes), 200)
		c.logVerbose("[API] Request body: %s\n", preview)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logVerbose("[API] Request failed: %v\n", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.VerboseLog != nil {
		duration := time.Since(start)
		c.logVerbose("[API] Response status: %d (took %v)\n", resp.StatusCode, duration)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logVerbose("[API] Error response: %s\n", string(respBody))
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if c.VerboseLog != nil {
		// Log response preview for successful requests (UTF-8 safe truncation)
		preview := truncateStringUTF8Safe(string(respBody), 200)
		c.logVerbose("[API] Response body: %s\n", preview)
	}

	return respBody, nil
}

type PageCreateRequest struct {
	SpaceID  string         `json:"spaceId"`
	Status   string         `json:"status"`
	Title    string         `json:"title"`
	Body     *PageBodyWrite `json:"body"`
	ParentID string         `json:"parentId,omitempty"`
}

type PageUpdateRequest struct {
	ID       string         `json:"id"`
	SpaceID  string         `json:"spaceId"`
	Status   string         `json:"status"`
	Title    string         `json:"title"`
	ParentID string         `json:"parentId,omitempty"`
	Body     *PageBodyWrite `json:"body"`
	Version  *Version       `json:"version"`
}

func (c *Client) CreatePage(ctx context.Context, req *PageCreateRequest) (*Page, error) {
	respBody, err := c.doRequest(ctx, "POST", "/wiki/api/v2/pages", req)
	if err != nil {
		return nil, fmt.Errorf("create page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create page response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("pageID cannot be empty")
	}

	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/wiki/api/v2/pages/%s?body-format=storage", pageID), nil)
	if err != nil {
		return nil, fmt.Errorf("get page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get page response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdatePage(ctx context.Context, pageID string, req *PageUpdateRequest) (*Page, error) {
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("pageID cannot be empty")
	}

	respBody, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/wiki/api/v2/pages/%s", pageID), req)
	if err != nil {
		return nil, fmt.Errorf("update page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update page response: %w", err)
	}

	return &result, nil
}

func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	if strings.TrimSpace(pageID) == "" {
		return fmt.Errorf("pageID cannot be empty")
	}

	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/wiki/api/v2/pages/%s", pageID), nil)
	if err != nil {
		return fmt.Errorf("delete page request failed: %w", err)
	}
	return nil
}

func (c *Client) MovePage(ctx context.Context, pageID, newParentID string) (*Page, error) {
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("pageID cannot be empty")
	}
	if strings.TrimSpace(newParentID) == "" {
		return nil, fmt.Errorf("newParentID cannot be empty")
	}

	// Fetch source page
	sourcePage, err := c.GetPage(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source page: %w", err)
	}

	// Fetch target parent page
	targetPage, err := c.GetPage(ctx, newParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target parent page: %w", err)
	}

	// Check for cross-space move
	if sourcePage.SpaceID != targetPage.SpaceID {
		return nil, fmt.Errorf("cross-space moves are not supported; use create and delete instead")
	}

	// Get body content
	bodyValue := ""
	if sourcePage.Body != nil && sourcePage.Body.Storage != nil {
		bodyValue = sourcePage.Body.Storage.Value
	}

	// Build update request
	newVersion := 1
	if sourcePage.Version != nil {
		newVersion = sourcePage.Version.Number + 1
	}

	req := &PageUpdateRequest{
		ID:       pageID,
		SpaceID:  sourcePage.SpaceID,
		Status:   "current",
		Title:    sourcePage.Title,
		ParentID: newParentID,
		Body: &PageBodyWrite{
			Representation: "storage",
			Value:          bodyValue,
		},
		Version: &Version{
			Number:  newVersion,
			Message: fmt.Sprintf("Moved to parent %s", newParentID),
		},
	}

	return c.UpdatePage(ctx, pageID, req)
}

const maxPerPage = 25 // Confluence API v2 max per request
const maxLimit = 1000 // Protect against memory exhaustion and excessive API calls (40 max requests)

// paginatePages handles common pagination logic for page list operations.
// It validates the limit, fetches pages across multiple API requests if needed,
// trims results to the exact limit, and returns whether more pages are available.
func (c *Client) paginatePages(ctx context.Context, initialPath string, limit int, errorContext string) ([]Page, bool, error) {
	if limit <= 0 {
		return nil, false, fmt.Errorf("limit must be greater than 0")
	}
	if limit > maxLimit {
		return nil, false, fmt.Errorf("limit cannot exceed %d", maxLimit)
	}

	c.logVerbose("[Pagination] Starting pagination: limit=%d\n", limit)

	var allPages []Page
	hasMore := false
	path := initialPath
	requestNum := 0

	for {
		requestNum++
		c.logVerbose("[Pagination] Request %d: fetching from API\n", requestNum)

		respBody, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, false, fmt.Errorf("%s request failed: %w", errorContext, err)
		}

		var result PageListResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, false, fmt.Errorf("failed to parse %s response: %w", errorContext, err)
		}

		c.logVerbose("[Pagination] Received %d pages (total so far: %d)\n", len(result.Results), len(allPages)+len(result.Results))
		allPages = append(allPages, result.Results...)

		// Check if there are more pages available from the API
		hasMore = result.Links.Next != ""

		// Stop if we have enough or no more pages
		if len(allPages) >= limit || !hasMore {
			break
		}

		// Use the API-provided next link for subsequent requests
		path = result.Links.Next
	}

	// Trim to exact limit if we accumulated more than requested
	trimmed := len(allPages) > limit
	if trimmed {
		c.logVerbose("[Pagination] Trimming results from %d to %d\n", len(allPages), limit)
		allPages = allPages[:limit]
	}

	// hasMore is true if either the API has more pages OR we trimmed local results
	hasMore = hasMore || trimmed
	c.logVerbose("[Pagination] Complete: returning %d pages, hasMore=%v\n", len(allPages), hasMore)

	return allPages, hasMore, nil
}

func (c *Client) ListPages(ctx context.Context, spaceID string, limit int, sort string) ([]Page, bool, error) {
	if strings.TrimSpace(spaceID) == "" {
		return nil, false, fmt.Errorf("spaceID cannot be empty")
	}

	path := fmt.Sprintf("/wiki/api/v2/pages?space-id=%s&limit=%d&body-format=storage", spaceID, min(limit, maxPerPage))
	if strings.TrimSpace(sort) != "" {
		path += fmt.Sprintf("&sort=%s", sort)
	}

	return c.paginatePages(ctx, path, limit, "list pages")
}

func (c *Client) GetChildPages(ctx context.Context, parentID string, limit int, sort string) ([]Page, bool, error) {
	if strings.TrimSpace(parentID) == "" {
		return nil, false, fmt.Errorf("parentID cannot be empty")
	}

	path := fmt.Sprintf("/wiki/api/v2/pages/%s/children?limit=%d", parentID, min(limit, maxPerPage))
	if strings.TrimSpace(sort) != "" {
		path += fmt.Sprintf("&sort=%s", sort)
	}

	return c.paginatePages(ctx, path, limit, "get child pages")
}

func (c *Client) GetSpace(ctx context.Context, spaceKey string) (*Space, error) {
	if strings.TrimSpace(spaceKey) == "" {
		return nil, fmt.Errorf("spaceKey cannot be empty")
	}

	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/wiki/api/v2/spaces?keys=%s", spaceKey), nil)
	if err != nil {
		return nil, fmt.Errorf("get space request failed: %w", err)
	}

	var result SpaceListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get space response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("space not found: %s", spaceKey)
	}

	return &result.Results[0], nil
}

func (c *Client) ListSpaces(ctx context.Context, limit int) ([]Space, error) {
	var allSpaces []Space
	perPage := min(limit, maxPerPage)

	path := fmt.Sprintf("/wiki/api/v2/spaces?limit=%d", perPage)

	for {
		respBody, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("list spaces request failed: %w", err)
		}

		var result SpaceListResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse list spaces response: %w", err)
		}

		allSpaces = append(allSpaces, result.Results...)

		// Stop if we have enough or no more pages
		if len(allSpaces) >= limit || result.Links.Next == "" {
			break
		}

		// Use the next link for subsequent requests
		path = result.Links.Next
	}

	// Trim to exact limit
	if len(allSpaces) > limit {
		allSpaces = allSpaces[:limit]
	}

	return allSpaces, nil
}
