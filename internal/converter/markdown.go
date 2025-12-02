package converter

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Compile regex patterns once at package initialization
	headerPattern1     = regexp.MustCompile(`(?m)^# (.+)$`)
	headerPattern2     = regexp.MustCompile(`(?m)^## (.+)$`)
	headerPattern3     = regexp.MustCompile(`(?m)^### (.+)$`)
	headerPattern4     = regexp.MustCompile(`(?m)^#### (.+)$`)
	headerPattern5     = regexp.MustCompile(`(?m)^##### (.+)$`)
	headerPattern6     = regexp.MustCompile(`(?m)^###### (.+)$`)
	boldPattern        = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	italicPattern      = regexp.MustCompile(`\*([^*\n]+)\*`)
	codePattern        = regexp.MustCompile("`([^`]+)`")
	codeBlockPattern   = regexp.MustCompile("```([a-z]*)\n([\\s\\S]*?)```")
	linkPattern        = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	orderedListPattern = regexp.MustCompile(`^\d+\.\s+`)
	tablePattern       = regexp.MustCompile(`(?m)^\|(.+)\|\s*\n\|[\s:|-]+\|\s*\n((?:\|.+\|\s*\n?)*)`)
)

func MarkdownToStorage(markdown string) string {
	// Use the goldmark-based converter for robust markdown parsing
	return MarkdownToStorageGoldmark(markdown)
}

// MarkdownToStorageRegex is the old regex-based converter (kept for reference)
func MarkdownToStorageRegex(markdown string) string {
	content := markdown

	// Step 1: Extract code blocks and inline code first to protect from HTML escaping
	codePlaceholders := make(map[string]string)
	content = extractCodeBlocks(content, codePlaceholders)
	content = extractInlineCode(content, codePlaceholders)

	// Step 2: Escape HTML entities in remaining content to prevent XSS
	content = escapeHTML(content)

	// Step 3: Convert markdown to HTML
	content = convertTables(content)
	content = convertHeaders(content)
	content = convertBold(content)
	content = convertItalic(content)
	content = convertLinks(content)
	content = convertUnorderedLists(content)
	content = convertOrderedLists(content)
	content = convertBlockquotes(content)

	// Step 4: Restore code blocks and inline code as Confluence format
	content = restorePlaceholders(content, codePlaceholders)

	// Step 5: Handle line breaks
	content = convertLineBreaks(content)

	return content
}

// extractCodeBlocks extracts code blocks and replaces them with placeholders
func extractCodeBlocks(content string, placeholders map[string]string) string {
	counter := 0
	return codeBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := codeBlockPattern.FindStringSubmatch(match)
		lang := parts[1]
		code := parts[2]
		placeholder := fmt.Sprintf("___CODEBLOCK_%d___", counter)
		if lang == "" {
			lang = "none"
		}
		// Store as Confluence code macro with newlines for proper formatting
		placeholders[placeholder] = `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">` + lang + `</ac:parameter><ac:plain-text-body><![CDATA[` + code + `]]></ac:plain-text-body></ac:structured-macro>` + "\n"
		counter++
		return "\n" + placeholder + "\n"
	})
}

// extractInlineCode extracts inline code and replaces them with placeholders
func extractInlineCode(content string, placeholders map[string]string) string {
	counter := 0
	return codePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := codePattern.FindStringSubmatch(match)
		code := parts[1]
		placeholder := fmt.Sprintf("___CODE_%d___", counter)
		placeholders[placeholder] = "<code>" + code + "</code>"
		counter++
		return placeholder
	})
}

// escapeHTML escapes HTML entities in the content
func escapeHTML(content string) string {
	content = strings.ReplaceAll(content, "&", "&amp;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	return content
}

// restorePlaceholders restores the code blocks and inline code
func restorePlaceholders(content string, placeholders map[string]string) string {
	for placeholder, replacement := range placeholders {
		content = strings.ReplaceAll(content, placeholder, replacement)
	}
	return content
}

func convertHeaders(content string) string {
	content = headerPattern6.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern6.FindStringSubmatch(match)
		return "<h6>" + parts[1] + "</h6>"
	})
	content = headerPattern5.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern5.FindStringSubmatch(match)
		return "<h5>" + parts[1] + "</h5>"
	})
	content = headerPattern4.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern4.FindStringSubmatch(match)
		return "<h4>" + parts[1] + "</h4>"
	})
	content = headerPattern3.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern3.FindStringSubmatch(match)
		return "<h3>" + parts[1] + "</h3>"
	})
	content = headerPattern2.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern2.FindStringSubmatch(match)
		return "<h2>" + parts[1] + "</h2>"
	})
	content = headerPattern1.ReplaceAllStringFunc(content, func(match string) string {
		parts := headerPattern1.FindStringSubmatch(match)
		return "<h1>" + parts[1] + "</h1>"
	})
	return content
}

func convertBold(content string) string {
	return boldPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := boldPattern.FindStringSubmatch(match)
		return "<strong>" + parts[1] + "</strong>"
	})
}

func convertItalic(content string) string {
	return italicPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := italicPattern.FindStringSubmatch(match)
		return "<em>" + parts[1] + "</em>"
	})
}

func convertTables(content string) string {
	return tablePattern.ReplaceAllStringFunc(content, func(match string) string {
		lines := strings.Split(strings.TrimSpace(match), "\n")
		if len(lines) < 3 {
			return match
		}

		var result strings.Builder
		result.WriteString("<table><tbody>\n")

		// Process header row
		headerCells := strings.Split(strings.Trim(lines[0], "|"), "|")
		result.WriteString("<tr>")
		for _, cell := range headerCells {
			result.WriteString("<th>")
			result.WriteString(strings.TrimSpace(cell))
			result.WriteString("</th>")
		}
		result.WriteString("</tr>\n")

		// Skip separator line (lines[1])
		// Process data rows
		for i := 2; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			dataCells := strings.Split(strings.Trim(line, "|"), "|")
			result.WriteString("<tr>")
			for _, cell := range dataCells {
				result.WriteString("<td>")
				result.WriteString(strings.TrimSpace(cell))
				result.WriteString("</td>")
			}
			result.WriteString("</tr>\n")
		}

		result.WriteString("</tbody></table>\n")
		return result.String()
	})
}

func convertLinks(content string) string {
	return linkPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := linkPattern.FindStringSubmatch(match)
		text := parts[1]
		href := parts[2]
		return `<a href="` + href + `">` + text + `</a>`
	})
}

func convertUnorderedLists(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.Grow(len(content) + 100) // Pre-allocate with buffer for tags
	inList := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "- ") {
			if !inList {
				result.WriteString("<ul>\n")
				inList = true
			}
			text := strings.TrimSpace(trimmed[2:])
			result.WriteString("<li>")
			result.WriteString(text)
			result.WriteString("</li>\n")
		} else {
			if inList {
				result.WriteString("</ul>\n")
				inList = false
			}
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}

	if inList {
		result.WriteString("</ul>")
	}

	return result.String()
}

func convertOrderedLists(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.Grow(len(content) + 100) // Pre-allocate with buffer for tags
	inList := false

	for i, line := range lines {
		if orderedListPattern.MatchString(strings.TrimSpace(line)) {
			if !inList {
				result.WriteString("<ol>\n")
				inList = true
			}
			text := orderedListPattern.ReplaceAllString(strings.TrimSpace(line), "")
			result.WriteString("<li>")
			result.WriteString(text)
			result.WriteString("</li>\n")
		} else {
			if inList {
				result.WriteString("</ol>\n")
				inList = false
			}
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}

	if inList {
		result.WriteString("</ol>")
	}

	return result.String()
}

func convertBlockquotes(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.Grow(len(content) + 100) // Pre-allocate with buffer for tags
	inQuote := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match escaped > which becomes &gt; after html.EscapeString
		if strings.HasPrefix(trimmed, "&gt; ") {
			if !inQuote {
				result.WriteString("<blockquote>\n")
				inQuote = true
			}
			text := strings.TrimSpace(trimmed[5:]) // Skip "&gt; "
			result.WriteString("<p>")
			result.WriteString(text)
			result.WriteString("</p>\n")
		} else {
			if inQuote {
				result.WriteString("</blockquote>\n")
				inQuote = false
			}
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}

	if inQuote {
		result.WriteString("</blockquote>")
	}

	return result.String()
}

func convertLineBreaks(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	result.Grow(len(content) + 100) // Pre-allocate with buffer for tags
	inBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if line is HTML tag or placeholder
		isBlockElement := strings.HasPrefix(trimmed, "<") || strings.HasPrefix(trimmed, "___")

		if isBlockElement {
			inBlock = true
		}

		if !inBlock && trimmed != "" && !isBlockElement {
			result.WriteString("<p>")
			result.WriteString(line)
			result.WriteString("</p>")
		} else {
			result.WriteString(line)
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}

		if (strings.HasSuffix(trimmed, ">") || strings.HasSuffix(trimmed, "___")) && inBlock {
			inBlock = false
		}
	}

	return result.String()
}
