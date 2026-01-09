package cmd

import (
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
