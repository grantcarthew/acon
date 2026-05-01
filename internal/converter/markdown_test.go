package converter

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

type mdCase struct {
	name     string
	input    string
	contains []string // substrings that must be present in output
	excludes []string // substrings that must NOT be present in output
}

func runMarkdownCases(t *testing.T, cases []mdCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownToStorage(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("MarkdownToStorage(%q)\n  got: %q\n  missing: %q", tt.input, result, want)
				}
			}
			for _, unwanted := range tt.excludes {
				if strings.Contains(result, unwanted) {
					t.Errorf("MarkdownToStorage(%q)\n  got: %q\n  unexpected: %q", tt.input, result, unwanted)
				}
			}
		})
	}
}

func TestMarkdownToStorage_Headings(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
			name:     "h4 heading",
			input:    "#### Level 4",
			contains: []string{"<h4", "Level 4", "</h4>"},
		},
		{
			name:     "h5 heading",
			input:    "##### Level 5",
			contains: []string{"<h5", "Level 5", "</h5>"},
		},
		{
			name:     "h6 heading",
			input:    "###### Level 6",
			contains: []string{"<h6", "Level 6", "</h6>"},
		},
		{
			name:     "unicode in heading",
			input:    "# 日本語 🚀 título",
			contains: []string{"<h1", "日本語", "🚀", "título", "</h1>"},
		},
	})
}

func TestMarkdownToStorage_InlineFormatting(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
			name:     "bold and italic combined",
			input:    "***both***",
			contains: []string{"<strong>", "<em>", "both"},
		},
		{
			name:     "strikethrough",
			input:    "Some ~~deleted~~ text",
			contains: []string{"<del>deleted</del>"},
		},
	})
}

func TestMarkdownToStorage_InlineCode(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "inline code",
			input:    "Use the `fmt.Println` function",
			contains: []string{"<code>fmt.Println</code>"},
		},
		{
			name:     "inline code escapes angle brackets",
			input:    "Replace `<name>` with the value",
			contains: []string{"<code>&lt;name&gt;</code>"},
		},
		{
			name:     "inline code escapes ampersand",
			input:    "Use `&cobra.Command{}` to construct",
			contains: []string{"<code>&amp;cobra.Command{}</code>"},
		},
		{
			name:     "inline code in list item escapes angle brackets",
			input:    "- `<name>`",
			contains: []string{"<li><code>&lt;name&gt;</code>"},
		},
		{
			name:     "double-backtick code with embedded backtick",
			input:    "Use `` `code` `` here",
			contains: []string{"<code>`code`</code>"},
		},
		{
			name:     "code span preserves backslashes",
			input:    "Use `path\\to\\file` here",
			contains: []string{`<code>path\to\file</code>`},
		},
		{
			name:     "unicode in code span",
			input:    "Use `日本語` for this",
			contains: []string{"<code>日本語</code>"},
		},
		{
			name:     "code span at start of paragraph",
			input:    "`first` followed by text",
			contains: []string{"<p><code>first</code>"},
		},
	})
}

func TestMarkdownToStorage_CodeBlocks(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
			name:  "indented code block",
			input: "Some text\n\n    indented code line\n    second line\n\nmore text",
			contains: []string{
				`ac:name="code"`,
				`ac:name="language"`,
				"none",
				"indented code line",
				"second line",
			},
		},
		{
			name:  "fenced code preserves html-significant chars verbatim",
			input: "```html\n<div>a < b && c > d</div>\n```",
			contains: []string{
				`ac:name="code"`,
				"<div>a < b && c > d</div>",
			},
		},
		{
			name:     "consecutive code blocks",
			input:    "```go\nfirst\n```\n\n```python\nsecond\n```",
			contains: []string{`ac:name="code"`, "first", "python", "second"},
		},
	})
}

func TestMarkdownToStorage_Lists(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
			name:  "nested unordered list",
			input: "- outer\n  - inner one\n  - inner two",
			contains: []string{
				"<ul>",
				"outer",
				"inner one",
				"inner two",
			},
		},
		{
			name:  "nested ordered inside unordered",
			input: "- outer\n  1. inner first\n  2. inner second",
			contains: []string{
				"<ul>",
				"<ol>",
				"outer",
				"inner first",
				"inner second",
			},
		},
	})
}

func TestMarkdownToStorage_TaskLists(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "task list unchecked",
			input:    "- [ ] todo item",
			contains: []string{"<ac:task-list>", "<ac:task>", "<ac:task-status>incomplete</ac:task-status>", "<ac:task-body>", "todo item", "</ac:task-body>", "</ac:task>", "</ac:task-list>"},
		},
		{
			name:     "task list checked",
			input:    "- [x] done item",
			contains: []string{"<ac:task-list>", "<ac:task-status>complete</ac:task-status>", "done item"},
		},
		{
			name:     "task list mixed",
			input:    "- [ ] todo\n- [x] done\n- [ ] another",
			contains: []string{"incomplete", "complete", "todo", "done", "another"},
		},
		{
			name:     "task list omits paragraph wrapper inside body",
			input:    "- [ ] no paragraph wrapper",
			contains: []string{"<ac:task-body>", "no paragraph wrapper", "</ac:task-body>"},
			excludes: []string{"<p>no paragraph wrapper</p>"},
		},
	})
}

func TestMarkdownToStorage_Tables(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
			name:  "table with left alignment",
			input: "| A | B |\n|:---|---|\n| 1 | 2 |",
			contains: []string{
				"<table>", "text-align:left", "A", "1", "B", "2", "</table>",
			},
		},
		{
			name:  "table with center alignment",
			input: "| A | B |\n|:---:|---|\n| 1 | 2 |",
			contains: []string{
				"<table>", "text-align:center", "A", "1", "</table>",
			},
		},
		{
			name:  "table with right alignment",
			input: "| A | B |\n|---:|---|\n| 1 | 2 |",
			contains: []string{
				"<table>", "text-align:right", "A", "1", "</table>",
			},
		},
		{
			name:  "table basic structure",
			input: "| A |\n|---|\n| 1 |",
			contains: []string{
				"<table>", "<thead>", "<th>A</th>", "<tbody>", "<td>1</td>", "</table>",
			},
		},
	})
}

func TestMarkdownToStorage_LinksAndImages(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "link",
			input:    "Visit [Google](https://google.com) for search",
			contains: []string{`<a href="https://google.com"`, "Google", "</a>"},
		},
		{
			name:     "link with title attribute",
			input:    `[text](https://example.com "the title")`,
			contains: []string{`<a href="https://example.com"`, `title="the title"`, "text", "</a>"},
		},
		{
			name:     "link escapes ampersand in url",
			input:    "[link](https://example.com/?a=1&b=2)",
			contains: []string{"&amp;b=2"},
		},
		{
			name:     "link escapes html in title",
			input:    `[text](https://example.com "a <b> & c")`,
			contains: []string{`title="a &lt;b&gt; &amp; c"`},
		},
		{
			name:     "autolink url",
			input:    "<https://example.com>",
			contains: []string{`<a href="https://example.com">`, "https://example.com", "</a>"},
		},
		{
			name:     "autolink email",
			input:    "<user@example.com>",
			contains: []string{`<a href="`, "user@example.com", "</a>"},
		},
		{
			name:     "autolink escapes ampersand in url",
			input:    "<https://example.com/path?a=1&b=2>",
			contains: []string{"&amp;b=2"},
		},
		{
			name:     "image basic",
			input:    "![alt text](https://example.com/img.png)",
			contains: []string{"<ac:image>", `ri:value="https://example.com/img.png"`, "</ac:image>"},
		},
		{
			name:     "image escapes ampersand in url",
			input:    "![](https://example.com/img.png?a=1&b=2)",
			contains: []string{"&amp;b=2"},
		},
	})
}

func TestMarkdownToStorage_Blocks(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "paragraph",
			input:    "Just a plain paragraph.",
			contains: []string{"<p>", "Just a plain paragraph.", "</p>"},
		},
		{
			name:     "blockquote",
			input:    "> This is a quote",
			contains: []string{"<blockquote>", "This is a quote", "</blockquote>"},
		},
		{
			name:     "blockquote with formatting",
			input:    "> **bold** and *italic* in quote",
			contains: []string{"<blockquote>", "<strong>bold</strong>", "<em>italic</em>", "</blockquote>"},
		},
		{
			name:     "nested blockquote",
			input:    "> outer\n>\n> > inner",
			contains: []string{"<blockquote>", "outer", "inner"},
		},
		{
			name:     "blockquote containing code block",
			input:    "> ```\n> code in quote\n> ```",
			contains: []string{"<blockquote>", `ac:name="code"`, "code in quote"},
		},
		{
			name:     "thematic break with dashes",
			input:    "before\n\n---\n\nafter",
			contains: []string{"<hr />"},
		},
		{
			name:     "thematic break with asterisks",
			input:    "before\n\n***\n\nafter",
			contains: []string{"<hr />"},
		},
		{
			name:     "thematic break with underscores",
			input:    "before\n\n___\n\nafter",
			contains: []string{"<hr />"},
		},
		{
			name:     "html block omitted",
			input:    "<div>raw block html</div>",
			contains: []string{"raw HTML omitted"},
			excludes: []string{"<div>", "raw block html"},
		},
		{
			name:     "inline raw html stripped",
			input:    "text <span>raw inline</span> more",
			contains: []string{"text", "more"},
			excludes: []string{"<span>", "</span>"},
		},
	})
}

func TestMarkdownToStorage_LineBreaks(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "hard line break with trailing spaces",
			input:    "first line  \nsecond line",
			contains: []string{"<br />"},
		},
		{
			name:     "hard line break with backslash",
			input:    "first line\\\nsecond line",
			contains: []string{"<br />"},
		},
		{
			name:     "soft line break",
			input:    "first line\nsecond line",
			contains: []string{"first line", "\n", "second line"},
		},
	})
}

func TestMarkdownToStorage_Escaping(t *testing.T) {
	runMarkdownCases(t, []mdCase{
		{
			name:     "text with html-significant chars escapes",
			input:    "5 < 10 && 10 > 5",
			contains: []string{"5 &lt; 10", "&amp;&amp;", "10 &gt; 5"},
		},
		{
			name:     "escaped markdown chars in text",
			input:    "Literal \\* asterisk and \\_ underscore",
			contains: []string{"Literal", "asterisk", "underscore"},
			excludes: []string{"<em>", "<strong>"},
		},
	})
}

func TestMarkdownToStorage_Edge(t *testing.T) {
	runMarkdownCases(t, []mdCase{
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
	})
}

// TestRenderStringDirect exercises renderString directly because goldmark's
// default parsers do not emit ast.String nodes from plain markdown — they
// come from extensions or programmatic AST construction.
//
// This is a defensive contract test, not a coverage signal: renderString is
// registered for ast.KindString in case future extensions emit String nodes.
// The function does not run in the live MarkdownToStorage pipeline today.
func TestRenderStringDirect(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"plain text", "hello", "hello"},
		{"escapes angle brackets", "a < b > c", "a &lt; b &gt; c"},
		{"escapes ampersand", "x & y", "x &amp; y"},
		{"escapes double quote", `say "hi"`, "say &quot;hi&quot;"},
		{"empty value", "", ""},
		{"unicode preserved", "日本語", "日本語"},
	}

	r := &ConfluenceRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			bw := bufio.NewWriter(&buf)
			node := ast.NewString([]byte(tt.value))

			status, err := r.renderString(bw, nil, node, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("got status %v, want WalkContinue", status)
			}

			// Exit branch should be a no-op.
			if _, err := r.renderString(bw, nil, node, false); err != nil {
				t.Fatalf("exit branch error: %v", err)
			}

			_ = bw.Flush()
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
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
