package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildCQL(t *testing.T) {
	tests := []struct {
		name        string
		params      SearchParams
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "default type only",
			params:  SearchParams{},
			want:    "type=page",
			wantErr: false,
		},
		{
			name:    "text search",
			params:  SearchParams{Text: "api docs"},
			want:    "type=page and text ~ \"api docs\"",
			wantErr: false,
		},
		{
			name:    "title search",
			params:  SearchParams{Title: "Security Review"},
			want:    "type=page and title ~ \"Security Review\"",
			wantErr: false,
		},
		{
			name:    "label search",
			params:  SearchParams{Label: "urgent"},
			want:    "type=page and label = \"urgent\"",
			wantErr: false,
		},
		{
			name:    "creator with email",
			params:  SearchParams{Creator: "user@example.com"},
			want:    "type=page and creator = \"user@example.com\"",
			wantErr: false,
		},
		{
			name:    "creator with me alias",
			params:  SearchParams{Creator: "me"},
			want:    "type=page and creator = currentUser()",
			wantErr: false,
		},
		{
			name:    "creator with me alias - uppercase",
			params:  SearchParams{Creator: "ME"},
			want:    "type=page and creator = currentUser()",
			wantErr: false,
		},
		{
			name:    "creator with me alias - mixed case",
			params:  SearchParams{Creator: "Me"},
			want:    "type=page and creator = currentUser()",
			wantErr: false,
		},
		{
			name:    "space filter",
			params:  SearchParams{Space: "DEV"},
			want:    "type=page and space = \"DEV\"",
			wantErr: false,
		},
		{
			name:    "custom type",
			params:  SearchParams{Type: "blogpost"},
			want:    "type=blogpost",
			wantErr: false,
		},
		{
			name: "multiple conditions",
			params: SearchParams{
				Text:    "api",
				Label:   "important",
				Creator: "me",
				Space:   "DEV",
			},
			want:    "type=page and text ~ \"api\" and label = \"important\" and creator = currentUser() and space = \"DEV\"",
			wantErr: false,
		},
		{
			name: "text and title combination",
			params: SearchParams{
				Text:  "critical",
				Title: "Security Review",
			},
			want:    "type=page and text ~ \"critical\" and title ~ \"Security Review\"",
			wantErr: false,
		},
		{
			name: "all parameters",
			params: SearchParams{
				Text:    "bug",
				Title:   "Critical",
				Label:   "urgent",
				Creator: "user@example.com",
				Space:   "TEAM",
				Type:    "page",
			},
			want:    "type=page and text ~ \"bug\" and title ~ \"Critical\" and label = \"urgent\" and creator = \"user@example.com\" and space = \"TEAM\"",
			wantErr: false,
		},
		{
			name: "attachment type with space",
			params: SearchParams{
				Text:  "diagram",
				Space: "TEAM",
				Type:  "attachment",
			},
			want:    "type=attachment and text ~ \"diagram\" and space = \"TEAM\"",
			wantErr: false,
		},
		{
			name:    "label only search",
			params:  SearchParams{Label: "todo"},
			want:    "type=page and label = \"todo\"",
			wantErr: false,
		},
		{
			name:    "text with quotes - escaping",
			params:  SearchParams{Text: "api \"test\""},
			want:    "type=page and text ~ \"api \\\"test\\\"\"",
			wantErr: false,
		},
		{
			name:    "title with quotes - escaping",
			params:  SearchParams{Title: "Review \"Critical\""},
			want:    "type=page and title ~ \"Review \\\"Critical\\\"\"",
			wantErr: false,
		},
		{
			name:    "label with quotes - escaping",
			params:  SearchParams{Label: "urgent\"important"},
			want:    "type=page and label = \"urgent\\\"important\"",
			wantErr: false,
		},
		{
			name:    "text with backslashes - escaping",
			params:  SearchParams{Text: "path\\to\\file"},
			want:    "type=page and text ~ \"path\\\\to\\\\file\"",
			wantErr: false,
		},
		{
			name:    "text with both backslashes and quotes",
			params:  SearchParams{Text: "\\\"escaped\\\""},
			want:    "type=page and text ~ \"\\\\\\\"escaped\\\\\\\"\"",
			wantErr: false,
		},
		{
			name:    "personal space with tilde",
			params:  SearchParams{Space: "~USERNAME"},
			want:    "type=page and space = \"~USERNAME\"",
			wantErr: false,
		},
		{
			name:    "valid space key - lowercase",
			params:  SearchParams{Space: "dev"},
			want:    "type=page and space = \"dev\"",
			wantErr: false,
		},
		{
			name:        "invalid space key - special chars",
			params:      SearchParams{Space: "DEV-TEAM"},
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "invalid space key - too long",
			params:      SearchParams{Space: strings.Repeat("A", 256)},
			wantErr:     true,
			errContains: "space key too long",
		},
		{
			name:    "space key with underscore",
			params:  SearchParams{Space: "DEV_TEAM"},
			want:    "type=page and space = \"DEV_TEAM\"",
			wantErr: false,
		},
		{
			name:        "invalid content type",
			params:      SearchParams{Type: "unknown"},
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid content type - uppercase",
			params:      SearchParams{Type: "PAGE"},
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid content type - injection attempt",
			params:      SearchParams{Type: "page;drop table"},
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:    "valid blogpost type",
			params:  SearchParams{Type: "blogpost", Text: "news"},
			want:    "type=blogpost and text ~ \"news\"",
			wantErr: false,
		},
		{
			name:    "valid comment type",
			params:  SearchParams{Type: "comment", Space: "DEV"},
			want:    "type=comment and space = \"DEV\"",
			wantErr: false,
		},
		{
			name:        "CQL injection via space key - double quote",
			params:      SearchParams{Space: "DEV\"evil"},
			wantErr:     true,
			errContains: "invalid space key",
		},
		{
			name:        "CQL injection via space key - OR clause",
			params:      SearchParams{Space: "DEV\" OR type=blogpost OR space=\"EVIL"},
			wantErr:     true,
			errContains: "invalid space key",
		},
		{
			name:        "CQL injection via space key - semicolon",
			params:      SearchParams{Space: "DEV;DROP TABLE"},
			wantErr:     true,
			errContains: "invalid space key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildCQL(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("BuildCQL() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BuildCQL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClient_Search(t *testing.T) {
	tests := []struct {
		name           string
		cql            string
		limit          int
		cursor         string
		statusCode     int
		response       any
		wantErr        bool
		wantNextCursor string
		errContains    string
		checkResult    func(*testing.T, *SearchResponse, string)
	}{
		{
			name:           "successful search with results",
			cql:            "type=page and space=DEV",
			limit:          25,
			cursor:         "",
			statusCode:     http.StatusOK,
			wantNextCursor: "",
			response: SearchResponse{
				Results: []SearchResult{
					{
						Title:        "Test Page",
						Excerpt:      "...matching text...",
						URL:          "/wiki/spaces/DEV/pages/123456/Test+Page",
						LastModified: "2024-01-15T10:30:00.000Z",
						Content: SearchContent{
							ID:     "123456",
							Type:   "page",
							Status: "current",
							Space: SearchSpace{
								Key:  "DEV",
								Name: "Development",
							},
						},
					},
				},
				Start:          0,
				Limit:          25,
				Size:           1,
				TotalSize:      1,
				CQLQuery:       "type=page and space=DEV",
				SearchDuration: 45,
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *SearchResponse, nextCursor string) {
				if len(resp.Results) != 1 {
					t.Errorf("Results length = %d, want 1", len(resp.Results))
				}
				if resp.Results[0].Title != "Test Page" {
					t.Errorf("Result title = %q, want %q", resp.Results[0].Title, "Test Page")
				}
				if resp.Results[0].Content.Space.Key != "DEV" {
					t.Errorf("Space key = %q, want %q", resp.Results[0].Content.Space.Key, "DEV")
				}
				if resp.TotalSize != 1 {
					t.Errorf("TotalSize = %d, want 1", resp.TotalSize)
				}
			},
		},
		{
			name:           "empty search results",
			cql:            "type=page and text ~ \"nonexistent\"",
			limit:          25,
			cursor:         "",
			statusCode:     http.StatusOK,
			wantNextCursor: "",
			response: SearchResponse{
				Results:        []SearchResult{},
				Start:          0,
				Limit:          25,
				Size:           0,
				TotalSize:      0,
				CQLQuery:       "type=page and text ~ \"nonexistent\"",
				SearchDuration: 12,
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *SearchResponse, nextCursor string) {
				if len(resp.Results) != 0 {
					t.Errorf("Results length = %d, want 0", len(resp.Results))
				}
				if resp.TotalSize != 0 {
					t.Errorf("TotalSize = %d, want 0", resp.TotalSize)
				}
			},
		},
		{
			name:           "pagination with more results",
			cql:            "type=page",
			limit:          25,
			cursor:         "",
			statusCode:     http.StatusOK,
			wantNextCursor: "abc123",
			response: SearchResponse{
				Results: []SearchResult{
					{
						Title:        "Page 1",
						Excerpt:      "excerpt 1",
						URL:          "/wiki/spaces/DEV/pages/111/Page+1",
						LastModified: "2024-01-01T00:00:00.000Z",
						Content: SearchContent{
							ID:   "111",
							Type: "page",
							Space: SearchSpace{
								Key: "DEV",
							},
						},
					},
				},
				Start:     0,
				Limit:     25,
				Size:      25,
				TotalSize: 150,
				Links: SearchPaginationLinks{
					Next: "/rest/api/search?cql=type=page&limit=25&cursor=abc123",
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *SearchResponse, nextCursor string) {
				if resp.Links.Next == "" {
					t.Error("Expected next link for pagination")
				}
				if resp.TotalSize != 150 {
					t.Errorf("TotalSize = %d, want 150", resp.TotalSize)
				}
			},
		},
		{
			name:        "empty CQL",
			cql:         "",
			limit:       25,
			cursor:      "",
			wantErr:     true,
			errContains: "cql query cannot be empty",
		},
		{
			name:        "whitespace CQL",
			cql:         "   ",
			limit:       25,
			cursor:      "",
			wantErr:     true,
			errContains: "cql query cannot be empty",
		},
		{
			name:        "zero limit returns error",
			cql:         "type=page",
			limit:       0,
			cursor:      "",
			wantErr:     true,
			errContains: "limit must be greater than 0",
		},
		{
			name:        "negative limit returns error",
			cql:         "type=page",
			limit:       -5,
			cursor:      "",
			wantErr:     true,
			errContains: "limit must be greater than 0",
		},
		{
			name:        "400 bad request - invalid CQL",
			cql:         "invalid syntax",
			limit:       25,
			cursor:      "",
			statusCode:  http.StatusBadRequest,
			response:    map[string]string{"message": "Invalid CQL query"},
			wantErr:     true,
			errContains: "API error (status 400)",
		},
		{
			name:        "401 unauthorized",
			cql:         "type=page",
			limit:       25,
			cursor:      "",
			statusCode:  http.StatusUnauthorized,
			response:    map[string]string{"message": "Unauthorized"},
			wantErr:     true,
			errContains: "API error (status 401)",
		},
		{
			name:        "404 not found",
			cql:         "type=page",
			limit:       25,
			cursor:      "",
			statusCode:  http.StatusNotFound,
			response:    map[string]string{"message": "Not found"},
			wantErr:     true,
			errContains: "API error (status 404)",
		},
		{
			name:        "429 rate limited",
			cql:         "type=page",
			limit:       25,
			cursor:      "",
			statusCode:  http.StatusTooManyRequests,
			response:    map[string]string{"message": "Rate limit exceeded"},
			wantErr:     true,
			errContains: "API error (status 429)",
		},
		{
			name:        "500 server error",
			cql:         "type=page",
			limit:       25,
			cursor:      "",
			statusCode:  http.StatusInternalServerError,
			response:    map[string]string{"message": "Internal error"},
			wantErr:     true,
			errContains: "API error (status 500)",
		},
		{
			name:           "URL encoding special characters",
			cql:            "text ~ \"api docs\" and label = \"urgent!\"",
			limit:          25,
			cursor:         "",
			statusCode:     http.StatusOK,
			wantNextCursor: "",
			response: SearchResponse{
				Results:   []SearchResult{},
				Start:     0,
				Size:      0,
				TotalSize: 0,
			},
			wantErr: false,
		},
		{
			name:           "pagination with cursor parameter",
			cql:            "type=page",
			limit:          25,
			cursor:         "abc123",
			statusCode:     http.StatusOK,
			wantNextCursor: "def456",
			response: SearchResponse{
				Results: []SearchResult{
					{
						Title:   "Page 26",
						Excerpt: "excerpt",
						URL:     "/wiki/spaces/DEV/pages/226/Page+26",
						Content: SearchContent{
							ID:   "226",
							Type: "page",
							Space: SearchSpace{
								Key: "DEV",
							},
						},
					},
				},
				Start:     25,
				Limit:     25,
				Size:      25,
				TotalSize: 150,
				Links: SearchPaginationLinks{
					Next: "/rest/api/search?cql=type=page&limit=25&cursor=def456",
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *SearchResponse, nextCursor string) {
				if nextCursor != "def456" {
					t.Errorf("nextCursor = %q, want %q", nextCursor, "def456")
				}
			},
		},
		{
			name:           "no more results - empty next link",
			cql:            "type=page",
			limit:          25,
			cursor:         "xyz789",
			statusCode:     http.StatusOK,
			wantNextCursor: "",
			response: SearchResponse{
				Results: []SearchResult{
					{
						Title: "Last Page",
						Content: SearchContent{
							ID:   "999",
							Type: "page",
							Space: SearchSpace{
								Key: "DEV",
							},
						},
					},
				},
				Start:     100,
				Limit:     25,
				Size:      10,
				TotalSize: 110,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Skip server validation for validation error tests
				if tt.wantErr && (tt.cql == "" || strings.TrimSpace(tt.cql) == "" || tt.limit <= 0) {
					return
				}

				// Verify request method and path
				if r.Method != http.MethodGet {
					t.Errorf("Method = %q, want %q", r.Method, http.MethodGet)
					return
				}
				if !strings.HasPrefix(r.URL.Path, "/wiki/rest/api/search") {
					t.Errorf("Path = %q, want prefix %q", r.URL.Path, "/wiki/rest/api/search")
					return
				}

				// Verify CQL query parameter is present and URL-encoded
				cqlParam := r.URL.Query().Get("cql")
				if cqlParam == "" {
					t.Error("CQL query parameter is missing")
					return
				}

				// Verify limit parameter
				limitParam := r.URL.Query().Get("limit")
				if limitParam == "" {
					t.Error("Limit query parameter is missing")
					return
				}

				// Verify cursor parameter if provided
				cursorParam := r.URL.Query().Get("cursor")
				if tt.cursor != "" && cursorParam != tt.cursor {
					t.Errorf("Cursor parameter = %q, want %q", cursorParam, tt.cursor)
					return
				}

				// Verify excerpt parameter
				excerptParam := r.URL.Query().Get("excerpt")
				if excerptParam != "highlight" {
					t.Errorf("Excerpt parameter = %q, want %q", excerptParam, "highlight")
					return
				}

				// Verify auth header is set
				if r.Header.Get("Authorization") == "" {
					t.Error("Authorization header not set")
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "test@example.com", "token")
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			result, nextCursor, err := client.Search(context.Background(), tt.cql, tt.limit, tt.cursor)

			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Search() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if nextCursor != tt.wantNextCursor {
					t.Errorf("Search() nextCursor = %q, want %q", nextCursor, tt.wantNextCursor)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, result, nextCursor)
				}
			}
		})
	}
}

func TestClient_Search_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler should not be reached
		t.Error("Handler should not be called with canceled context")
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test@example.com", "token")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err = client.Search(ctx, "type=page", 25, "")
	if err == nil {
		t.Error("Search() with canceled context should return error")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Error should mention context cancellation, got: %v", err)
	}
}

func TestExtractCursorFromLink(t *testing.T) {
	tests := []struct {
		name     string
		nextLink string
		want     string
	}{
		{
			name:     "valid link with cursor",
			nextLink: "/rest/api/search?cql=type=page&limit=25&cursor=abc123",
			want:     "abc123",
		},
		{
			name:     "empty link",
			nextLink: "",
			want:     "",
		},
		{
			name:     "link without cursor",
			nextLink: "/rest/api/search?cql=type=page&limit=25",
			want:     "",
		},
		{
			name:     "complex cursor value",
			nextLink: "/rest/api/search?cql=type=page&limit=25&cursor=eyJsaW1pdCI6MjUsInN0YXJ0IjoyNX0%3D",
			want:     "eyJsaW1pdCI6MjUsInN0YXJ0IjoyNX0=",
		},
		{
			name:     "full URL with cursor",
			nextLink: "https://example.atlassian.net/wiki/rest/api/search?cql=type=page&cursor=xyz789",
			want:     "xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCursorFromLink(tt.nextLink)
			if got != tt.want {
				t.Errorf("extractCursorFromLink(%q) = %q, want %q", tt.nextLink, got, tt.want)
			}
		})
	}
}

func TestEscapeCQLString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "simple text",
			want:  "simple text",
		},
		{
			name:  "double quotes",
			input: "text with \"quotes\"",
			want:  "text with \\\"quotes\\\"",
		},
		{
			name:  "backslashes",
			input: "path\\to\\file",
			want:  "path\\\\to\\\\file",
		},
		{
			name:  "backslashes and quotes",
			input: "\\\"escaped\\\"",
			want:  "\\\\\\\"escaped\\\\\\\"",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only quotes",
			input: "\"\"\"",
			want:  "\\\"\\\"\\\"",
		},
		{
			name:  "only backslashes",
			input: "\\\\\\",
			want:  "\\\\\\\\\\\\",
		},
		{
			name:  "mixed special chars",
			input: "test\\\"value\"\\end",
			want:  "test\\\\\\\"value\\\"\\\\end",
		},
		// Test all CQL/Lucene special characters
		{
			name:  "plus sign",
			input: "C++ programming",
			want:  "C\\+\\+ programming",
		},
		{
			name:  "minus/hyphen",
			input: "test-case",
			want:  "test\\-case",
		},
		{
			name:  "ampersand",
			input: "rock & roll",
			want:  "rock \\& roll",
		},
		{
			name:  "pipe",
			input: "option1 | option2",
			want:  "option1 \\| option2",
		},
		{
			name:  "exclamation",
			input: "not!important",
			want:  "not\\!important",
		},
		{
			name:  "parentheses",
			input: "test (case)",
			want:  "test \\(case\\)",
		},
		{
			name:  "curly braces",
			input: "code {block}",
			want:  "code \\{block\\}",
		},
		{
			name:  "square brackets",
			input: "array[0]",
			want:  "array\\[0\\]",
		},
		{
			name:  "caret",
			input: "x^2",
			want:  "x\\^2",
		},
		{
			name:  "tilde",
			input: "~username",
			want:  "\\~username",
		},
		{
			name:  "asterisk",
			input: "wild*card",
			want:  "wild\\*card",
		},
		{
			name:  "question mark",
			input: "what?",
			want:  "what\\?",
		},
		{
			name:  "colon",
			input: "key:value",
			want:  "key\\:value",
		},
		{
			name:  "forward slash",
			input: "path/to/file",
			want:  "path\\/to\\/file",
		},
		{
			name:  "all special chars combined",
			input: "+-&|!(){}[]^\"~*?:\\/",
			want:  "\\+\\-\\&\\|\\!\\(\\)\\{\\}\\[\\]\\^\\\"\\~\\*\\?\\:\\\\\\/",
		},
		{
			name:  "CQL injection attempt - OR clause",
			input: "test\" OR type=attachment OR text~\"malicious",
			want:  "test\\\" OR type=attachment OR text\\~\\\"malicious",
		},
		{
			name:  "CQL injection attempt - function call",
			input: "test\") OR currentUser()=creator OR text~(\"attack",
			want:  "test\\\"\\) OR currentUser\\(\\)=creator OR text\\~\\(\\\"attack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeCQLString(tt.input)
			if got != tt.want {
				t.Errorf("escapeCQLString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateSpaceKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid uppercase key",
			key:     "DEV",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "DEV2024",
			wantErr: false,
		},
		{
			name:    "valid key with underscores",
			key:     "DEV_TEAM_2024",
			wantErr: false,
		},
		{
			name:    "valid personal space",
			key:     "~USERNAME",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: false,
		},
		{
			name:    "lowercase key",
			key:     "dev",
			wantErr: false,
		},
		{
			name:    "mixed case key",
			key:     "Dev",
			wantErr: false,
		},
		{
			name:        "key with hyphen",
			key:         "DEV-TEAM",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "key with spaces",
			key:         "DEV TEAM",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "key with special chars",
			key:         "DEV@TEAM",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "too long key",
			key:         strings.Repeat("A", 256),
			wantErr:     true,
			errContains: "space key too long",
		},
		{
			name:    "max length key",
			key:     strings.Repeat("A", 255),
			wantErr: false,
		},
		{
			name:        "only tilde",
			key:         "~",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "multiple tildes",
			key:         "~~USER",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - double quote",
			key:         "DEV\"evil",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - OR clause",
			key:         "DEV\" OR type=blogpost OR space=\"EVIL",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - AND clause",
			key:         "DEV\" AND creator=currentUser() AND space=\"EVIL",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - single quote",
			key:         "DEV'evil",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - semicolon",
			key:         "DEV;DROP TABLE",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - parentheses",
			key:         "DEV()",
			wantErr:     true,
			errContains: "invalid space key format",
		},
		{
			name:        "CQL injection attempt - equals sign",
			key:         "DEV=test",
			wantErr:     true,
			errContains: "invalid space key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpaceKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSpaceKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateSpaceKey(%q) error = %q, want containing %q", tt.key, err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid type - page",
			contentType: "page",
			wantErr:     false,
		},
		{
			name:        "valid type - blogpost",
			contentType: "blogpost",
			wantErr:     false,
		},
		{
			name:        "valid type - attachment",
			contentType: "attachment",
			wantErr:     false,
		},
		{
			name:        "valid type - comment",
			contentType: "comment",
			wantErr:     false,
		},
		{
			name:        "valid type - whiteboard",
			contentType: "whiteboard",
			wantErr:     false,
		},
		{
			name:        "valid type - database",
			contentType: "database",
			wantErr:     false,
		},
		{
			name:        "valid type - embed",
			contentType: "embed",
			wantErr:     false,
		},
		{
			name:        "valid type - folder",
			contentType: "folder",
			wantErr:     false,
		},
		{
			name:        "empty type",
			contentType: "",
			wantErr:     false,
		},
		{
			name:        "invalid type - unknown",
			contentType: "unknown",
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid type - space",
			contentType: "space",
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid type - custom",
			contentType: "custom",
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid type - uppercase",
			contentType: "PAGE",
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "invalid type - with special chars",
			contentType: "page;drop table",
			wantErr:     true,
			errContains: "invalid content type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentType(tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContentType(%q) error = %v, wantErr %v", tt.contentType, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateContentType(%q) error = %q, want containing %q", tt.contentType, err.Error(), tt.errContains)
				}
			}
		})
	}
}
