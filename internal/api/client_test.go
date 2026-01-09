package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://example.atlassian.net", "test@example.com", "token123")

	if client.BaseURL != "https://example.atlassian.net" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "https://example.atlassian.net")
	}
	if client.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", client.Email, "test@example.com")
	}
	if client.APIToken != "token123" {
		t.Errorf("APIToken = %q, want %q", client.APIToken, "token123")
	}
	if client.client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestClient_GetPage(t *testing.T) {
	tests := []struct {
		name        string
		pageID      string
		statusCode  int
		response    any
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful get",
			pageID:     "123456",
			statusCode: http.StatusOK,
			response: Page{
				ID:      "123456",
				Title:   "Test Page",
				Status:  "current",
				SpaceID: "space-1",
			},
			wantErr: false,
		},
		{
			name:        "empty page ID",
			pageID:      "",
			wantErr:     true,
			errContains: "pageID cannot be empty",
		},
		{
			name:        "whitespace page ID",
			pageID:      "   ",
			wantErr:     true,
			errContains: "pageID cannot be empty",
		},
		{
			name:        "404 not found",
			pageID:      "999999",
			statusCode:  http.StatusNotFound,
			response:    map[string]string{"message": "Page not found"},
			wantErr:     true,
			errContains: "API error (status 404)",
		},
		{
			name:        "401 unauthorized",
			pageID:      "123456",
			statusCode:  http.StatusUnauthorized,
			response:    map[string]string{"message": "Unauthorized"},
			wantErr:     true,
			errContains: "API error (status 401)",
		},
		{
			name:        "500 server error",
			pageID:      "123456",
			statusCode:  http.StatusInternalServerError,
			response:    map[string]string{"message": "Internal error"},
			wantErr:     true,
			errContains: "API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodGet)
				}
				expectedPath := "/wiki/api/v2/pages/" + tt.pageID
				if !strings.HasPrefix(r.URL.Path, expectedPath) {
					t.Errorf("Path = %q, want prefix %q", r.URL.Path, expectedPath)
				}

				// Verify auth header is set
				if r.Header.Get("Authorization") == "" {
					t.Error("Authorization header not set")
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			result, err := client.GetPage(context.Background(), tt.pageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetPage() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if result.ID != "123456" {
					t.Errorf("GetPage() ID = %q, want %q", result.ID, "123456")
				}
				if result.Title != "Test Page" {
					t.Errorf("GetPage() Title = %q, want %q", result.Title, "Test Page")
				}
			}
		})
	}
}

func TestClient_CreatePage(t *testing.T) {
	tests := []struct {
		name        string
		request     *PageCreateRequest
		statusCode  int
		response    any
		wantErr     bool
		errContains string
	}{
		{
			name: "successful create",
			request: &PageCreateRequest{
				SpaceID: "space-1",
				Status:  "current",
				Title:   "New Page",
				Body: &PageBodyWrite{
					Representation: "storage",
					Value:          "<p>Content</p>",
				},
			},
			statusCode: http.StatusOK,
			response: Page{
				ID:      "789",
				Title:   "New Page",
				Status:  "current",
				SpaceID: "space-1",
			},
			wantErr: false,
		},
		{
			name: "create with parent",
			request: &PageCreateRequest{
				SpaceID:  "space-1",
				Status:   "current",
				Title:    "Child Page",
				ParentID: "parent-123",
				Body: &PageBodyWrite{
					Representation: "storage",
					Value:          "<p>Child content</p>",
				},
			},
			statusCode: http.StatusOK,
			response: Page{
				ID:       "790",
				Title:    "Child Page",
				ParentID: "parent-123",
			},
			wantErr: false,
		},
		{
			name: "400 bad request",
			request: &PageCreateRequest{
				SpaceID: "space-1",
				Title:   "Bad Page",
			},
			statusCode:  http.StatusBadRequest,
			response:    map[string]string{"message": "Invalid request"},
			wantErr:     true,
			errContains: "API error (status 400)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodPost)
				}
				if r.URL.Path != "/wiki/api/v2/pages" {
					t.Errorf("Path = %q, want %q", r.URL.Path, "/wiki/api/v2/pages")
				}

				// Verify request body
				var reqBody PageCreateRequest
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}
				if reqBody.Title != tt.request.Title {
					t.Errorf("Request Title = %q, want %q", reqBody.Title, tt.request.Title)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			result, err := client.CreatePage(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CreatePage() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr && result.Title != tt.request.Title {
				t.Errorf("CreatePage() Title = %q, want %q", result.Title, tt.request.Title)
			}
		})
	}
}

func TestClient_UpdatePage(t *testing.T) {
	tests := []struct {
		name        string
		pageID      string
		request     *PageUpdateRequest
		statusCode  int
		response    any
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful update",
			pageID: "123",
			request: &PageUpdateRequest{
				ID:      "123",
				SpaceID: "space-1",
				Status:  "current",
				Title:   "Updated Title",
				Body: &PageBodyWrite{
					Representation: "storage",
					Value:          "<p>Updated content</p>",
				},
				Version: &Version{Number: 2},
			},
			statusCode: http.StatusOK,
			response: Page{
				ID:      "123",
				Title:   "Updated Title",
				Version: &Version{Number: 2},
			},
			wantErr: false,
		},
		{
			name:        "empty page ID",
			pageID:      "",
			request:     &PageUpdateRequest{},
			wantErr:     true,
			errContains: "pageID cannot be empty",
		},
		{
			name:   "409 conflict",
			pageID: "123",
			request: &PageUpdateRequest{
				ID:      "123",
				Version: &Version{Number: 1},
			},
			statusCode:  http.StatusConflict,
			response:    map[string]string{"message": "Version conflict"},
			wantErr:     true,
			errContains: "API error (status 409)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodPut)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			result, err := client.UpdatePage(context.Background(), tt.pageID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("UpdatePage() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr && result.Title != "Updated Title" {
				t.Errorf("UpdatePage() Title = %q, want %q", result.Title, "Updated Title")
			}
		})
	}
}

func TestClient_DeletePage(t *testing.T) {
	tests := []struct {
		name        string
		pageID      string
		statusCode  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful delete",
			pageID:     "123",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:        "empty page ID",
			pageID:      "",
			wantErr:     true,
			errContains: "pageID cannot be empty",
		},
		{
			name:        "404 not found",
			pageID:      "999",
			statusCode:  http.StatusNotFound,
			wantErr:     true,
			errContains: "API error (status 404)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodDelete)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			err := client.DeletePage(context.Background(), tt.pageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("DeletePage() error = %q, want containing %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestClient_MovePage(t *testing.T) {
	tests := []struct {
		name        string
		pageID      string
		newParentID string
		setupServer func(t *testing.T) http.HandlerFunc
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful move",
			pageID:      "123",
			newParentID: "456",
			setupServer: func(t *testing.T) http.HandlerFunc {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.Header().Set("Content-Type", "application/json")

					switch {
					case callCount == 1 && strings.Contains(r.URL.Path, "/pages/123"):
						// Get source page
						json.NewEncoder(w).Encode(Page{
							ID:      "123",
							Title:   "Source Page",
							SpaceID: "space-1",
							Body:    &PageBodyGet{Storage: &BodyContent{Value: "<p>Content</p>"}},
							Version: &Version{Number: 1},
						})
					case callCount == 2 && strings.Contains(r.URL.Path, "/pages/456"):
						// Get target page
						json.NewEncoder(w).Encode(Page{
							ID:      "456",
							Title:   "Target Page",
							SpaceID: "space-1",
						})
					case callCount == 3 && r.Method == http.MethodPut:
						// Update page
						json.NewEncoder(w).Encode(Page{
							ID:       "123",
							Title:    "Source Page",
							ParentID: "456",
						})
					default:
						t.Errorf("Unexpected request: %s %s (call %d)", r.Method, r.URL.Path, callCount)
						w.WriteHeader(http.StatusBadRequest)
					}
				}
			},
			wantErr: false,
		},
		{
			name:        "empty page ID",
			pageID:      "",
			newParentID: "456",
			wantErr:     true,
			errContains: "pageID cannot be empty",
		},
		{
			name:        "empty parent ID",
			pageID:      "123",
			newParentID: "",
			wantErr:     true,
			errContains: "newParentID cannot be empty",
		},
		{
			name:        "cross-space move",
			pageID:      "123",
			newParentID: "456",
			setupServer: func(t *testing.T) http.HandlerFunc {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.Header().Set("Content-Type", "application/json")

					switch callCount {
					case 1:
						json.NewEncoder(w).Encode(Page{ID: "123", SpaceID: "space-1"})
					case 2:
						json.NewEncoder(w).Encode(Page{ID: "456", SpaceID: "space-2"})
					}
				}
			},
			wantErr:     true,
			errContains: "cross-space moves are not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(tt.setupServer(t))
				defer server.Close()
			}

			baseURL := "http://localhost"
			if server != nil {
				baseURL = server.URL
			}

			client := NewClient(baseURL, "test@example.com", "token")
			result, err := client.MovePage(context.Background(), tt.pageID, tt.newParentID)

			if (err != nil) != tt.wantErr {
				t.Errorf("MovePage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("MovePage() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr && result.ParentID != "456" {
				t.Errorf("MovePage() ParentID = %q, want %q", result.ParentID, "456")
			}
		})
	}
}

func TestClient_ListPages(t *testing.T) {
	tests := []struct {
		name        string
		spaceID     string
		limit       int
		sort        string
		setupServer func(t *testing.T) http.HandlerFunc
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful list",
			spaceID: "space-1",
			limit:   10,
			sort:    "",
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(PageListResponse{
						Results: []Page{
							{ID: "1", Title: "Page 1"},
							{ID: "2", Title: "Page 2"},
							{ID: "3", Title: "Page 3"},
						},
					})
				}
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "with sort parameter",
			spaceID: "space-1",
			limit:   10,
			sort:    "-created-date",
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "sort=-created-date") {
						t.Errorf("Sort parameter not found in query: %s", r.URL.RawQuery)
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(PageListResponse{
						Results: []Page{{ID: "1", Title: "Page 1"}},
					})
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "empty space ID",
			spaceID:     "",
			limit:       10,
			wantErr:     true,
			errContains: "spaceID cannot be empty",
		},
		{
			name:    "pagination",
			spaceID: "space-1",
			limit:   50,
			setupServer: func(t *testing.T) http.HandlerFunc {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.Header().Set("Content-Type", "application/json")

					if callCount == 1 {
						// First page
						pages := make([]Page, 25)
						for i := range 25 {
							pages[i] = Page{ID: string(rune('a' + i)), Title: "Page"}
						}
						json.NewEncoder(w).Encode(PageListResponse{
							Results: pages,
							Links:   PaginationLinks{Next: "/wiki/api/v2/pages?cursor=abc"},
						})
					} else {
						// Second page
						pages := make([]Page, 25)
						for i := range 25 {
							pages[i] = Page{ID: string(rune('A' + i)), Title: "Page"}
						}
						json.NewEncoder(w).Encode(PageListResponse{
							Results: pages,
						})
					}
				}
			},
			wantCount: 50,
			wantErr:   false,
		},
		{
			name:    "limit trims results",
			spaceID: "space-1",
			limit:   2,
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(PageListResponse{
						Results: []Page{
							{ID: "1", Title: "Page 1"},
							{ID: "2", Title: "Page 2"},
							{ID: "3", Title: "Page 3"},
						},
					})
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(tt.setupServer(t))
				defer server.Close()
			}

			baseURL := "http://localhost"
			if server != nil {
				baseURL = server.URL
			}

			client := NewClient(baseURL, "test@example.com", "token")
			result, err := client.ListPages(context.Background(), tt.spaceID, tt.limit, tt.sort)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ListPages() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("ListPages() returned %d pages, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestClient_GetChildPages(t *testing.T) {
	tests := []struct {
		name        string
		parentID    string
		limit       int
		sort        string
		setupServer func(t *testing.T) http.HandlerFunc
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful get children",
			parentID: "parent-1",
			limit:    10,
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.Path, "/children") {
						t.Errorf("Expected /children in path: %s", r.URL.Path)
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(PageListResponse{
						Results: []Page{
							{ID: "c1", Title: "Child 1"},
							{ID: "c2", Title: "Child 2"},
						},
					})
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:        "empty parent ID",
			parentID:    "",
			limit:       10,
			wantErr:     true,
			errContains: "parentID cannot be empty",
		},
		{
			name:     "with sort",
			parentID: "parent-1",
			limit:    10,
			sort:     "child-position",
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "sort=child-position") {
						t.Errorf("Sort parameter not found: %s", r.URL.RawQuery)
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(PageListResponse{
						Results: []Page{{ID: "c1", Title: "Child 1"}},
					})
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(tt.setupServer(t))
				defer server.Close()
			}

			baseURL := "http://localhost"
			if server != nil {
				baseURL = server.URL
			}

			client := NewClient(baseURL, "test@example.com", "token")
			result, err := client.GetChildPages(context.Background(), tt.parentID, tt.limit, tt.sort)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetChildPages() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("GetChildPages() returned %d pages, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestClient_GetSpace(t *testing.T) {
	tests := []struct {
		name        string
		spaceKey    string
		statusCode  int
		response    any
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful get",
			spaceKey:   "TEST",
			statusCode: http.StatusOK,
			response: SpaceListResponse{
				Results: []Space{
					{ID: "space-1", Key: "TEST", Name: "Test Space", Type: "global"},
				},
			},
			wantErr: false,
		},
		{
			name:        "empty space key",
			spaceKey:    "",
			wantErr:     true,
			errContains: "spaceKey cannot be empty",
		},
		{
			name:       "space not found",
			spaceKey:   "NOTFOUND",
			statusCode: http.StatusOK,
			response: SpaceListResponse{
				Results: []Space{},
			},
			wantErr:     true,
			errContains: "space not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.RawQuery, "keys="+tt.spaceKey) {
					t.Errorf("Expected keys=%s in query: %s", tt.spaceKey, r.URL.RawQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			result, err := client.GetSpace(context.Background(), tt.spaceKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetSpace() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if result.Key != "TEST" {
					t.Errorf("GetSpace() Key = %q, want %q", result.Key, "TEST")
				}
			}
		})
	}
}

func TestClient_ListSpaces(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		setupServer func(t *testing.T) http.HandlerFunc
		wantCount   int
		wantErr     bool
	}{
		{
			name:  "successful list",
			limit: 10,
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(SpaceListResponse{
						Results: []Space{
							{ID: "1", Key: "SPACE1", Name: "Space 1"},
							{ID: "2", Key: "SPACE2", Name: "Space 2"},
						},
					})
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "pagination",
			limit: 50,
			setupServer: func(t *testing.T) http.HandlerFunc {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.Header().Set("Content-Type", "application/json")

					if callCount == 1 {
						spaces := make([]Space, 25)
						for i := range 25 {
							spaces[i] = Space{ID: string(rune('a' + i)), Key: "KEY"}
						}
						json.NewEncoder(w).Encode(SpaceListResponse{
							Results: spaces,
							Links:   PaginationLinks{Next: "/wiki/api/v2/spaces?cursor=abc"},
						})
					} else {
						spaces := make([]Space, 25)
						for i := range 25 {
							spaces[i] = Space{ID: string(rune('A' + i)), Key: "KEY"}
						}
						json.NewEncoder(w).Encode(SpaceListResponse{
							Results: spaces,
						})
					}
				}
			},
			wantCount: 50,
			wantErr:   false,
		},
		{
			name:  "limit trims results",
			limit: 1,
			setupServer: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(SpaceListResponse{
						Results: []Space{
							{ID: "1", Key: "SPACE1"},
							{ID: "2", Key: "SPACE2"},
						},
					})
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.setupServer(t))
			defer server.Close()

			client := NewClient(server.URL, "test@example.com", "token")
			result, err := client.ListSpaces(context.Background(), tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSpaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("ListSpaces() returned %d spaces, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestClient_doRequest_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Content-Type header
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		// Verify Accept header
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}

		// Verify Basic Auth
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("Basic auth not set")
		}
		if user != "test@example.com" {
			t.Errorf("Auth user = %q, want %q", user, "test@example.com")
		}
		if pass != "secret-token" {
			t.Errorf("Auth pass = %q, want %q", pass, "secret-token")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page{ID: "1"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "secret-token")
	_, err := client.GetPage(context.Background(), "1")
	if err != nil {
		t.Errorf("GetPage() error = %v", err)
	}
}

func TestClient_doRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "token")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetPage(ctx, "123")
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}
