package converter

import (
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ConfluenceRenderer is a renderer that outputs Confluence Storage Format (XHTML).
type ConfluenceRenderer struct{}

// NewConfluenceRenderer creates a new ConfluenceRenderer.
func NewConfluenceRenderer() renderer.NodeRenderer {
	return &ConfluenceRenderer{}
}

// RegisterFuncs registers node rendering functions.
func (r *ConfluenceRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Block elements
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// Inline elements
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

	// Table extension (GFM)
	reg.Register(extast.KindTable, r.renderTable)
	reg.Register(extast.KindTableHeader, r.renderTableHeader)
	reg.Register(extast.KindTableRow, r.renderTableRow)
	reg.Register(extast.KindTableCell, r.renderTableCell)

	// Task list extension (GFM)
	reg.Register(extast.KindTaskCheckBox, r.renderTaskCheckBox)

	// Strikethrough extension (GFM)
	reg.Register(extast.KindStrikethrough, r.renderStrikethrough)
}

// Helper to write lines from a node
func (r *ConfluenceRenderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		w.Write(line.Value(source))
	}
}

// isTaskList checks if a list contains task checkboxes
func isTaskList(node ast.Node) bool {
	// Check first list item for a task checkbox
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindListItem {
			// Look for TaskCheckBox in the list item's children
			for grandchild := child.FirstChild(); grandchild != nil; grandchild = grandchild.NextSibling() {
				// TaskCheckBox can be direct child or inside a paragraph/textblock
				if grandchild.Kind() == extast.KindTaskCheckBox {
					return true
				}
				// Check inside paragraph or textblock
				for greatgrandchild := grandchild.FirstChild(); greatgrandchild != nil; greatgrandchild = greatgrandchild.NextSibling() {
					if greatgrandchild.Kind() == extast.KindTaskCheckBox {
						return true
					}
				}
			}
		}
	}
	return false
}

// getTaskCheckBox finds the TaskCheckBox node in a list item
func getTaskCheckBox(listItem ast.Node) *extast.TaskCheckBox {
	for child := listItem.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == extast.KindTaskCheckBox {
			return child.(*extast.TaskCheckBox)
		}
		// Check inside paragraph or textblock
		for grandchild := child.FirstChild(); grandchild != nil; grandchild = grandchild.NextSibling() {
			if grandchild.Kind() == extast.KindTaskCheckBox {
				return grandchild.(*extast.TaskCheckBox)
			}
		}
	}
	return nil
}

// Document
func (r *ConfluenceRenderer) renderDocument(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Don't wrap document in any tags
	return ast.WalkContinue, nil
}

// Heading
func (r *ConfluenceRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		w.WriteString("<h")
		w.WriteByte("0123456"[n.Level])
		w.WriteByte('>')
	} else {
		w.WriteString("</h")
		w.WriteByte("0123456"[n.Level])
		w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

// Blockquote
func (r *ConfluenceRenderer) renderBlockquote(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<blockquote>\n")
	} else {
		w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

// CodeBlock (indented code)
func (r *ConfluenceRenderer) renderCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(`<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">none</ac:parameter><ac:plain-text-body><![CDATA[`)
		r.writeLines(w, source, node)
	} else {
		w.WriteString("]]></ac:plain-text-body></ac:structured-macro>\n")
	}
	return ast.WalkContinue, nil
}

// FencedCodeBlock
func (r *ConfluenceRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		lang := "none"
		if n.Language(source) != nil {
			lang = string(n.Language(source))
		}
		w.WriteString(`<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">`)
		w.WriteString(lang)
		w.WriteString(`</ac:parameter><ac:plain-text-body><![CDATA[`)
		r.writeLines(w, source, n)
	} else {
		w.WriteString("]]></ac:plain-text-body></ac:structured-macro>\n")
	}
	return ast.WalkContinue, nil
}

// HTMLBlock - skip raw HTML for security
func (r *ConfluenceRenderer) renderHTMLBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<!-- raw HTML omitted -->\n")
	}
	return ast.WalkContinue, nil
}

// List
func (r *ConfluenceRenderer) renderList(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)

	// Check if this is a task list
	if isTaskList(node) {
		if entering {
			w.WriteString("<ac:task-list>\n")
		} else {
			w.WriteString("</ac:task-list>\n")
		}
		return ast.WalkContinue, nil
	}

	// Regular list
	if entering {
		if n.IsOrdered() {
			w.WriteString("<ol>\n")
		} else {
			w.WriteString("<ul>\n")
		}
	} else {
		if n.IsOrdered() {
			w.WriteString("</ol>\n")
		} else {
			w.WriteString("</ul>\n")
		}
	}
	return ast.WalkContinue, nil
}

// ListItem
func (r *ConfluenceRenderer) renderListItem(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Check if parent is a task list
	parent := node.Parent()
	if parent != nil && isTaskList(parent) {
		checkbox := getTaskCheckBox(node)
		if entering {
			w.WriteString("<ac:task>\n")
			if checkbox != nil && checkbox.IsChecked {
				w.WriteString("<ac:task-status>complete</ac:task-status>\n")
			} else {
				w.WriteString("<ac:task-status>incomplete</ac:task-status>\n")
			}
			w.WriteString("<ac:task-body>")
		} else {
			w.WriteString("</ac:task-body>\n</ac:task>\n")
		}
		return ast.WalkContinue, nil
	}

	// Regular list item
	if entering {
		w.WriteString("<li>")
	} else {
		w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

// Paragraph
func (r *ConfluenceRenderer) renderParagraph(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Skip paragraph tags inside task list items (ac:task-body handles content directly)
	parent := node.Parent()
	if parent != nil && parent.Kind() == ast.KindListItem {
		grandparent := parent.Parent()
		if grandparent != nil && isTaskList(grandparent) {
			// Don't wrap task item content in <p> tags
			return ast.WalkContinue, nil
		}
	}

	if entering {
		w.WriteString("<p>")
	} else {
		w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

// TextBlock
func (r *ConfluenceRenderer) renderTextBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

// ThematicBreak (horizontal rule)
func (r *ConfluenceRenderer) renderThematicBreak(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<hr />\n")
	}
	return ast.WalkContinue, nil
}

// AutoLink
func (r *ConfluenceRenderer) renderAutoLink(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if entering {
		w.WriteString(`<a href="`)
		w.Write(util.EscapeHTML(util.URLEscape(n.URL(source), true)))
		w.WriteString(`">`)
		w.Write(util.EscapeHTML(n.Label(source)))
	} else {
		w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// CodeSpan (inline code)
func (r *ConfluenceRenderer) renderCodeSpan(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<code>")
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			w.Write(segment.Value(source))
		}
		w.WriteString("</code>")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// Emphasis (italic or bold)
func (r *ConfluenceRenderer) renderEmphasis(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	tag := "em"
	if n.Level == 2 {
		tag = "strong"
	}
	if entering {
		w.WriteByte('<')
		w.WriteString(tag)
		w.WriteByte('>')
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteByte('>')
	}
	return ast.WalkContinue, nil
}

// Image
func (r *ConfluenceRenderer) renderImage(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Image)
	if entering {
		w.WriteString(`<ac:image><ri:url ri:value="`)
		w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		w.WriteString(`" /></ac:image>`)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// Link
func (r *ConfluenceRenderer) renderLink(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		w.WriteString(`<a href="`)
		w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		w.WriteByte('"')
		// Add title attribute if present
		if len(n.Title) > 0 {
			w.WriteString(` title="`)
			w.Write(util.EscapeHTML(n.Title))
			w.WriteByte('"')
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// RawHTML - skip for security
func (r *ConfluenceRenderer) renderRawHTML(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Skip raw HTML
	return ast.WalkContinue, nil
}

// Text
func (r *ConfluenceRenderer) renderText(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Text)
		segment := n.Segment
		w.Write(util.EscapeHTML(segment.Value(source)))
		if n.HardLineBreak() {
			w.WriteString("<br />\n")
		} else if n.SoftLineBreak() {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

// String
func (r *ConfluenceRenderer) renderString(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.String)
		w.Write(util.EscapeHTML(n.Value))
	}
	return ast.WalkContinue, nil
}

// Table
func (r *ConfluenceRenderer) renderTable(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<table><tbody>\n")
	} else {
		w.WriteString("</tbody></table>\n")
	}
	return ast.WalkContinue, nil
}

// TableHeader
func (r *ConfluenceRenderer) renderTableHeader(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// TableHeader is just a container, don't output tags
	return ast.WalkContinue, nil
}

// TableRow
func (r *ConfluenceRenderer) renderTableRow(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<tr>")
	} else {
		w.WriteString("</tr>\n")
	}
	return ast.WalkContinue, nil
}

// TableCell
func (r *ConfluenceRenderer) renderTableCell(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*extast.TableCell)
	tag := "td"
	if n.Parent().Kind() == extast.KindTableHeader {
		tag = "th"
	}

	if entering {
		w.WriteByte('<')
		w.WriteString(tag)

		// Handle alignment
		if n.Alignment != extast.AlignNone {
			w.WriteString(` align="`)
			switch n.Alignment {
			case extast.AlignLeft:
				w.WriteString("left")
			case extast.AlignCenter:
				w.WriteString("center")
			case extast.AlignRight:
				w.WriteString("right")
			}
			w.WriteByte('"')
		}

		w.WriteByte('>')
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteByte('>')
	}
	return ast.WalkContinue, nil
}

// TaskCheckBox - rendered as part of task list item, skip here
func (r *ConfluenceRenderer) renderTaskCheckBox(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// TaskCheckBox is handled by the list/listItem rendering for Confluence task format
	// Skip rendering here to avoid duplicate output
	return ast.WalkContinue, nil
}

// Strikethrough
func (r *ConfluenceRenderer) renderStrikethrough(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<del>")
	} else {
		w.WriteString("</del>")
	}
	return ast.WalkContinue, nil
}
