package converter

import (
	"html"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

var storageConverter = converter.NewConverter(
	converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
		strikethrough.NewStrikethroughPlugin(),
	),
)

// codeMacroRegex matches Confluence code macro WITH content
// Uses \s* for explicit whitespace handling between elements
var codeMacroRegex = regexp.MustCompile(
	`<ac:structured-macro[^>]*ac:name="code"[^>]*>\s*` +
		`((?:<ac:parameter[^>]*>[^<]*</ac:parameter>\s*)*)` + // capture parameters with trailing whitespace
		`<ac:plain-text-body><!\[CDATA\[([\s\S]*?)\]\]></ac:plain-text-body>\s*` +
		`</ac:structured-macro>`)

// emptyCodeMacroRegex matches Confluence code macro WITHOUT content (empty code block)
var emptyCodeMacroRegex = regexp.MustCompile(
	`<ac:structured-macro[^>]*ac:name="code"[^>]*>\s*` +
		`((?:<ac:parameter[^>]*>[^<]*</ac:parameter>\s*)*)` + // capture parameters with trailing whitespace
		`</ac:structured-macro>`)

// languageRegex extracts language value from parameters
var languageRegex = regexp.MustCompile(
	`<ac:parameter[^>]*ac:name="language"[^>]*>([^<]*)</ac:parameter>`)

// taskListRegex matches Confluence task list macro
var taskListRegex = regexp.MustCompile(`<ac:task-list>([\s\S]*?)</ac:task-list>`)

// taskRegex matches individual task items
var taskRegex = regexp.MustCompile(
	`<ac:task>\s*<ac:task-status>([^<]*)</ac:task-status>\s*<ac:task-body>([\s\S]*?)</ac:task-body>\s*</ac:task>`)

// imageRegex matches Confluence image macro with external URL
var imageRegex = regexp.MustCompile(
	`<ac:image[^>]*>\s*<ri:url\s+ri:value="([^"]*)"[^/]*/>\s*</ac:image>`)

func StorageToMarkdown(storage string) (string, error) {
	// Pre-process: convert Confluence code macros WITH content to standard HTML pre/code blocks
	processed := codeMacroRegex.ReplaceAllStringFunc(storage, func(match string) string {
		submatches := codeMacroRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		params := submatches[1]
		code := submatches[2]

		// Extract language from parameters
		var language string
		if langMatch := languageRegex.FindStringSubmatch(params); len(langMatch) >= 2 {
			language = strings.TrimSpace(langMatch[1])
		}

		// Escape HTML entities in code content (< and > must be escaped for HTML parsing)
		code = strings.ReplaceAll(code, "<", "&lt;")
		code = strings.ReplaceAll(code, ">", "&gt;")

		// Build pre/code with optional language class
		if language != "" {
			return `<pre><code class="language-` + language + `">` + code + `</code></pre>`
		}
		return `<pre><code>` + code + `</code></pre>`
	})

	// Pre-process: convert empty code macros (no content) to empty code blocks
	processed = emptyCodeMacroRegex.ReplaceAllStringFunc(processed, func(match string) string {
		submatches := emptyCodeMacroRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		params := submatches[1]

		// Extract language from parameters
		var language string
		if langMatch := languageRegex.FindStringSubmatch(params); len(langMatch) >= 2 {
			language = strings.TrimSpace(langMatch[1])
		}

		// Build empty pre/code with optional language class
		if language != "" {
			return `<pre><code class="language-` + language + `"></code></pre>`
		}
		return `<pre><code></code></pre>`
	})

	// Pre-process: convert Confluence task lists to HTML checkboxes
	processed = taskListRegex.ReplaceAllStringFunc(processed, func(match string) string {
		// Extract task list content
		listMatch := taskListRegex.FindStringSubmatch(match)
		if len(listMatch) < 2 {
			return match
		}
		taskListContent := listMatch[1]

		// Convert each task to a list item with checkbox
		var result strings.Builder
		result.WriteString("<ul>\n")

		tasks := taskRegex.FindAllStringSubmatch(taskListContent, -1)
		for _, task := range tasks {
			if len(task) < 3 {
				continue
			}
			status := strings.TrimSpace(task[1])
			body := strings.TrimSpace(task[2])

			// Remove any paragraph tags from body
			body = strings.TrimPrefix(body, "<p>")
			body = strings.TrimSuffix(body, "</p>")
			body = strings.TrimSpace(body)

			if status == "complete" {
				result.WriteString("<li>[x] " + body + "</li>\n")
			} else {
				result.WriteString("<li>[ ] " + body + "</li>\n")
			}
		}
		result.WriteString("</ul>")
		return result.String()
	})

	// Pre-process: convert Confluence images to standard HTML img tags
	processed = imageRegex.ReplaceAllStringFunc(processed, func(match string) string {
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		url := submatches[1]
		// Use empty alt text since Confluence doesn't store it
		return `<img src="` + url + `" alt="" />`
	})

	markdown, err := storageConverter.ConvertString(processed)
	if err != nil {
		return "", err
	}

	// Decode HTML entities (e.g., &lt; → <, &gt; → >, &amp; → &)
	markdown = html.UnescapeString(markdown)

	// Fix over-escaped task list checkboxes: \[ ] -> [ ] and \[x] -> [x]
	markdown = strings.ReplaceAll(markdown, `\[ ]`, `[ ]`)
	markdown = strings.ReplaceAll(markdown, `\[x]`, `[x]`)
	markdown = strings.ReplaceAll(markdown, `\[X]`, `[x]`)

	// Fix over-escaped markdown characters from html-to-markdown library
	// Pattern 1: \\\X -> \X (triple backslash: both backslash and special char were escaped)
	// Pattern 2: \\X -> \X (double backslash: only backslash was escaped, char is literal)
	markdown = fixOverEscaping(markdown)

	// Fix intra-word underscores globally (safe even in code blocks since pattern is specific)
	// The pattern alphanumeric\_alphanumeric never needs escaping in Markdown
	for intraWordUnderscoreRegex.MatchString(markdown) {
		markdown = intraWordUnderscoreRegex.ReplaceAllString(markdown, "${1}_${2}")
	}

	// Fix extra blank lines in nested lists
	// The html-to-markdown library creates "loose" lists with blank lines before nested items
	markdown = fixNestedListSpacing(markdown)

	return markdown, nil
}

// nestedListBlankLineRegex matches blank lines before nested list items
// Pattern: newline, spaces, newline, spaces, list marker (- or digit.)
var nestedListBlankLineRegex = regexp.MustCompile(`(\n)([ \t]+)\n([ \t]+[-*]|[ \t]+\d+\.)`)

// fixNestedListSpacing removes extra blank lines before nested list items
func fixNestedListSpacing(markdown string) string {
	// Replace: \n<indent>\n<indent>- with \n<indent>-
	// This converts "loose" nested lists to "tight" lists
	return nestedListBlankLineRegex.ReplaceAllString(markdown, "$1$3")
}

// codeBlockRegex matches fenced code blocks to protect their content
// Matches ``` followed by optional language, newline, content, and closing ```
// The closing ``` may be prefixed with > for blockquote code blocks
var codeBlockRegex = regexp.MustCompile("(?s)```[a-zA-Z]*\\n.*?\\n>? ?```")

// fixOverEscaping removes redundant backslash escapes added by html-to-markdown
// but preserves content inside code blocks
func fixOverEscaping(markdown string) string {
	// Find all code blocks and their positions
	matches := codeBlockRegex.FindAllStringIndex(markdown, -1)

	// If no code blocks, process the entire string
	if len(matches) == 0 {
		return fixEscapesInText(markdown)
	}

	// Process text between code blocks, preserving code block content
	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Process text before this code block
		if start > lastEnd {
			result.WriteString(fixEscapesInText(markdown[lastEnd:start]))
		}

		// Preserve code block content unchanged
		result.WriteString(markdown[start:end])
		lastEnd = end
	}

	// Process any remaining text after the last code block
	if lastEnd < len(markdown) {
		result.WriteString(fixEscapesInText(markdown[lastEnd:]))
	}

	return result.String()
}

// intraWordUnderscoreRegex matches escaped underscores between alphanumeric chars
// These never create emphasis and don't need escaping
var intraWordUnderscoreRegex = regexp.MustCompile(`([a-zA-Z0-9])\\_([a-zA-Z0-9])`)

// fixEscapesInText removes redundant backslash escapes from non-code text
func fixEscapesInText(text string) string {
	// These chars get triple-escaped when preceded by backslash in source
	// (the library escapes both the backslash AND the special char)
	// Pipe is included because in tables: \| → \\\ + \| = \\\|
	tripleEscapeChars := "*_`[(|"

	// These chars get double-escaped (only the backslash is escaped)
	// Note: some chars like * can appear with either double or triple escaping
	// depending on position, so we include them in both lists
	doubleEscapeChars := "*_`[]#-+.!|()}<>"

	result := text

	// First pass: fix triple escapes -> single escape
	for _, char := range tripleEscapeChars {
		pattern := `\\\` + string(char)
		replacement := `\` + string(char)
		result = strings.ReplaceAll(result, pattern, replacement)
	}

	// Second pass: fix double escapes -> single escape
	for _, char := range doubleEscapeChars {
		pattern := `\\` + string(char)
		replacement := `\` + string(char)
		result = strings.ReplaceAll(result, pattern, replacement)
	}

	return result
}
