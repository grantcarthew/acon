package api

import (
	"bytes"
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

type PageListResponse struct {
	Results []Page `json:"results"`
}

type SpaceListResponse struct {
	Results []Space `json:"results"`
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

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
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
	ID      string         `json:"id"`
	SpaceID string         `json:"spaceId"`
	Status  string         `json:"status"`
	Title   string         `json:"title"`
	Body    *PageBodyWrite `json:"body"`
	Version *Version       `json:"version"`
}

func (c *Client) CreatePage(req *PageCreateRequest) (*Page, error) {
	respBody, err := c.doRequest("POST", "/wiki/api/v2/pages", req)
	if err != nil {
		return nil, fmt.Errorf("create page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create page response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetPage(pageID string) (*Page, error) {
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("pageID cannot be empty")
	}

	respBody, err := c.doRequest("GET", fmt.Sprintf("/wiki/api/v2/pages/%s?body-format=storage", pageID), nil)
	if err != nil {
		return nil, fmt.Errorf("get page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get page response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdatePage(pageID string, req *PageUpdateRequest) (*Page, error) {
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("pageID cannot be empty")
	}

	respBody, err := c.doRequest("PUT", fmt.Sprintf("/wiki/api/v2/pages/%s", pageID), req)
	if err != nil {
		return nil, fmt.Errorf("update page request failed: %w", err)
	}

	var result Page
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update page response: %w", err)
	}

	return &result, nil
}

func (c *Client) DeletePage(pageID string) error {
	if strings.TrimSpace(pageID) == "" {
		return fmt.Errorf("pageID cannot be empty")
	}

	_, err := c.doRequest("DELETE", fmt.Sprintf("/wiki/api/v2/pages/%s", pageID), nil)
	if err != nil {
		return fmt.Errorf("delete page request failed: %w", err)
	}
	return nil
}

func (c *Client) ListPages(spaceID string, limit int) ([]Page, error) {
	if strings.TrimSpace(spaceID) == "" {
		return nil, fmt.Errorf("spaceID cannot be empty")
	}

	path := fmt.Sprintf("/wiki/api/v2/pages?space-id=%s&limit=%d&body-format=storage", spaceID, limit)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list pages request failed: %w", err)
	}

	var result PageListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list pages response: %w", err)
	}

	return result.Results, nil
}

func (c *Client) GetSpace(spaceKey string) (*Space, error) {
	if strings.TrimSpace(spaceKey) == "" {
		return nil, fmt.Errorf("spaceKey cannot be empty")
	}

	respBody, err := c.doRequest("GET", fmt.Sprintf("/wiki/api/v2/spaces?keys=%s", spaceKey), nil)
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

func (c *Client) ListSpaces(limit int) ([]Space, error) {
	path := fmt.Sprintf("/wiki/api/v2/spaces?limit=%d", limit)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("list spaces request failed: %w", err)
	}

	var result SpaceListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list spaces response: %w", err)
	}

	return result.Results, nil
}
