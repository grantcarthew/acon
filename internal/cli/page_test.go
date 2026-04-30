package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/grantcarthew/acon/internal/config"
	"github.com/spf13/cobra"
)

func TestReadAndValidateContent_FileSizeLimits(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "small file",
			size:    1024, // 1KB
			wantErr: false,
		},
		{
			name:    "medium file",
			size:    1024 * 1024, // 1MB
			wantErr: false,
		},
		{
			name:    "at limit",
			size:    maxContentSize, // exactly 10MB
			wantErr: false,
		},
		{
			name:    "over limit",
			size:    maxContentSize + 1,
			wantErr: true,
			errMsg:  "file too large",
		},
		{
			name:    "well over limit",
			size:    maxContentSize + 1024*1024, // 11MB
			wantErr: true,
			errMsg:  "file too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with specified size
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")

			content := strings.Repeat("x", tt.size)
			if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := readAndValidateContent(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("readAndValidateContent() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("readAndValidateContent() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("readAndValidateContent() unexpected error = %v", err)
				return
			}

			if len(result) != tt.size {
				t.Errorf("readAndValidateContent() returned %d bytes, want %d", len(result), tt.size)
			}
		})
	}
}

func TestReadAndValidateContent_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.md")

	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := readAndValidateContent(tmpFile)
	if err == nil {
		t.Error("readAndValidateContent() expected error for empty file")
		return
	}
	if !strings.Contains(err.Error(), "content cannot be empty") {
		t.Errorf("readAndValidateContent() error = %q, want containing 'content cannot be empty'", err.Error())
	}
}

func TestReadAndValidateContent_WhitespaceOnlyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "whitespace.md")

	if err := os.WriteFile(tmpFile, []byte("   \n\t\n   "), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := readAndValidateContent(tmpFile)
	if err == nil {
		t.Error("readAndValidateContent() expected error for whitespace-only file")
		return
	}
	if !strings.Contains(err.Error(), "content cannot be empty") {
		t.Errorf("readAndValidateContent() error = %q, want containing 'content cannot be empty'", err.Error())
	}
}

func TestReadAndValidateContent_NonexistentFile(t *testing.T) {
	_, err := readAndValidateContent("/nonexistent/path/file.md")
	if err == nil {
		t.Error("readAndValidateContent() expected error for nonexistent file")
		return
	}
}

func TestReadAndValidateContent_ContentTrimmed(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "padded.md")

	if err := os.WriteFile(tmpFile, []byte("  \n\n  content here  \n\n  "), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := readAndValidateContent(tmpFile)
	if err != nil {
		t.Fatalf("readAndValidateContent() unexpected error = %v", err)
	}

	if string(result) != "content here" {
		t.Errorf("readAndValidateContent() = %q, want %q", string(result), "content here")
	}
}

func TestMapChildSortValue(t *testing.T) {
	tests := []struct {
		name      string
		sort      string
		desc      bool
		wantValue string
		wantValid bool
	}{
		{"empty defaults to web", "", false, "child-position", true},
		{"empty desc defaults to web desc", "", true, "-child-position", true},
		{"web ascending", "web", false, "child-position", true},
		{"web descending", "web", true, "-child-position", true},
		{"created ascending", "created", false, "created-date", true},
		{"created descending", "created", true, "-created-date", true},
		{"modified ascending", "modified", false, "modified-date", true},
		{"modified descending", "modified", true, "-modified-date", true},
		{"id ascending", "id", false, "id", true},
		{"id descending", "id", true, "-id", true},
		{"title returns empty (client-side)", "title", false, "", true},
		{"title desc returns empty (client-side)", "title", true, "", true},
		{"invalid sort", "invalid", false, "", false},
		{"unknown sort", "unknown", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotValid := mapChildSortValue(tt.sort, tt.desc)
			if gotValue != tt.wantValue {
				t.Errorf("mapChildSortValue(%q, %v) value = %q, want %q", tt.sort, tt.desc, gotValue, tt.wantValue)
			}
			if gotValid != tt.wantValid {
				t.Errorf("mapChildSortValue(%q, %v) valid = %v, want %v", tt.sort, tt.desc, gotValid, tt.wantValid)
			}
		})
	}
}

func TestMapSpaceSortValue(t *testing.T) {
	tests := []struct {
		name string
		sort string
		desc bool
		want string
	}{
		{"empty no desc", "", false, ""},
		{"empty with desc", "", true, "-id"},
		{"title ascending", "title", false, "title"},
		{"title descending", "title", true, "-title"},
		{"created ascending", "created", false, "created-date"},
		{"created descending", "created", true, "-created-date"},
		{"modified ascending", "modified", false, "modified-date"},
		{"modified descending", "modified", true, "-modified-date"},
		{"id ascending", "id", false, "id"},
		{"id descending", "id", true, "-id"},
		{"invalid sort", "invalid", false, ""},
		{"web not valid for space", "web", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapSpaceSortValue(tt.sort, tt.desc)
			if got != tt.want {
				t.Errorf("mapSpaceSortValue(%q, %v) = %q, want %q", tt.sort, tt.desc, got, tt.want)
			}
		})
	}
}

// withMockStdin temporarily replaces stdinReader for testing and restores it after.
func withMockStdin(t *testing.T, content string) {
	t.Helper()
	originalReader := stdinReader
	originalStat := stdinStat
	t.Cleanup(func() {
		stdinReader = originalReader
		stdinStat = originalStat
	})
	stdinReader = strings.NewReader(content)
	// Mock stat to indicate piped input (not a terminal)
	stdinStat = func() (os.FileInfo, error) {
		return nil, nil // Won't be called when pageFile is "-"
	}
}

func TestReadAndValidateContent_StdinWithDash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "normal content",
			input:   "# Hello World\n\nThis is content.",
			want:    "# Hello World\n\nThis is content.",
			wantErr: false,
		},
		{
			name:    "content with surrounding whitespace",
			input:   "\n\n  # Trimmed  \n\n",
			want:    "# Trimmed",
			wantErr: false,
		},
		{
			name:    "empty stdin",
			input:   "",
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name:    "whitespace only stdin",
			input:   "   \n\t\n   ",
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name:    "single character",
			input:   "x",
			want:    "x",
			wantErr: false,
		},
		{
			name:    "markdown with code block",
			input:   "# Title\n\n```go\nfunc main() {}\n```\n",
			want:    "# Title\n\n```go\nfunc main() {}\n```",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockStdin(t, tt.input)

			result, err := readAndValidateContent("-")

			if tt.wantErr {
				if err == nil {
					t.Errorf("readAndValidateContent(\"-\") expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("readAndValidateContent(\"-\") error = %q, want containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("readAndValidateContent(\"-\") unexpected error = %v", err)
				return
			}

			if string(result) != tt.want {
				t.Errorf("readAndValidateContent(\"-\") = %q, want %q", string(result), tt.want)
			}
		})
	}
}

func TestReadAndValidateContent_StdinSizeLimit(t *testing.T) {
	// Create content just over the limit
	overLimitContent := strings.Repeat("x", maxContentSize+1)

	originalReader := stdinReader
	originalStat := stdinStat
	t.Cleanup(func() {
		stdinReader = originalReader
		stdinStat = originalStat
	})
	stdinReader = strings.NewReader(overLimitContent)
	stdinStat = func() (os.FileInfo, error) { return nil, nil }

	_, err := readAndValidateContent("-")
	if err == nil {
		t.Error("readAndValidateContent(\"-\") expected error for oversized stdin")
		return
	}
	if !strings.Contains(err.Error(), "stdin too large") {
		t.Errorf("readAndValidateContent(\"-\") error = %q, want containing 'stdin too large'", err.Error())
	}
}

func TestReadAndValidateContent_StdinAtLimit(t *testing.T) {
	// Create content exactly at the limit
	atLimitContent := strings.Repeat("x", maxContentSize)

	originalReader := stdinReader
	originalStat := stdinStat
	t.Cleanup(func() {
		stdinReader = originalReader
		stdinStat = originalStat
	})
	stdinReader = strings.NewReader(atLimitContent)
	stdinStat = func() (os.FileInfo, error) { return nil, nil }

	result, err := readAndValidateContent("-")
	if err != nil {
		t.Errorf("readAndValidateContent(\"-\") unexpected error = %v", err)
		return
	}
	if len(result) != maxContentSize {
		t.Errorf("readAndValidateContent(\"-\") returned %d bytes, want %d", len(result), maxContentSize)
	}
}

func TestReadAndValidateContent_DashIsNotFilePath(t *testing.T) {
	// Ensure that "-" is treated as stdin, not as a file path
	// Even if a file named "-" exists, we should read from stdin
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	_ = os.Chdir(tmpDir)                            //nolint:errcheck
	t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck

	// Create a file literally named "-"
	if err := os.WriteFile("-", []byte("file content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock stdin with different content
	originalReader := stdinReader
	originalStat := stdinStat
	t.Cleanup(func() {
		stdinReader = originalReader
		stdinStat = originalStat
	})
	stdinReader = strings.NewReader("stdin content")
	stdinStat = func() (os.FileInfo, error) { return nil, nil }

	result, err := readAndValidateContent("-")
	if err != nil {
		t.Errorf("readAndValidateContent(\"-\") unexpected error = %v", err)
		return
	}

	// Should get stdin content, not file content
	if string(result) != "stdin content" {
		t.Errorf("readAndValidateContent(\"-\") = %q, want %q (stdin should take precedence over file)", string(result), "stdin content")
	}
}

func TestReadAndValidateContent_StdinReadError(t *testing.T) {
	originalReader := stdinReader
	originalStat := stdinStat
	t.Cleanup(func() {
		stdinReader = originalReader
		stdinStat = originalStat
	})

	// Create a reader that returns an error
	stdinReader = &errorReader{err: io.ErrUnexpectedEOF}
	stdinStat = func() (os.FileInfo, error) { return nil, nil }

	_, err := readAndValidateContent("-")
	if err == nil {
		t.Error("readAndValidateContent(\"-\") expected error for read failure")
		return
	}
	if !strings.Contains(err.Error(), "reading stdin") {
		t.Errorf("readAndValidateContent(\"-\") error = %q, want containing 'reading stdin'", err.Error())
	}
}

// errorReader is an io.Reader that always returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func Test_pageURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		spaceKey string
		pageID   string
		want     string
	}{
		{
			name:     "with space key",
			baseURL:  "https://example.atlassian.net",
			spaceKey: "MYSPACE",
			pageID:   "12345",
			want:     "https://example.atlassian.net/wiki/spaces/MYSPACE/pages/12345",
		},
		{
			name:     "different base URL",
			baseURL:  "https://acme.atlassian.net",
			spaceKey: "DEV",
			pageID:   "67890",
			want:     "https://acme.atlassian.net/wiki/spaces/DEV/pages/67890",
		},
		{
			name:     "long page ID",
			baseURL:  "https://example.atlassian.net",
			spaceKey: "DOCS",
			pageID:   "1234567890123",
			want:     "https://example.atlassian.net/wiki/spaces/DOCS/pages/1234567890123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pageURL(tt.baseURL, tt.spaceKey, tt.pageID)
			if got != tt.want {
				t.Errorf("pageURL(%q, %q, %q) = %q, want %q", tt.baseURL, tt.spaceKey, tt.pageID, got, tt.want)
			}
		})
	}
}

// withMockClient substitutes newClient for the duration of the test.
func withMockClient(t *testing.T, client *api.Client, cfg *config.Config) {
	t.Helper()
	prev := newClient
	newClient = func() (*api.Client, *config.Config, error) {
		return client, cfg, nil
	}
	t.Cleanup(func() { newClient = prev })
}

// resetPageFlags restores package-level flag vars to their defaults and
// ensures they are reset again after the test.
func resetPageFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		pageTitle = ""
		pageFile = ""
		pageSpace = ""
		pageParent = ""
		pageLimit = 25
		pageSort = ""
		pageDesc = false
		outputJSON = false
		updateMsg = ""
		moveParent = ""
	}
	reset()
	t.Cleanup(reset)
}

// captureStdStreams replaces os.Stdout and os.Stderr with pipes. The returned
// finish function closes the pipes, drains them, restores the originals, and
// returns the captured text.
// Mutates package globals; tests using this helper must not call t.Parallel().
func captureStdStreams(t *testing.T) (finish func() (stdout, stderr string)) {
	t.Helper()
	origStdout, origStderr := os.Stdout, os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout, os.Stderr = wOut, wErr

	outCh := make(chan string, 1)
	errCh := make(chan string, 1)
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, rOut)
		outCh <- b.String()
	}()
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, rErr)
		errCh <- b.String()
	}()

	return func() (string, string) {
		_ = wOut.Close()
		_ = wErr.Close()
		stdout := <-outCh
		stderr := <-errCh
		os.Stdout, os.Stderr = origStdout, origStderr
		return stdout, stderr
	}
}

// testCommand returns a minimal cobra.Command carrying a background context,
// suitable for invoking a handler's RunE directly.
func testCommand() *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(context.Background())
	return c
}

// updateMoveHandler returns an http.Handler covering GetPage/UpdatePage/MovePage
// request flows and GetSpaceByID. spaceStatus controls the response code for
// the /spaces/{id} endpoint; when 200, spaceKey is returned in the body (use
// "" to exercise the empty-key warning path).
func updateMoveHandler(t *testing.T, spaceStatus int, spaceKey string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		// GetPage: /wiki/api/v2/pages/{id}?body-format=storage
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/pages/") && !strings.HasSuffix(r.URL.Path, "/children"):
			id := strings.TrimPrefix(r.URL.Path, "/wiki/api/v2/pages/")
			_ = json.NewEncoder(w).Encode(api.Page{
				ID:      id,
				SpaceID: "space-1",
				Title:   "page-" + id,
				Version: &api.Version{Number: 3},
				Body:    &api.PageBodyGet{Storage: &api.BodyContent{Representation: "storage", Value: "<p>body</p>"}},
			})
		// UpdatePage: PUT /wiki/api/v2/pages/{id}
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/pages/"):
			id := strings.TrimPrefix(r.URL.Path, "/wiki/api/v2/pages/")
			_ = json.NewEncoder(w).Encode(api.Page{
				ID:      id,
				SpaceID: "space-1",
				Title:   "page-" + id,
				Version: &api.Version{Number: 4},
			})
		// GetSpaceByID: GET /wiki/api/v2/spaces/{id}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/"):
			if spaceStatus != http.StatusOK {
				w.WriteHeader(spaceStatus)
				_, _ = w.Write([]byte(`{"message":"boom"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(api.Space{ID: "space-1", Key: spaceKey, Name: "My Space"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func TestPageUpdateCmd_HappyPath(t *testing.T) {
	resetPageFlags(t)
	pageFile = "-"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, "MYSPACE"))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})
	withMockStdin(t, "# updated body")

	finish := captureStdStreams(t)
	runErr := pageUpdateCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}
	wantURL := server.URL + "/wiki/spaces/MYSPACE/pages/123"
	if !strings.Contains(stdout, wantURL) {
		t.Errorf("stdout = %q, want containing %q", stdout, wantURL)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestPageUpdateCmd_SpaceLookupFails(t *testing.T) {
	resetPageFlags(t)
	pageFile = "-"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusInternalServerError, ""))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})
	withMockStdin(t, "# updated body")

	finish := captureStdStreams(t)
	runErr := pageUpdateCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE should succeed when only the space lookup fails, got: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "123" {
		t.Errorf("stdout = %q, want bare page ID %q", stdout, "123")
	}
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "page updated") {
		t.Errorf("stderr = %q, want warning about update space-key resolution", stderr)
	}
}

func TestPageMoveCmd_HappyPath(t *testing.T) {
	resetPageFlags(t)
	moveParent = "456"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, "MYSPACE"))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})

	finish := captureStdStreams(t)
	runErr := pageMoveCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}
	wantURL := server.URL + "/wiki/spaces/MYSPACE/pages/123"
	if !strings.Contains(stdout, wantURL) {
		t.Errorf("stdout = %q, want containing %q", stdout, wantURL)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestPageMoveCmd_SpaceLookupFails(t *testing.T) {
	resetPageFlags(t)
	moveParent = "456"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusInternalServerError, ""))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})

	finish := captureStdStreams(t)
	runErr := pageMoveCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE should succeed when only the space lookup fails, got: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "123" {
		t.Errorf("stdout = %q, want bare page ID %q", stdout, "123")
	}
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "page moved") {
		t.Errorf("stderr = %q, want warning about move space-key resolution", stderr)
	}
}

func TestPageListCmd_SpaceBranch_NoExtraLookups(t *testing.T) {
	resetPageFlags(t)
	pageSpace = "MYSPACE"

	var spaceByIDHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		// GetSpace by key: /wiki/api/v2/spaces?keys=MYSPACE
		case r.Method == http.MethodGet && r.URL.Path == "/wiki/api/v2/spaces":
			_ = json.NewEncoder(w).Encode(api.SpaceListResponse{
				Results: []api.Space{{ID: "space-1", Key: "MYSPACE", Name: "My Space"}},
			})
		// ListPages: /wiki/api/v2/pages?space-id=...
		case r.Method == http.MethodGet && r.URL.Path == "/wiki/api/v2/pages":
			_ = json.NewEncoder(w).Encode(api.PageListResponse{
				Results: []api.Page{
					{ID: "1", SpaceID: "space-1", Title: "A", Status: "current"},
					{ID: "2", SpaceID: "space-1", Title: "B", Status: "current"},
					{ID: "3", SpaceID: "space-1", Title: "C", Status: "current"},
				},
			})
		// GetSpaceByID should not be called in this branch.
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/"):
			spaceByIDHits.Add(1)
			_ = json.NewEncoder(w).Encode(api.Space{ID: "space-1", Key: "MYSPACE"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cfg := &config.Config{BaseURL: server.URL}

	pages, _, cache, err := listPagesBySpace(context.Background(), client, cfg)
	if err != nil {
		t.Fatalf("listPagesBySpace: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3", len(pages))
	}
	if cache["space-1"] != "MYSPACE" {
		t.Errorf("cache = %v, want {space-1: MYSPACE}", cache)
	}

	var buf bytes.Buffer
	if err := printPageList(context.Background(), client, &buf, cfg.BaseURL, pages, false, cache); err != nil {
		t.Fatalf("printPageList: %v", err)
	}
	if got := spaceByIDHits.Load(); got != 0 {
		t.Errorf("GetSpaceByID hits = %d, want 0 (cache should have been primed)", got)
	}
	wantURL := server.URL + "/wiki/spaces/MYSPACE/pages/1"
	if !strings.Contains(buf.String(), wantURL) {
		t.Errorf("output missing canonical URL %q in: %s", wantURL, buf.String())
	}
}

func TestPageListCmd_ParentBranch_CacheDedup(t *testing.T) {
	resetPageFlags(t)
	pageParent = "999"

	var spaceByIDHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		// GetChildPages: /wiki/api/v2/pages/{parentID}/children
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/children"):
			_ = json.NewEncoder(w).Encode(api.PageListResponse{
				Results: []api.Page{
					{ID: "1", SpaceID: "space-1", Title: "A", Status: "current"},
					{ID: "2", SpaceID: "space-1", Title: "B", Status: "current"},
					{ID: "3", SpaceID: "space-1", Title: "C", Status: "current"},
					{ID: "4", SpaceID: "space-1", Title: "D", Status: "current"},
					{ID: "5", SpaceID: "space-1", Title: "E", Status: "current"},
				},
			})
		// GetSpaceByID: counted.
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/"):
			spaceByIDHits.Add(1)
			_ = json.NewEncoder(w).Encode(api.Space{ID: "space-1", Key: "MYSPACE"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cfg := &config.Config{BaseURL: server.URL}

	pages, _, cache, err := listChildPages(context.Background(), client)
	if err != nil {
		t.Fatalf("listChildPages: %v", err)
	}
	if len(pages) != 5 {
		t.Fatalf("got %d pages, want 5", len(pages))
	}
	if len(cache) != 0 {
		t.Errorf("initial cache = %v, want empty", cache)
	}

	var buf bytes.Buffer
	if err := printPageList(context.Background(), client, &buf, cfg.BaseURL, pages, false, cache); err != nil {
		t.Fatalf("printPageList: %v", err)
	}
	if got := spaceByIDHits.Load(); got != 1 {
		t.Errorf("GetSpaceByID hits = %d, want 1 (cache should dedup after first miss)", got)
	}
	if cache["space-1"] != "MYSPACE" {
		t.Errorf("cache = %v, want {space-1: MYSPACE} after print", cache)
	}
}

func TestPageUpdateCmd_SpaceEmptyKey(t *testing.T) {
	resetPageFlags(t)
	pageFile = "-"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, ""))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})
	withMockStdin(t, "# updated body")

	finish := captureStdStreams(t)
	runErr := pageUpdateCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE should succeed when only the space lookup returns empty key, got: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "123" {
		t.Errorf("stdout = %q, want bare page ID %q", stdout, "123")
	}
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "page updated") || !strings.Contains(stderr, "returned empty key") {
		t.Errorf("stderr = %q, want warning about empty space key on update", stderr)
	}
}

func TestPageUpdateCmd_JSONOutput(t *testing.T) {
	resetPageFlags(t)
	pageFile = "-"
	outputJSON = true

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, "MYSPACE"))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})
	withMockStdin(t, "# updated body")

	finish := captureStdStreams(t)
	runErr := pageUpdateCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var got api.Page
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout = %q", err, stdout)
	}
	if got.ID != "123" {
		t.Errorf("got.ID = %q, want %q", got.ID, "123")
	}
	if strings.Contains(stdout, "/wiki/spaces/") {
		t.Errorf("stdout contains a URL, want JSON only: %q", stdout)
	}
}

func TestPageMoveCmd_SpaceEmptyKey(t *testing.T) {
	resetPageFlags(t)
	moveParent = "456"

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, ""))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})

	finish := captureStdStreams(t)
	runErr := pageMoveCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE should succeed when only the space lookup returns empty key, got: %v", runErr)
	}
	if strings.TrimSpace(stdout) != "123" {
		t.Errorf("stdout = %q, want bare page ID %q", stdout, "123")
	}
	if !strings.Contains(stderr, "Warning") || !strings.Contains(stderr, "page moved") || !strings.Contains(stderr, "returned empty key") {
		t.Errorf("stderr = %q, want warning about empty space key on move", stderr)
	}
}

func TestPageMoveCmd_JSONOutput(t *testing.T) {
	resetPageFlags(t)
	moveParent = "456"
	outputJSON = true

	server := httptest.NewServer(updateMoveHandler(t, http.StatusOK, "MYSPACE"))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})

	finish := captureStdStreams(t)
	runErr := pageMoveCmd.RunE(testCommand(), []string{"123"})
	stdout, stderr := finish()

	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var got api.Page
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout = %q", err, stdout)
	}
	if got.ID != "123" {
		t.Errorf("got.ID = %q, want %q", got.ID, "123")
	}
	if strings.Contains(stdout, "/wiki/spaces/") {
		t.Errorf("stdout contains a URL, want JSON only: %q", stdout)
	}
}

func TestPageMoveCmd_MissingParent(t *testing.T) {
	resetPageFlags(t)
	// moveParent intentionally left empty

	// Server should not be hit — MovePage validates before the API call.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	withMockClient(t, client, &config.Config{BaseURL: server.URL})

	runErr := pageMoveCmd.RunE(testCommand(), []string{"123"})
	if runErr == nil {
		t.Fatal("RunE expected error for missing --parent, got nil")
	}
	if !strings.Contains(runErr.Error(), "--parent flag is required") {
		t.Errorf("err = %v, want containing '--parent flag is required'", runErr)
	}
}

// errClient is an *api.Client built against a test server that returns 500 for
// every request — used by tests that should never reach an HTTP call.
func errClient(t *testing.T) (*api.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"unexpected request"}`))
	}))
	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		server.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return client, server
}

func TestListPagesBySpace_MissingSpaceKey(t *testing.T) {
	resetPageFlags(t)
	// pageSpace and cfg.SpaceKey both empty.
	client, server := errClient(t)
	defer server.Close()

	_, _, _, err := listPagesBySpace(context.Background(), client, &config.Config{BaseURL: server.URL})
	if err == nil {
		t.Fatal("listPagesBySpace expected error for missing space key, got nil")
	}
	if !strings.Contains(err.Error(), "space key required") {
		t.Errorf("err = %v, want containing 'space key required'", err)
	}
}

func TestListPagesBySpace_InvalidSort(t *testing.T) {
	resetPageFlags(t)
	pageSpace = "MYSPACE"
	pageSort = "invalid"

	client, server := errClient(t)
	defer server.Close()

	_, _, _, err := listPagesBySpace(context.Background(), client, &config.Config{BaseURL: server.URL})
	if err == nil {
		t.Fatal("listPagesBySpace expected error for invalid sort, got nil")
	}
	if !strings.Contains(err.Error(), "invalid sort value") {
		t.Errorf("err = %v, want containing 'invalid sort value'", err)
	}
}

func TestListPagesBySpace_GetSpaceFails(t *testing.T) {
	resetPageFlags(t)
	pageSpace = "MYSPACE"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/wiki/api/v2/spaces" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, _, _, err = listPagesBySpace(context.Background(), client, &config.Config{BaseURL: server.URL})
	if err == nil {
		t.Fatal("listPagesBySpace expected error from GetSpace, got nil")
	}
	if !strings.Contains(err.Error(), "getting space") {
		t.Errorf("err = %v, want containing 'getting space'", err)
	}
}

func TestListPagesBySpace_ListPagesFails(t *testing.T) {
	resetPageFlags(t)
	pageSpace = "MYSPACE"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/wiki/api/v2/spaces":
			_ = json.NewEncoder(w).Encode(api.SpaceListResponse{
				Results: []api.Space{{ID: "space-1", Key: "MYSPACE"}},
			})
		case "/wiki/api/v2/pages":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, _, _, err = listPagesBySpace(context.Background(), client, &config.Config{BaseURL: server.URL})
	if err == nil {
		t.Fatal("listPagesBySpace expected error from ListPages, got nil")
	}
	if !strings.Contains(err.Error(), "listing pages") {
		t.Errorf("err = %v, want containing 'listing pages'", err)
	}
}

func TestListChildPages_InvalidSort(t *testing.T) {
	resetPageFlags(t)
	pageParent = "999"
	pageSort = "invalid"

	client, server := errClient(t)
	defer server.Close()

	_, _, _, err := listChildPages(context.Background(), client)
	if err == nil {
		t.Fatal("listChildPages expected error for invalid sort, got nil")
	}
	if !strings.Contains(err.Error(), "invalid sort value") {
		t.Errorf("err = %v, want containing 'invalid sort value'", err)
	}
}

func TestListChildPages_GetChildPagesFails(t *testing.T) {
	resetPageFlags(t)
	pageParent = "999"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, _, _, err = listChildPages(context.Background(), client)
	if err == nil {
		t.Fatal("listChildPages expected error from GetChildPages, got nil")
	}
	if !strings.Contains(err.Error(), "listing child pages") {
		t.Errorf("err = %v, want containing 'listing child pages'", err)
	}
}

// childPagesUnsortedHandler returns three pages with mixed-case titles in a
// non-alphabetical order so a client-side sort can be observed.
func childPagesUnsortedHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/children") {
			_ = json.NewEncoder(w).Encode(api.PageListResponse{
				Results: []api.Page{
					{ID: "1", SpaceID: "space-1", Title: "banana", Status: "current"},
					{ID: "2", SpaceID: "space-1", Title: "Apple", Status: "current"},
					{ID: "3", SpaceID: "space-1", Title: "cherry", Status: "current"},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
}

func TestListChildPages_TitleSort_Asc(t *testing.T) {
	resetPageFlags(t)
	pageParent = "999"
	pageSort = "title"

	server := httptest.NewServer(childPagesUnsortedHandler(t))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pages, _, _, err := listChildPages(context.Background(), client)
	if err != nil {
		t.Fatalf("listChildPages: %v", err)
	}
	wantTitles := []string{"Apple", "banana", "cherry"}
	if len(pages) != len(wantTitles) {
		t.Fatalf("got %d pages, want %d", len(pages), len(wantTitles))
	}
	for i, want := range wantTitles {
		if pages[i].Title != want {
			t.Errorf("pages[%d].Title = %q, want %q (sorted ascending, case-insensitive)", i, pages[i].Title, want)
		}
	}
}

func TestListChildPages_TitleSort_Desc(t *testing.T) {
	resetPageFlags(t)
	pageParent = "999"
	pageSort = "title"
	pageDesc = true

	server := httptest.NewServer(childPagesUnsortedHandler(t))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pages, _, _, err := listChildPages(context.Background(), client)
	if err != nil {
		t.Fatalf("listChildPages: %v", err)
	}
	wantTitles := []string{"cherry", "banana", "Apple"}
	if len(pages) != len(wantTitles) {
		t.Fatalf("got %d pages, want %d", len(pages), len(wantTitles))
	}
	for i, want := range wantTitles {
		if pages[i].Title != want {
			t.Errorf("pages[%d].Title = %q, want %q (sorted descending, case-insensitive)", i, pages[i].Title, want)
		}
	}
}

func TestPrintPageList_SinglePage_AllShown(t *testing.T) {
	client, server := errClient(t)
	defer server.Close()

	pages := []api.Page{{ID: "1", SpaceID: "space-1", Title: "Solo", Status: "current"}}
	cache := map[string]string{"space-1": "MYSPACE"}

	var buf bytes.Buffer
	if err := printPageList(context.Background(), client, &buf, "https://example.atlassian.net", pages, false, cache); err != nil {
		t.Fatalf("printPageList: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Showing all 1 result\n") {
		t.Errorf("output missing singular summary 'Showing all 1 result':\n%s", out)
	}
	if strings.Contains(out, "results") {
		t.Errorf("output contains plural 'results' for single page:\n%s", out)
	}
}

func TestPrintPageList_HasMore(t *testing.T) {
	client, server := errClient(t)
	defer server.Close()

	pages := []api.Page{
		{ID: "1", SpaceID: "space-1", Title: "A", Status: "current"},
		{ID: "2", SpaceID: "space-1", Title: "B", Status: "current"},
		{ID: "3", SpaceID: "space-1", Title: "C", Status: "current"},
	}
	cache := map[string]string{"space-1": "MYSPACE"}

	var buf bytes.Buffer
	if err := printPageList(context.Background(), client, &buf, "https://example.atlassian.net", pages, true, cache); err != nil {
		t.Fatalf("printPageList: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Showing 3 results (more available - increase --limit to see more)") {
		t.Errorf("output missing hasMore summary:\n%s", out)
	}
	if strings.Contains(out, "Showing all") {
		t.Errorf("output contains 'Showing all' when hasMore is true:\n%s", out)
	}
}

func TestPrintPageList_GetSpaceByIDError(t *testing.T) {
	var spaceByIDHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/") {
			spaceByIDHits.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pages := []api.Page{
		{ID: "1", SpaceID: "space-1", Title: "A", Status: "current"},
		{ID: "2", SpaceID: "space-1", Title: "B", Status: "current"},
	}
	cache := map[string]string{}

	finish := captureStdStreams(t)
	var buf bytes.Buffer
	printErr := printPageList(context.Background(), client, &buf, server.URL, pages, false, cache)
	_, stderr := finish()

	if printErr != nil {
		t.Fatalf("printPageList: %v", printErr)
	}
	if got := spaceByIDHits.Load(); got != 1 {
		t.Errorf("GetSpaceByID hits = %d, want 1 (negative cache should suppress second call)", got)
	}
	out := buf.String()
	if !strings.Contains(out, "URL: (unresolved, page ID: 1)") {
		t.Errorf("output missing unresolved URL line for page 1:\n%s", out)
	}
	if !strings.Contains(out, "URL: (unresolved, page ID: 2)") {
		t.Errorf("output missing unresolved URL line for page 2:\n%s", out)
	}
	if !strings.Contains(stderr, "Warning: could not resolve space key for page 1") {
		t.Errorf("stderr missing warning for page 1:\n%s", stderr)
	}
	if strings.Contains(stderr, "could not resolve space key for page 2") {
		t.Errorf("stderr should not warn twice for the same SpaceID:\n%s", stderr)
	}
	if cache["space-1"] != "" {
		t.Errorf("cache[space-1] = %q, want \"\" (negative-cached)", cache["space-1"])
	}
}

func TestPrintPageList_GetSpaceByIDEmptyKey(t *testing.T) {
	var spaceByIDHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/") {
			spaceByIDHits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Space{ID: "space-1", Key: ""})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pages := []api.Page{
		{ID: "1", SpaceID: "space-1", Title: "A", Status: "current"},
		{ID: "2", SpaceID: "space-1", Title: "B", Status: "current"},
	}
	cache := map[string]string{}

	finish := captureStdStreams(t)
	var buf bytes.Buffer
	printErr := printPageList(context.Background(), client, &buf, server.URL, pages, false, cache)
	_, stderr := finish()

	if printErr != nil {
		t.Fatalf("printPageList: %v", printErr)
	}
	if got := spaceByIDHits.Load(); got != 1 {
		t.Errorf("GetSpaceByID hits = %d, want 1 (negative cache should suppress second call)", got)
	}
	out := buf.String()
	if !strings.Contains(out, "URL: (unresolved, page ID: 1)") {
		t.Errorf("output missing unresolved URL line for page 1:\n%s", out)
	}
	if !strings.Contains(out, "URL: (unresolved, page ID: 2)") {
		t.Errorf("output missing unresolved URL line for page 2:\n%s", out)
	}
	if !strings.Contains(stderr, "Warning: space space-1 returned empty key for page 1") {
		t.Errorf("stderr missing empty-key warning for page 1:\n%s", stderr)
	}
	if strings.Contains(stderr, "returned empty key for page 2") {
		t.Errorf("stderr should not warn twice for the same SpaceID:\n%s", stderr)
	}
	if got, ok := cache["space-1"]; !ok || got != "" {
		t.Errorf("cache[space-1] = %q (ok=%v), want \"\" (negative-cached)", got, ok)
	}
}

func TestPrintPageList_MultipleSpaces(t *testing.T) {
	var spaceByIDHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/wiki/api/v2/spaces/") {
			spaceByIDHits.Add(1)
			id := strings.TrimPrefix(r.URL.Path, "/wiki/api/v2/spaces/")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.Space{ID: id, Key: strings.ToUpper(id)})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "e@x", "t")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pages := []api.Page{
		{ID: "1", SpaceID: "alpha", Title: "A", Status: "current"},
		{ID: "2", SpaceID: "beta", Title: "B", Status: "current"},
		{ID: "3", SpaceID: "alpha", Title: "C", Status: "current"},
	}
	cache := map[string]string{}

	var buf bytes.Buffer
	if err := printPageList(context.Background(), client, &buf, server.URL, pages, false, cache); err != nil {
		t.Fatalf("printPageList: %v", err)
	}

	if got := spaceByIDHits.Load(); got != 2 {
		t.Errorf("GetSpaceByID hits = %d, want 2 (one per unique SpaceID, third page is a cache hit)", got)
	}
	if cache["alpha"] != "ALPHA" || cache["beta"] != "BETA" {
		t.Errorf("cache = %v, want {alpha: ALPHA, beta: BETA}", cache)
	}

	out := buf.String()
	if !strings.Contains(out, server.URL+"/wiki/spaces/ALPHA/pages/1") {
		t.Errorf("output missing URL for page 1 in space alpha:\n%s", out)
	}
	if !strings.Contains(out, server.URL+"/wiki/spaces/BETA/pages/2") {
		t.Errorf("output missing URL for page 2 in space beta:\n%s", out)
	}
	if !strings.Contains(out, server.URL+"/wiki/spaces/ALPHA/pages/3") {
		t.Errorf("output missing URL for page 3 in space alpha:\n%s", out)
	}
}
