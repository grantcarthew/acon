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
	BaseURL  string
	Email    string
	APIToken string
	client   *http.Client
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

func NewClient(baseURL, email, apiToken string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := strings.TrimRight(c.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
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

func (c *Client) ListPages(ctx context.Context, spaceID string, limit int, sort string) ([]Page, error) {
	if strings.TrimSpace(spaceID) == "" {
		return nil, fmt.Errorf("spaceID cannot be empty")
	}

	var allPages []Page
	perPage := min(limit, maxPerPage)

	path := fmt.Sprintf("/wiki/api/v2/pages?space-id=%s&limit=%d&body-format=storage", spaceID, perPage)
	if strings.TrimSpace(sort) != "" {
		path += fmt.Sprintf("&sort=%s", sort)
	}

	for {
		respBody, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("list pages request failed: %w", err)
		}

		var result PageListResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse list pages response: %w", err)
		}

		allPages = append(allPages, result.Results...)

		// Stop if we have enough or no more pages
		if len(allPages) >= limit || result.Links.Next == "" {
			break
		}

		// Use the next link for subsequent requests
		path = result.Links.Next
	}

	// Trim to exact limit
	if len(allPages) > limit {
		allPages = allPages[:limit]
	}

	return allPages, nil
}

func (c *Client) GetChildPages(ctx context.Context, parentID string, limit int, sort string) ([]Page, error) {
	if strings.TrimSpace(parentID) == "" {
		return nil, fmt.Errorf("parentID cannot be empty")
	}

	var allPages []Page
	perPage := min(limit, maxPerPage)

	path := fmt.Sprintf("/wiki/api/v2/pages/%s/children?limit=%d", parentID, perPage)
	if strings.TrimSpace(sort) != "" {
		path += fmt.Sprintf("&sort=%s", sort)
	}

	for {
		respBody, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("get child pages request failed: %w", err)
		}

		var result PageListResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse child pages response: %w", err)
		}

		allPages = append(allPages, result.Results...)

		// Stop if we have enough or no more pages
		if len(allPages) >= limit || result.Links.Next == "" {
			break
		}

		// Use the next link for subsequent requests
		path = result.Links.Next
	}

	// Trim to exact limit
	if len(allPages) > limit {
		allPages = allPages[:limit]
	}

	return allPages, nil
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
