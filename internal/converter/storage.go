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

// codeMacroRegex matches Confluence code macro and captures parameters + content
var codeMacroRegex = regexp.MustCompile(
	`<ac:structured-macro[^>]*ac:name="code"[^>]*>` +
		`([\s\S]*?)` + // capture all parameters (group 1)
		`<ac:plain-text-body><!\[CDATA\[([\s\S]*?)\]\]></ac:plain-text-body>` +
		`[\s\S]*?</ac:structured-macro>`)

// languageRegex extracts language value from parameters
var languageRegex = regexp.MustCompile(
	`<ac:parameter[^>]*ac:name="language"[^>]*>([^<]*)</ac:parameter>`)

func StorageToMarkdown(storage string) (string, error) {
	// Pre-process: convert Confluence code macros to standard HTML pre/code blocks
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

	markdown, err := storageConverter.ConvertString(processed)
	if err != nil {
		return "", err
	}
	// Decode HTML entities (e.g., &lt; → <, &gt; → >, &amp; → &)
	return html.UnescapeString(markdown), nil
}
