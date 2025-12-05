# Test Data and Feature Gap Documentation

This directory contains test fixtures for validating acon's bidirectional Markdown ↔ Confluence conversion.

## Quick Start

```bash
# Run the automated round-trip test (recommended)
./testdata/roundtrip-test.sh PARENT_PAGE_ID

# Or manually create a test page
cat testdata/comprehensive-test.md | ./acon page create -t "Test Page" --parent PAGE_ID

# View it back to check round-trip conversion
./acon page view PAGE_ID
```

## Automated Testing

The `roundtrip-test.sh` script provides comprehensive round-trip testing:

```bash
# Run with a parent page ID
./testdata/roundtrip-test.sh 1857355975

# Features:
# - Creates a test page from comprehensive-test.md
# - Retrieves and validates 28 feature checks
# - Opens the page in your browser for visual review
# - Reports pass/fail status for each feature
# - Does NOT auto-delete (manual cleanup for safety)
```

## Feature Support Matrix

| Feature                  | MD→Confluence | Confluence→MD | Status  | Notes                                       |
| ------------------------ | :-----------: | :-----------: | ------- | ------------------------------------------- |
| **Text Formatting**      |               |               |         |                                             |
| Bold `**text**`          |       ✅       |       ✅       | Working |                                             |
| Italic `*text*`          |       ✅       |       ✅       | Working |                                             |
| Bold+Italic `***text***` |       ✅       |       ✅       | Working |                                             |
| Strikethrough `~~text~~` |       ✅       |       ✅       | Working | GFM extension                               |
| Inline code `` `code` `` |       ✅       |       ✅       | Working |                                             |
| **Headings**             |               |               |         |                                             |
| H1-H6                    |       ✅       |       ✅       | Working |                                             |
| **Code Blocks**          |               |               |         |                                             |
| Fenced with language     |       ✅       |       ✅       | Working | Uses `<ac:structured-macro ac:name="code">` |
| Fenced without language  |       ✅       |       ✅       | Working | Language set to `none`                      |
| Indented (4-space)       |       ✅       |       ✅       | Working |                                             |
| Special chars in code    |       ✅       |       ✅       | Working | Backslashes, regex, quotes preserved        |
| Empty code blocks        |       ✅       |       ⚠️       | Minor   | Returns empty block with `none` language    |
| **Lists**                |               |               |         |                                             |
| Unordered                |       ✅       |       ✅       | Working |                                             |
| Ordered                  |       ✅       |       ✅       | Working |                                             |
| Nested (3+ levels)       |       ✅       |       ✅       | Working | Tight list formatting preserved             |
| Mixed nested             |       ✅       |       ✅       | Working |                                             |
| Task lists `- [ ]`       |       ✅       |       ✅       | Working | Uses Confluence `<ac:task-list>` macros     |
| **Tables**               |               |               |         |                                             |
| Basic tables             |       ✅       |       ✅       | Working |                                             |
| Column alignment         |       ✅       |       ⚠️       | Partial | Alignment lost on return (CSS-based)        |
| Empty cells              |       ✅       |       ✅       | Working |                                             |
| Escaped pipes `\|`       |       ✅       |       ✅       | Working |                                             |
| Formatted headers        |       ✅       |       ✅       | Working |                                             |
| **Links**                |               |               |         |                                             |
| Basic links              |       ✅       |       ✅       | Working |                                             |
| Link titles              |       ✅       |       ⚠️       | Partial | Rendered but Confluence may strip           |
| AutoLinks `<url>`        |       ✅       |       ✅       | Working | Converted to regular links                  |
| Email autolinks          |       ✅       |       ✅       | Working |                                             |
| Reference-style links    |       ✅       |       ✅       | Working | Resolved during parse                       |
| **Images**               |               |               |         |                                             |
| External images          |       ✅       |       ✅       | Working | Uses `<ac:image>` macro                     |
| Alt text                 |       ⚠️       |       ❌       | Partial | Alt text not preserved in Confluence        |
| **Blockquotes**          |               |               |         |                                             |
| Simple                   |       ✅       |       ✅       | Working |                                             |
| Nested                   |       ✅       |       ✅       | Working |                                             |
| With formatting          |       ✅       |       ✅       | Working |                                             |
| With lists               |       ✅       |       ✅       | Working |                                             |
| With code blocks         |       ✅       |       ✅       | Working |                                             |
| **Horizontal Rules**     |               |               |         |                                             |
| `---`, `***`, `___`      |       ✅       |       ✅       | Working | All render as `<hr />`                      |
| **Special Characters**   |               |               |         |                                             |
| Unicode text             |       ✅       |       ✅       | Working |                                             |
| HTML entities            |       ✅       |       ✅       | Working | Properly escaped/unescaped                  |
| Emoji                    |       ✅       |       ✅       | Working |                                             |
| **Edge Cases**           |               |               |         |                                             |
| Escaped chars `\*`       |       ✅       |       ✅       | Working | Properly preserved in non-code text         |
| Hard line breaks         |       ✅       |       ✅       | Working | Both `  ` and `\` work                      |
| Double-backtick code     |       ✅       |       ✅       | Working |                                             |
| Consecutive code blocks  |       ✅       |       ✅       | Working |                                             |

### Legend

- ✅ Working correctly
- ⚠️ Works with minor issues
- ❌ Not working

## Confluence Limitations

These are limitations of Confluence itself that cannot be fixed in acon.

### Table Column Alignment

Confluence stores table alignment as CSS styles on cells. When converting back to Markdown, these styles are not easily recoverable. Tables will render correctly in Confluence but alignment markers (`:---`, `:---:`, `---:`) are lost on round-trip.

### Image Alt Text

Confluence's `<ac:image>` macro does not have a direct equivalent for alt text. Images are converted correctly but alt text is not preserved:

```xml
<!-- Confluence image format - no alt text field -->
<ac:image><ri:url ri:value="https://example.com/image.png" /></ac:image>
```

### Link Title Attributes

Markdown link titles `[text](url "title")` are rendered to Confluence with the `title` attribute, but Confluence may strip this attribute during storage. The title is included in the output but may not survive Confluence's processing.

## Unfixable Gaps

These limitations cannot be resolved due to fundamental differences between Markdown and Confluence.

### 1. Confluence-Specific Macros

Confluence has many macros (panels, info boxes, expand, etc.) that have no Markdown equivalent. When converting from Confluence to Markdown, these are either:

- Stripped entirely
- Converted to basic HTML (which is then omitted for security)

### 2. Page Links and Attachments

Confluence internal links use `<ac:link>` with `<ri:page>` references:

```xml
<ac:link>
    <ri:page ri:content-title="Page Title" />
</ac:link>
```

These cannot be meaningfully converted to Markdown without access to the Confluence instance.

### 3. Mentions and User References

Confluence `@mentions` use `<ac:link>` with `<ri:user>` references that require user account IDs.

### 4. Inline Comments

Confluence inline comments are stored separately and not part of the storage format body.

## Confluence Storage Format Reference

### Code Blocks

```xml
<ac:structured-macro ac:name="code">
    <ac:parameter ac:name="language">go</ac:parameter>
    <ac:plain-text-body><![CDATA[
func main() {
    fmt.Println("Hello")
}
]]></ac:plain-text-body>
</ac:structured-macro>
```

### Images

```xml
<ac:image>
    <ri:url ri:value="https://example.com/image.png" />
</ac:image>
```

### Task Lists

```xml
<ac:task-list>
    <ac:task>
        <ac:task-status>incomplete</ac:task-status>
        <ac:task-body>Unchecked item</ac:task-body>
    </ac:task>
    <ac:task>
        <ac:task-status>complete</ac:task-status>
        <ac:task-body>Checked item</ac:task-body>
    </ac:task>
</ac:task-list>
```

### Internal Links

```xml
<ac:link>
    <ri:page ri:content-title="Page Title" />
    <ac:plain-text-link-body><![CDATA[Link Text]]></ac:plain-text-link-body>
</ac:link>
```

## Testing Instructions

### Using the Automated Test Script

```bash
# Build acon first
go build -o acon

# Run the round-trip test with a parent page ID
./testdata/roundtrip-test.sh PARENT_PAGE_ID

# The script will:
# 1. Create a test page under the specified parent
# 2. Retrieve the page and validate features
# 3. Open the page in your browser
# 4. Report pass/fail for 28 feature checks
```

### Manual Testing

```bash
# Create test page under a parent
cat testdata/comprehensive-test.md | ./acon page create -t "Comprehensive Test" --parent PARENT_ID

# View the page content as Markdown
./acon page view PAGE_ID

# View raw storage format
./acon page view PAGE_ID -j | jq -r '.body.storage.value'

# Update and verify changes persist
cat testdata/comprehensive-test.md | ./acon page update PAGE_ID
./acon page view PAGE_ID
```

### Testing Specific Features

Create focused test pages for isolation:

```bash
# Test just tables
echo '| A | B |\n|---|---|\n| 1 | 2 |' | ./acon page create -t "Table Test"

# Test just task lists
echo '- [ ] Todo\n- [x] Done' | ./acon page create -t "Task Test"

# Test code blocks with special characters
echo '```go\nregexp.MustCompile(`^[A-Z]:\\\\[\\w\\\\]+$`)\n```' | ./acon page create -t "Code Test"
```

## References

### Atlassian Documentation

- [Confluence Storage Format](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)
- [Confluence Storage Format for Macros](https://confluence.atlassian.com/conf59/confluence-storage-format-for-macros-792499117.html)
- [Add, Assign, and View Tasks](https://confluence.atlassian.com/doc/add-assign-and-view-tasks-590260030.html)
- [Confluence REST API v2](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/)

### Libraries Used

- [Goldmark](https://github.com/yuin/goldmark) - Markdown parser (with GFM extension)
- [html-to-Markdown v2](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown converter
