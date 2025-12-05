#!/bin/bash
#
# Round-trip test for acon Markdown conversion
#
# Tests the full cycle: Markdown -> Confluence -> Markdown
# Compares key features to verify conversion fidelity
#
# Usage: ./testdata/roundtrip-test.sh <parent-page-id>
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "\n${GREEN}==>${NC} $1"
}

# Check for required parent ID argument
if [[ $# -lt 1 ]]; then
    log_error "Parent page ID is required"
    echo ""
    echo "Usage: $0 <parent-page-id>"
    echo ""
    echo "Example: $0 1857355975"
    exit 1
fi

PARENT_ID="$1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TEST_FILE="$SCRIPT_DIR/comprehensive-test.md"
ACON_BIN="$PROJECT_DIR/acon"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
PAGE_TITLE="Round-Trip Test $TIMESTAMP"

log_step "Starting round-trip test"
log_info "Parent page ID: $PARENT_ID"
log_info "Test file: $TEST_FILE"
log_info "Page title: $PAGE_TITLE"

# Check test file exists
if [[ ! -f "$TEST_FILE" ]]; then
    log_error "Test file not found: $TEST_FILE"
    exit 1
fi
log_success "Test file exists"

# Check environment variables
log_step "Checking environment configuration"
if [[ -z "${CONFLUENCE_BASE_URL:-}" ]]; then
    log_error "CONFLUENCE_BASE_URL is not set"
    exit 1
fi
log_success "CONFLUENCE_BASE_URL is set"

if [[ -z "${CONFLUENCE_EMAIL:-}" ]]; then
    log_error "CONFLUENCE_EMAIL is not set"
    exit 1
fi
log_success "CONFLUENCE_EMAIL is set"

if [[ -z "${CONFLUENCE_API_TOKEN:-}${ATLASSIAN_API_TOKEN:-}${JIRA_API_TOKEN:-}" ]]; then
    log_error "No API token set (CONFLUENCE_API_TOKEN, ATLASSIAN_API_TOKEN, or JIRA_API_TOKEN)"
    exit 1
fi
log_success "API token is set"

# Build acon if needed
log_step "Building acon"
cd "$PROJECT_DIR"
if [[ ! -f "$ACON_BIN" ]] || [[ "main.go" -nt "$ACON_BIN" ]] || [[ -n "$(find cmd internal -name '*.go' -newer "$ACON_BIN" 2>/dev/null)" ]]; then
    log_info "Source files changed, rebuilding..."
    go build -o acon
    log_success "Build completed"
else
    log_info "Binary is up to date, skipping build"
fi

# Verify acon works
ACON_VERSION=$("$ACON_BIN" --version 2>&1 || echo "unknown")
log_info "acon version: $ACON_VERSION"

# Create the test page
log_step "Creating test page in Confluence"
log_info "Uploading comprehensive-test.md..."

CREATE_OUTPUT=$(cat "$TEST_FILE" | "$ACON_BIN" page create -t "$PAGE_TITLE" --parent "$PARENT_ID" 2>&1)

# Extract page ID from output
PAGE_ID=$(echo "$CREATE_OUTPUT" | grep -E "^ID:" | awk '{print $2}')
PAGE_URL=$(echo "$CREATE_OUTPUT" | grep -E "^URL:" | awk '{print $2}')

if [[ -z "$PAGE_ID" ]]; then
    log_error "Failed to create page"
    echo "$CREATE_OUTPUT"
    exit 1
fi

log_success "Page created successfully"
log_info "Page ID: $PAGE_ID"
log_info "Page URL: $PAGE_URL"

# View the page back
log_step "Retrieving page from Confluence"
log_info "Fetching page content as Markdown..."

TEMP_DIR=$(mktemp -d)
RETRIEVED_FILE="$TEMP_DIR/retrieved.md"

# Get the content (skip the header lines from acon page view)
"$ACON_BIN" page view "$PAGE_ID" | tail -n +6 > "$RETRIEVED_FILE"

ORIGINAL_LINES=$(wc -l < "$TEST_FILE" | tr -d ' ')
RETRIEVED_LINES=$(wc -l < "$RETRIEVED_FILE" | tr -d ' ')

log_success "Page retrieved successfully"
log_info "Original file: $ORIGINAL_LINES lines"
log_info "Retrieved file: $RETRIEVED_LINES lines"

# Run feature checks
log_step "Verifying round-trip conversion"

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

check_feature() {
    local name="$1"
    local pattern="$2"
    local file="$RETRIEVED_FILE"

    if grep -qE "$pattern" "$file"; then
        log_success "$name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_error "$name - pattern not found: $pattern"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

check_feature_warn() {
    local name="$1"
    local pattern="$2"
    local file="$RETRIEVED_FILE"

    if grep -qE "$pattern" "$file"; then
        log_success "$name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        log_warn "$name - pattern not found (known limitation)"
        WARN_COUNT=$((WARN_COUNT + 1))
    fi
}

# Text formatting
log_info "Checking text formatting..."
check_feature "Bold text" "\*\*bold text\*\*"
check_feature "Italic text" "\*italic text\*"
check_feature "Strikethrough" "~~strikethrough~~"
check_feature "Inline code" "\`inline code\`"

# Headings
log_info "Checking headings..."
check_feature "H2 heading" "^## Text Formatting"
check_feature "H3 heading" "^### Basic Formatting"
check_feature "H4 heading" "^#### This is H4"
check_feature "H5 heading" "^##### This is H5"
check_feature "H6 heading" "^###### This is H6"

# Code blocks
log_info "Checking code blocks..."
check_feature "Go code block" '```go'
check_feature "Python code block" '```python'
check_feature "JavaScript code block" '```javascript'
check_feature "Bash code block" '```bash'
check_feature "HTML code block" '```html'

# Lists
log_info "Checking lists..."
check_feature "Unordered list" "^- First item"
check_feature "Ordered list" "^1\. Step one"
check_feature "Nested list" "Child item"

# Task lists
log_info "Checking task lists..."
check_feature_warn "Unchecked task" "^\- \[ \]"
check_feature_warn "Checked task" "^\- \[x\]"

# Tables
log_info "Checking tables..."
check_feature "Table header" "\| Feature .* \| Status .* \| Notes"
check_feature "Table row" "\| Headings .* Working"

# Links
log_info "Checking links..."
check_feature "External link" "\[Atlassian Documentation\]"
check_feature "Link URL" "https://developer.atlassian.com"

# Blockquotes
log_info "Checking blockquotes..."
check_feature "Simple blockquote" "^> This is a simple blockquote"

# Unicode
log_info "Checking unicode..."
check_feature "Japanese text" "こんにちは世界"
check_feature "Emoji" "🚀"

# Special characters
log_info "Checking special characters..."
check_feature "Ampersand" "Ampersands &"
check_feature "Angle brackets" "< >"

# Cleanup temp files
rm -rf "$TEMP_DIR"

# Summary
log_step "Test Summary"
echo ""
log_info "Passed: $PASS_COUNT"
if [[ $WARN_COUNT -gt 0 ]]; then
    log_info "Warnings: $WARN_COUNT (known limitations)"
fi
if [[ $FAIL_COUNT -gt 0 ]]; then
    log_info "Failed: $FAIL_COUNT"
fi
echo ""
log_info "Page URL: $PAGE_URL"
log_info "Page ID: $PAGE_ID"
echo ""

# Open in browser
log_step "Opening page in browser"
if [[ "$OSTYPE" == "darwin"* ]]; then
    open "$PAGE_URL"
    log_success "Opened in default browser (macOS)"
elif command -v xdg-open &> /dev/null; then
    xdg-open "$PAGE_URL"
    log_success "Opened in default browser (Linux)"
elif command -v wslview &> /dev/null; then
    wslview "$PAGE_URL"
    log_success "Opened in default browser (WSL)"
else
    log_warn "Could not detect browser opener - please open URL manually"
fi

echo ""
log_warn "Remember to delete the test page manually when done reviewing"
log_info "Page ID to delete: $PAGE_ID"

# Exit with appropriate code
if [[ $FAIL_COUNT -gt 0 ]]; then
    exit 1
else
    exit 0
fi
