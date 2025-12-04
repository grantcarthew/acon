package converter

import (
	"strings"
	"testing"
)

func TestStorageToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		wantErr  bool
	}{
		{
			name:     "h1 heading",
			input:    "<h1>Hello World</h1>",
			contains: []string{"# Hello World"},
		},
		{
			name:     "h2 heading",
			input:    "<h2>Section</h2>",
			contains: []string{"## Section"},
		},
		{
			name:     "h3 heading",
			input:    "<h3>Subsection</h3>",
			contains: []string{"### Subsection"},
		},
		{
			name:     "bold text",
			input:    "<p>This is <strong>bold</strong> text</p>",
			contains: []string{"**bold**"},
		},
		{
			name:     "italic text",
			input:    "<p>This is <em>italic</em> text</p>",
			contains: []string{"*italic*"},
		},
		{
			name:     "inline code",
			input:    "<p>Use <code>fmt.Println</code> function</p>",
			contains: []string{"`fmt.Println`"},
		},
		{
			name:     "link",
			input:    `<p>Visit <a href="https://google.com">Google</a></p>`,
			contains: []string{"[Google](https://google.com)"},
		},
		{
			name:  "unordered list",
			input: "<ul><li>Item one</li><li>Item two</li></ul>",
			contains: []string{
				"- Item one",
				"- Item two",
			},
		},
		{
			name:  "ordered list",
			input: "<ol><li>First</li><li>Second</li></ol>",
			contains: []string{
				"1. First",
				"2. Second",
			},
		},
		{
			name:  "simple table",
			input: "<table><thead><tr><th>A</th><th>B</th></tr></thead><tbody><tr><td>1</td><td>2</td></tr></tbody></table>",
			contains: []string{
				"| A | B |",
				"| 1 | 2 |",
			},
		},
		{
			name:     "blockquote",
			input:    "<blockquote><p>This is a quote</p></blockquote>",
			contains: []string{"> This is a quote"},
		},
		{
			name:     "paragraph",
			input:    "<p>Just a paragraph.</p>",
			contains: []string{"Just a paragraph."},
		},
		{
			name:     "empty input",
			input:    "",
			contains: []string{},
		},
		{
			name:     "nested formatting",
			input:    "<p>This is <strong><em>bold italic</em></strong> text</p>",
			contains: []string{"***bold italic***"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StorageToMarkdown(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("StorageToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("StorageToMarkdown(%q)\n  got: %q\n  missing: %q", tt.input, result, want)
				}
			}
		})
	}
}

func TestStorageToMarkdown_ComplexDocument(t *testing.T) {
	input := `<h1>Title</h1>
<p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
<h2>List</h2>
<ul>
<li>Item 1</li>
<li>Item 2</li>
</ul>
<p>Done.</p>`

	result, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("StorageToMarkdown() error = %v", err)
	}

	expected := []string{
		"# Title",
		"**bold**",
		"*italic*",
		"## List",
		"- Item 1",
		"- Item 2",
		"Done.",
	}

	for _, want := range expected {
		if !strings.Contains(result, want) {
			t.Errorf("Complex document missing: %q\nGot: %s", want, result)
		}
	}
}
