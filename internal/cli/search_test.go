package cli

import (
	"strings"
	"testing"
)

func TestFormatExcerptForTerminal(t *testing.T) {
	tests := []struct {
		name       string
		excerpt    string
		searchTerm string
		exact      bool     // when true, assert got == want and ignore contains/excludes
		want       string   // used only when exact is true
		contains   []string // substrings that must appear in got
		excludes   []string // substrings that must not appear in got
	}{
		{
			name:    "empty excerpt returns empty",
			excerpt: "",
			exact:   true,
			want:    "",
		},
		{
			name:    "excerpt with only html tags returns empty",
			excerpt: "<p></p><div></div>",
			exact:   true,
			want:    "",
		},
		{
			name:       "no search term truncates from start",
			excerpt:    "This is a short excerpt without any search term to find.",
			searchTerm: "",
			contains:   []string{"This is a short excerpt"},
			excludes:   []string{"\033["},
		},
		{
			name:       "search term not in text falls back to start",
			excerpt:    "Some content here",
			searchTerm: "missing",
			contains:   []string{"Some content here"},
			excludes:   []string{"\033["},
		},
		{
			name:       "search term highlighted with ansi bold",
			excerpt:    "The quick brown fox jumps over the lazy dog",
			searchTerm: "fox",
			contains:   []string{"\033[1mfox\033[0m"},
		},
		{
			name:       "search term match is case insensitive",
			excerpt:    "The Quick Brown FOX jumps",
			searchTerm: "fox",
			contains:   []string{"\033[1mFOX\033[0m"},
		},
		{
			name:       "html tags stripped before matching",
			excerpt:    "<p>Hello <strong>world</strong></p>",
			searchTerm: "world",
			contains:   []string{"\033[1mworld\033[0m"},
			excludes:   []string{"<p>", "<strong>"},
		},
		{
			name:       "html entities decoded",
			excerpt:    "five &lt; ten &amp; ten &gt; five",
			searchTerm: "",
			contains:   []string{"five < ten & ten > five"},
		},
		{
			name:       "whitespace normalised",
			excerpt:    "many\n\n\nspaces    and\ttabs",
			searchTerm: "",
			contains:   []string{"many spaces and tabs"},
		},
		{
			name:       "match at start has no leading ellipsis",
			excerpt:    "fox jumps over the lazy dog and runs into the forest where it lives",
			searchTerm: "fox",
			excludes:   []string{"..."},
			contains:   []string{"\033[1mfox\033[0m"},
		},
		{
			name:       "match in middle has ellipsis on both sides",
			excerpt:    strings.Repeat("padding word ", 30) + " needle " + strings.Repeat("padding word ", 30),
			searchTerm: "needle",
			contains:   []string{"...", "\033[1mneedle\033[0m"},
		},
		{
			name:       "match near end has leading ellipsis",
			excerpt:    strings.Repeat("padding word ", 30) + " needle",
			searchTerm: "needle",
			contains:   []string{"...", "\033[1mneedle\033[0m"},
		},
		{
			name:       "long text without search term is truncated",
			excerpt:    strings.Repeat("word ", 100),
			searchTerm: "",
			contains:   []string{"..."},
		},
		{
			name:       "short text without search term is not truncated",
			excerpt:    "short",
			searchTerm: "",
			exact:      true,
			want:       "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatExcerptForTerminal(tt.excerpt, tt.searchTerm)

			if tt.exact {
				if got != tt.want {
					t.Errorf("formatExcerptForTerminal(%q, %q)\n  got:  %q\n  want: %q", tt.excerpt, tt.searchTerm, got, tt.want)
				}
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("formatExcerptForTerminal(%q, %q)\n  got: %q\n  missing: %q", tt.excerpt, tt.searchTerm, got, want)
				}
			}
			for _, unwanted := range tt.excludes {
				if strings.Contains(got, unwanted) {
					t.Errorf("formatExcerptForTerminal(%q, %q)\n  got: %q\n  unexpected: %q", tt.excerpt, tt.searchTerm, got, unwanted)
				}
			}
		})
	}
}

func TestTruncateExcerpt(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{
			name:   "shorter than max returns unchanged",
			text:   "short text",
			maxLen: 50,
			want:   "short text",
		},
		{
			name:   "exactly max length returns unchanged",
			text:   "exactly ten",
			maxLen: 11,
			want:   "exactly ten",
		},
		{
			name:   "truncates at word boundary when one exists past midpoint",
			text:   "the quick brown fox jumps over the lazy dog",
			maxLen: 20,
			want:   "the quick brown fox...",
		},
		{
			name:   "appends ellipsis when truncating",
			text:   "abcdefghij abcdefghij abcdefghij",
			maxLen: 15,
			want:   "abcdefghij...",
		},
		{
			name:   "no good word boundary keeps the hard cut",
			text:   "abcdefghijklmnopqrstuvwxyz",
			maxLen: 10,
			want:   "abcdefghij...",
		},
		{
			name:   "empty input returns empty",
			text:   "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateExcerpt(tt.text, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateExcerpt(%q, %d)\n  got:  %q\n  want: %q", tt.text, tt.maxLen, got, tt.want)
			}
		})
	}
}
