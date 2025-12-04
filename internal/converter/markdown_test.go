package converter

import (
	"strings"
	"testing"
)

func TestMarkdownToStorage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // Substrings that should be present in output
	}{
		{
			name:     "h1 heading",
			input:    "# Hello World",
			contains: []string{"<h1", "Hello World", "</h1>"},
		},
		{
			name:     "h2 heading",
			input:    "## Section Title",
			contains: []string{"<h2", "Section Title", "</h2>"},
		},
		{
			name:     "h3 heading",
			input:    "### Subsection",
			contains: []string{"<h3", "Subsection", "</h3>"},
		},
		{
			name:     "bold text",
			input:    "This is **bold** text",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "italic text",
			input:    "This is *italic* text",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "inline code",
			input:    "Use the `fmt.Println` function",
			contains: []string{"<code>fmt.Println</code>"},
		},
		{
			name:  "code block with language",
			input: "```go\nfunc main() {}\n```",
			contains: []string{
				`ac:name="code"`,
				`ac:name="language"`,
				"go",
				"func main() {}",
			},
		},
		{
			name:  "code block without language",
			input: "```\nsome code\n```",
			contains: []string{
				`ac:name="code"`,
				"some code",
			},
		},
		{
			name:  "unordered list",
			input: "- Item one\n- Item two\n- Item three",
			contains: []string{
				"<ul>",
				"<li>",
				"Item one",
				"Item two",
				"Item three",
				"</li>",
				"</ul>",
			},
		},
		{
			name:  "ordered list",
			input: "1. First\n2. Second\n3. Third",
			contains: []string{
				"<ol>",
				"<li>",
				"First",
				"Second",
				"Third",
				"</li>",
				"</ol>",
			},
		},
		{
			name:     "link",
			input:    "Visit [Google](https://google.com) for search",
			contains: []string{`<a href="https://google.com"`, "Google", "</a>"},
		},
		{
			name:  "simple table",
			input: "| A | B |\n|---|---|\n| 1 | 2 |",
			contains: []string{
				"<table>",
				"<th>", "A", "B",
				"<td>", "1", "2",
				"</table>",
			},
		},
		{
			name:     "blockquote",
			input:    "> This is a quote",
			contains: []string{"<blockquote>", "This is a quote", "</blockquote>"},
		},
		{
			name:     "paragraph",
			input:    "Just a plain paragraph.",
			contains: []string{"<p>", "Just a plain paragraph.", "</p>"},
		},
		{
			name:     "empty input",
			input:    "",
			contains: []string{}, // Should not panic
		},
		{
			name:     "whitespace only",
			input:    "   \n\t\n   ",
			contains: []string{}, // Should not panic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownToStorage(tt.input)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("MarkdownToStorage(%q)\n  got: %q\n  missing: %q", tt.input, result, want)
				}
			}
		})
	}
}

func TestMarkdownToStorage_MultipleElements(t *testing.T) {
	input := `# Title

This is a paragraph with **bold** and *italic* text.

## Code Example

` + "```go" + `
func hello() {
    fmt.Println("Hello")
}
` + "```" + `

## List

- Item 1
- Item 2

Done.
`

	result := MarkdownToStorage(input)

	expected := []string{
		"<h1", "Title", "</h1>",
		"<strong>bold</strong>",
		"<em>italic</em>",
		"<h2", "Code Example", "</h2>",
		`ac:name="code"`,
		"func hello()",
		"<h2", "List", "</h2>",
		"<ul>",
		"Item 1",
		"Item 2",
		"</ul>",
		"Done.",
	}

	for _, want := range expected {
		if !strings.Contains(result, want) {
			t.Errorf("Complex document missing: %q\nGot: %s", want, result)
		}
	}
}
