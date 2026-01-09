package converter

import (
	"os"
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
			name:     "HTML entities decoded",
			input:    "<p>Angle brackets &lt; and &gt; with ampersand &amp; decoded</p>",
			contains: []string{"Angle brackets < and > with ampersand & decoded"},
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

func TestStorageToMarkdown_CodeMacros(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "basic code block with language",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[fmt.Println("hello")]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", `fmt.Println("hello")`, "```"},
		},
		{
			name:     "code block without language",
			input:    `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[plain code]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```", "plain code"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "code block with schema-version",
			input:    `<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:plain-text-body><![CDATA[test]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```", "test"},
		},
		{
			name:     "code block with macro-id",
			input:    `<ac:structured-macro ac:name="code" ac:schema-version="1" ac:macro-id="abc-123"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[print("hi")]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```python", `print("hi")`},
		},
		{
			name:     "attributes in different order",
			input:    `<ac:structured-macro ac:schema-version="1" ac:macro-id="xyz" ac:name="code"><ac:parameter ac:name="language">rust</ac:parameter><ac:plain-text-body><![CDATA[fn main() {}]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```rust", "fn main() {}"},
		},
		{
			name:     "empty code block",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "code with HTML special characters",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">html</ac:parameter><ac:plain-text-body><![CDATA[<div>a < b && c > d</div>]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```html", "<div>", "a < b", "c > d", "</div>"},
		},
		{
			name:     "code with unicode",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[fmt.Println("こんにちは 🚀")]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", "こんにちは", "🚀"},
		},
		{
			name:     "multiple parameters with language",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="title">Example</ac:parameter><ac:parameter ac:name="language">python</ac:parameter><ac:parameter ac:name="collapse">true</ac:parameter><ac:plain-text-body><![CDATA[print("hello")]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```python", `print("hello")`},
			excludes: []string{"CDATA", "title", "collapse"},
		},
		{
			name:     "language parameter after other parameters",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="title">Test</ac:parameter><ac:parameter ac:name="linenumbers">true</ac:parameter><ac:parameter ac:name="language">java</ac:parameter><ac:plain-text-body><![CDATA[System.out.println();]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```java", "System.out.println();"},
		},
		{
			name: "newlines between elements",
			input: `<ac:structured-macro ac:name="code">
  <ac:parameter ac:name="language">go</ac:parameter>
  <ac:plain-text-body><![CDATA[func main() {}]]></ac:plain-text-body>
</ac:structured-macro>`,
			contains: []string{"```go", "func main() {}"},
		},
		{
			name:     "multiline code content",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[func main() {` + "\n" + `    fmt.Println("line 1")` + "\n" + `    fmt.Println("line 2")` + "\n" + `}]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", "func main() {", `fmt.Println("line 1")`, `fmt.Println("line 2")`, "}"},
		},
		{
			name: "multiple code blocks",
			input: `<p>First:</p>
<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[fmt.Println("one")]]></ac:plain-text-body></ac:structured-macro>
<p>Second:</p>
<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[print("two")]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", `fmt.Println("one")`, "```python", `print("two")`, "First:", "Second:"},
		},
		{
			name: "code block mixed with other macros",
			input: `<ac:structured-macro ac:name="info"><ac:rich-text-body><p>Info</p></ac:rich-text-body></ac:structured-macro>
<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[code here]]></ac:plain-text-body></ac:structured-macro>
<ac:structured-macro ac:name="warning"><ac:rich-text-body><p>Warning</p></ac:rich-text-body></ac:structured-macro>`,
			contains: []string{"```go", "code here", "Info", "Warning"},
		},
		{
			name:     "code with array brackets",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">js</ac:parameter><ac:plain-text-body><![CDATA[if (arr[i] > 5) { console.log(arr[j]); }]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```js", "arr[i]", "arr[j]"},
		},
		{
			name:     "whitespace only code",
			input:    `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[` + "\n   " + `]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "code with triple backticks",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">md</ac:parameter><ac:plain-text-body><![CDATA[` + "```python\nprint('nested')\n```" + `]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"print('nested')"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "code with backslashes",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[path := "C:\\Users\\test"` + "\n" + `regex := "\\d+\\.\\d+"]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", `C:\\Users\\test`, `\\d+\\.\\d+`},
		},
		{
			name:     "empty language value",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language"></ac:parameter><ac:plain-text-body><![CDATA[code here]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```", "code here"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "language with whitespace",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">  go  </ac:parameter><ac:plain-text-body><![CDATA[test]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"go", "test"},
			excludes: []string{"CDATA"},
		},
		{
			name:     "code with ampersands",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[if a && b || c & d {}]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", "a && b", "c & d"},
		},
		{
			name:     "code with quotes",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[s := "hello \"world\"" + 'c']]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```go", `"hello \"world\""`, "'c'"},
		},
		{
			name:     "very long single line",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">txt</ac:parameter><ac:plain-text-body><![CDATA[` + strings.Repeat("abcdefghij", 100) + `]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```txt", strings.Repeat("abcdefghij", 100)},
		},
		{
			name: "real confluence format with all attributes",
			input: `<ac:structured-macro ac:name="code" ac:schema-version="1" ac:macro-id="550e8400-e29b-41d4-a716-446655440000">
  <ac:parameter ac:name="language">go</ac:parameter>
  <ac:parameter ac:name="title">Example Code</ac:parameter>
  <ac:plain-text-body><![CDATA[package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
]]></ac:plain-text-body>
</ac:structured-macro>`,
			contains: []string{"```go", "package main", `import "fmt"`, "func main()", `fmt.Println("Hello, World!")`},
			excludes: []string{"CDATA", "Example Code", "macro-id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StorageToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("StorageToMarkdown() error = %v", err)
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing expected content %q\nGot:\n%s", want, result)
				}
			}

			for _, exclude := range tt.excludes {
				if strings.Contains(result, exclude) {
					t.Errorf("output contains unexpected content %q\nGot:\n%s", exclude, result)
				}
			}
		})
	}
}

func TestStorageToMarkdown_IntraWordUnderscores(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "simple intra-word underscore",
			input:    "<p>my_variable_name</p>",
			contains: []string{"my_variable_name"},
			excludes: []string{`\_`},
		},
		{
			name:     "multiple underscores in identifier",
			input:    "<p>user_account_settings_config</p>",
			contains: []string{"user_account_settings_config"},
			excludes: []string{`\_`},
		},
		{
			name:     "underscore in table cell",
			input:    "<table><tbody><tr><td>database_pool</td><td>config</td></tr></tbody></table>",
			contains: []string{"database_pool"},
			excludes: []string{`\_`},
		},
		{
			name:     "mixed content with underscores",
			input:    "<p>The module name is api_handler in the services directory.</p>",
			contains: []string{"api_handler"},
			excludes: []string{`\_`},
		},
		{
			name:     "underscore at word boundary preserved",
			input:    "<p><em>italic</em> and word_</p>",
			contains: []string{"*italic*", "word_"},
		},
		{
			name:     "code span with underscores not affected",
			input:    "<p>Use <code>my_variable_name</code> for this</p>",
			contains: []string{"`my_variable_name`"},
		},
		{
			name:     "consecutive underscores in identifier",
			input:    "<p>a_b_c_d_e</p>",
			contains: []string{"a_b_c_d_e"},
			excludes: []string{`\_`},
		},
		{
			name:     "underscores with numbers",
			input:    "<p>config_v2_test_123</p>",
			contains: []string{"config_v2_test_123"},
			excludes: []string{`\_`},
		},
		{
			name:     "snake_case in heading",
			input:    "<h2>The database_connection Module</h2>",
			contains: []string{"## The database_connection Module"},
			excludes: []string{`\_`},
		},
		{
			name:     "snake_case in link text",
			input:    `<p><a href="https://example.com">my_function_name</a></p>`,
			contains: []string{"[my_function_name]"},
			excludes: []string{`\_`},
		},
		{
			name:     "snake_case in list item with newline",
			input:    "<ul>\n<li>my_variable_name\n</li>\n<li>user_account_settings\n</li>\n</ul>",
			contains: []string{"my_variable_name", "user_account_settings"},
			excludes: []string{`\_`},
		},
		{
			name: "snake_case after multiple code blocks",
			input: `<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[func first() {}]]></ac:plain-text-body></ac:structured-macro>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[def second(): pass]]></ac:plain-text-body></ac:structured-macro>
<h2>Snake Case</h2>
<ul>
<li>my_variable_name
</li>
</ul>`,
			contains: []string{"```go", "```python", "my_variable_name"},
			excludes: []string{`my\_variable`},
		},
		{
			name: "snake_case after many code blocks with various content",
			input: `<h2>Code 1</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[package main
func first() {}
]]></ac:plain-text-body></ac:structured-macro>
<h2>Code 2</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[def second():
    pass
]]></ac:plain-text-body></ac:structured-macro>
<h2>Code 3</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">javascript</ac:parameter><ac:plain-text-body><![CDATA[const third = () => {};
]]></ac:plain-text-body></ac:structured-macro>
<h2>Code 4</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">bash</ac:parameter><ac:plain-text-body><![CDATA[#!/bin/bash
echo "test"
]]></ac:plain-text-body></ac:structured-macro>
<h2>Code 5</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[func fifth() {}
]]></ac:plain-text-body></ac:structured-macro>
<h2>Snake Case Section</h2>
<p>Text with my_variable_name here.</p>
<ul>
<li>api_handler
</li>
<li>config_manager
</li>
</ul>`,
			contains: []string{"```go", "```python", "my_variable_name", "api_handler", "config_manager"},
			excludes: []string{`\_`},
		},
		{
			name: "snake_case after code blocks including empty ones",
			input: `<h2>Normal Code</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[func test() {}
]]></ac:plain-text-body></ac:structured-macro>
<h2>Empty Code Block</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">none</ac:parameter></ac:structured-macro>
<h2>Another Empty</h2>
<ac:structured-macro ac:name="code" ac:schema-version="1"><ac:parameter ac:name="language">none</ac:parameter></ac:structured-macro>
<h2>Snake Case</h2>
<ul>
<li>my_variable_name
</li>
</ul>`,
			contains: []string{"my_variable_name"},
			excludes: []string{`\_`},
		},
		{
			name: "snake_case after code block containing nested backticks",
			input: "<h2>Code with nested backticks</h2>\n" +
				"<ac:structured-macro ac:name=\"code\" ac:schema-version=\"1\"><ac:parameter ac:name=\"language\">md</ac:parameter><ac:plain-text-body><![CDATA[```python\nprint('nested')\n```\n]]></ac:plain-text-body></ac:structured-macro>\n" +
				"<h2>More code</h2>\n" +
				"<ac:structured-macro ac:name=\"code\" ac:schema-version=\"1\"><ac:parameter ac:name=\"language\">go</ac:parameter><ac:plain-text-body><![CDATA[func test() {}\n]]></ac:plain-text-body></ac:structured-macro>\n" +
				"<h2>Snake Case</h2>\n" +
				"<ul>\n<li>my_variable_name\n</li>\n</ul>",
			contains: []string{"my_variable_name"},
			excludes: []string{`my\_variable`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StorageToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("StorageToMarkdown() error = %v", err)
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing expected content %q\nGot:\n%s", want, result)
				}
			}

			for _, exclude := range tt.excludes {
				if strings.Contains(result, exclude) {
					t.Errorf("output contains unexpected escaped underscore %q\nGot:\n%s", exclude, result)
				}
			}
		})
	}
}

func TestStorageToMarkdown_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		wantErr  bool
	}{
		{
			name:     "unclosed heading tag",
			input:    "<h1>Unclosed heading",
			contains: []string{"Unclosed heading"},
			wantErr:  false,
		},
		{
			name:     "unclosed paragraph",
			input:    "<p>Paragraph without closing",
			contains: []string{"Paragraph without closing"},
			wantErr:  false,
		},
		{
			name:     "mismatched tags",
			input:    "<h1>Title</h2>",
			contains: []string{"Title"},
			wantErr:  false,
		},
		{
			name:     "empty CDATA in code block",
			input:    `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[]]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"```"},
			wantErr:  false,
		},
		{
			name:     "deeply nested lists (10 levels)",
			input:    "<ul><li>L1<ul><li>L2<ul><li>L3<ul><li>L4<ul><li>L5<ul><li>L6<ul><li>L7<ul><li>L8<ul><li>L9<ul><li>L10</li></ul></li></ul></li></ul></li></ul></li></ul></li></ul></li></ul></li></ul></li></ul></li></ul>",
			contains: []string{"L1", "L10"},
			wantErr:  false,
		},
		{
			name:     "mixed valid and invalid HTML",
			input:    "<p>Valid paragraph</p><invalid>Unknown tag</invalid><p>Another valid</p>",
			contains: []string{"Valid paragraph", "Another valid"},
			wantErr:  false,
		},
		{
			name:     "HTML comments",
			input:    "<p>Before</p><!-- This is a comment --><p>After</p>",
			contains: []string{"Before", "After"},
			wantErr:  false,
		},
		{
			name:     "self-closing tags",
			input:    "<p>Line one<br/>Line two</p>",
			contains: []string{"Line one", "Line two"},
			wantErr:  false,
		},
		{
			name:     "unicode content",
			input:    "<p>日本語テキスト with emoji 🚀🎉 and symbols ™©®</p>",
			contains: []string{"日本語テキスト", "🚀🎉", "™©®"},
			wantErr:  false,
		},
		{
			name:     "whitespace only content",
			input:    "<p>   \n\t   </p>",
			contains: []string{},
			wantErr:  false,
		},
		{
			name:     "nested formatting deeply",
			input:    "<p><strong><em><code>deeply nested</code></em></strong></p>",
			contains: []string{"deeply nested"},
			wantErr:  false,
		},
		{
			name:     "table with empty cells",
			input:    "<table><thead><tr><th>A</th><th></th><th>C</th></tr></thead><tbody><tr><td></td><td>B</td><td></td></tr></tbody></table>",
			contains: []string{"| A |", "| C |", "| B |"},
			wantErr:  false,
		},
		{
			name:     "multiple sequential headings",
			input:    "<h1>One</h1><h2>Two</h2><h3>Three</h3><h4>Four</h4><h5>Five</h5><h6>Six</h6>",
			contains: []string{"# One", "## Two", "### Three", "#### Four", "##### Five", "###### Six"},
			wantErr:  false,
		},
		{
			name:     "special XML characters in attributes",
			input:    `<a href="https://example.com?a=1&amp;b=2">Link</a>`,
			contains: []string{"[Link]", "example.com"},
			wantErr:  false,
		},
		{
			name:     "script tags stripped",
			input:    "<p>Before</p><script>alert('xss')</script><p>After</p>",
			contains: []string{"Before", "After"},
			wantErr:  false,
		},
		{
			name:     "style tags stripped",
			input:    "<p>Content</p><style>body { color: red; }</style>",
			contains: []string{"Content"},
			wantErr:  false,
		},
		{
			name:     "CDATA with special characters",
			input:    `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[<>&"']]></ac:plain-text-body></ac:structured-macro>`,
			contains: []string{"<>&"},
			wantErr:  false,
		},
		{
			name:     "empty document",
			input:    "",
			contains: []string{},
			wantErr:  false,
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n   ",
			contains: []string{},
			wantErr:  false,
		},
		{
			name:     "plain text without HTML",
			input:    "Just plain text without any HTML tags",
			contains: []string{"Just plain text without any HTML tags"},
			wantErr:  false,
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
					t.Errorf("StorageToMarkdown() missing %q\nGot: %q", want, result)
				}
			}
		})
	}
}

func TestRoundTrip_ComprehensiveFile(t *testing.T) {
	// Read the comprehensive test markdown file
	mdContent, err := os.ReadFile("../../testdata/comprehensive-test.md")
	if err != nil {
		t.Skipf("Skipping: cannot read comprehensive-test.md: %v", err)
	}

	// Convert Markdown -> Storage
	storage := MarkdownToStorage(string(mdContent))

	// Convert Storage -> Markdown
	result, err := StorageToMarkdown(storage)
	if err != nil {
		t.Fatalf("StorageToMarkdown() error = %v", err)
	}

	// Check that snake_case identifiers are not escaped
	snakeCaseTests := []string{
		"my_variable_name",
		"user_account_settings",
		"database_connection_pool",
		"api_handler",
		"config_manager",
		"data_processor",
	}

	for _, identifier := range snakeCaseTests {
		escaped := strings.ReplaceAll(identifier, "_", `\_`)
		if strings.Contains(result, escaped) {
			t.Errorf("Found escaped underscore in %q - should be %q", escaped, identifier)
		}
		if !strings.Contains(result, identifier) {
			t.Errorf("Missing identifier %q in output", identifier)
		}
	}
}

// Benchmark tests for performance tracking

var benchmarkMarkdown = `# Benchmark Document

This is a paragraph with **bold**, *italic*, and ` + "`code`" + ` formatting.

## Section One

- Item one with some text
- Item two with more text
- Item three with even more text

### Subsection

1. First ordered item
2. Second ordered item
3. Third ordered item

## Code Example

Here is some inline ` + "`code`" + ` and a code block:

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## Table

| Column A | Column B | Column C |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
| Value 4  | Value 5  | Value 6  |

## Links and Images

Visit [Google](https://google.com) for more information.

> This is a blockquote with some important information.

The end.
`

var benchmarkStorage = `<h1>Benchmark Document</h1>
<p>This is a paragraph with <strong>bold</strong>, <em>italic</em>, and <code>code</code> formatting.</p>
<h2>Section One</h2>
<ul>
<li>Item one with some text</li>
<li>Item two with more text</li>
<li>Item three with even more text</li>
</ul>
<h3>Subsection</h3>
<ol>
<li>First ordered item</li>
<li>Second ordered item</li>
<li>Third ordered item</li>
</ol>
<h2>Code Example</h2>
<p>Here is some inline <code>code</code> and a code block:</p>
<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
]]></ac:plain-text-body></ac:structured-macro>
<h2>Table</h2>
<table><thead><tr><th>Column A</th><th>Column B</th><th>Column C</th></tr></thead><tbody><tr><td>Value 1</td><td>Value 2</td><td>Value 3</td></tr><tr><td>Value 4</td><td>Value 5</td><td>Value 6</td></tr></tbody></table>
<h2>Links and Images</h2>
<p>Visit <a href="https://google.com">Google</a> for more information.</p>
<blockquote><p>This is a blockquote with some important information.</p></blockquote>
<p>The end.</p>`

func BenchmarkMarkdownToStorage(b *testing.B) {
	for b.Loop() {
		MarkdownToStorage(benchmarkMarkdown)
	}
}

func BenchmarkStorageToMarkdown(b *testing.B) {
	for b.Loop() {
		StorageToMarkdown(benchmarkStorage)
	}
}

func BenchmarkMarkdownToStorage_Large(b *testing.B) {
	// Create a larger document by repeating the benchmark content
	large := strings.Repeat(benchmarkMarkdown, 10)
	b.ResetTimer()
	for b.Loop() {
		MarkdownToStorage(large)
	}
}

func BenchmarkStorageToMarkdown_Large(b *testing.B) {
	// Create a larger document by repeating the benchmark content
	large := strings.Repeat(benchmarkStorage, 10)
	b.ResetTimer()
	for b.Loop() {
		StorageToMarkdown(large)
	}
}
