package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(originalDir) })

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

func TestPageURL(t *testing.T) {
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
			name:     "without space key",
			baseURL:  "https://example.atlassian.net",
			spaceKey: "",
			pageID:   "12345",
			want:     "https://example.atlassian.net/wiki/pages/12345",
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
			got := PageURL(tt.baseURL, tt.spaceKey, tt.pageID)
			if got != tt.want {
				t.Errorf("PageURL(%q, %q, %q) = %q, want %q", tt.baseURL, tt.spaceKey, tt.pageID, got, tt.want)
			}
		})
	}
}
