package converter

import (
	"html"
	"regexp"
	"strings"
)

var (
	// Compile regex patterns once at package initialization
	headerPattern1    = regexp.MustCompile(`(?m)^# (.+)$`)
	headerPattern2    = regexp.MustCompile(`(?m)^## (.+)$`)
	headerPattern3    = regexp.MustCompile(`(?m)^### (.+)$`)
	headerPattern4    = regexp.MustCompile(`(?m)^#### (.+)$`)
	headerPattern5    = regexp.MustCompile(`(?m)^##### (.+)$`)
	headerPattern6    = regexp.MustCompile(`(?m)^###### (.+)$`)
	boldPattern       = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	italicPattern     = regexp.MustCompile(`\*([^*\n]+)\*`)
	codePattern       = regexp.MustCompile("`([^`]+)`")
	codeBlockPattern  = regexp.MustCompile("```([a-z]*)\n([\\s\\S]*?)```")
	linkPattern       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	orderedListPattern = regexp.MustCompile(`^\d+\.\s+`)
)

func MarkdownToStorage(markdown string) string {
	content := markdown

	// Escape all HTML entities in the input to prevent XSS
	// This converts user's < > & to &lt; &gt; &amp;
	// Our markdown conversions will then generate intentional HTML tags
	content = html.EscapeString(content)

	// Process code to protect content from other conversions
	content = convertCodeBlocks(content)
	content = convertCode(content)

	// Then formatting
	content = convertHeaders(content)
	content = convertBold(content)
	content = convertItalic(content)
	content = convertLinks(content)

	// Block-level elements (blockquotes use escaped &gt; now)
	content = convertUnorderedLists(content)
	content = convertOrderedLists(content)
	content = convertBlockquotes(content)

	// Line breaks last
	content = convertLineBreaks(content)

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

func convertCode(content string) string {
	return codePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := codePattern.FindStringSubmatch(match)
		return "<code>" + parts[1] + "</code>"
	})
}

func convertCodeBlocks(content string) string {
	return codeBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := codeBlockPattern.FindStringSubmatch(match)
		lang := parts[1]
		code := parts[2]
		if lang == "" {
			lang = "none"
		}
		return `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">` + lang + `</ac:parameter><ac:plain-text-body><![CDATA[` + code + `]]></ac:plain-text-body></ac:structured-macro>`
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
		if strings.HasPrefix(trimmed, "<") {
			inBlock = true
		}

		if !inBlock && trimmed != "" && !strings.HasPrefix(trimmed, "<") {
			result.WriteString("<p>")
			result.WriteString(line)
			result.WriteString("</p>")
		} else {
			result.WriteString(line)
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}

		if strings.HasSuffix(trimmed, ">") && inBlock {
			inBlock = false
		}
	}

	return result.String()
}
