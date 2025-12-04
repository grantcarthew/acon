package converter

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// MarkdownToStorage converts markdown to Confluence Storage Format using Goldmark.
func MarkdownToStorage(markdown string) string {
	// Create Goldmark parser with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown (includes tables)
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Add IDs to headings
		),
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(NewConfluenceRenderer(), 1000),
				),
			),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		// If conversion fails, return original markdown as fallback
		return markdown
	}

	return buf.String()
}
