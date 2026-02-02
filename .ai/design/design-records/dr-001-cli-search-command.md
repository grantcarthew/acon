# DR-001: Search Command CLI Interface

- Date: 2025-02-02
- Status: Accepted
- Category: cli

## Problem

Users need to search Confluence content from the command line. The Confluence REST API v2 does not provide a search endpoint - only v1 does. The v1 search endpoint uses CQL (Confluence Query Language) which is powerful but complex for simple searches.

Forces at play:

- Most users want simple searches (find pages by text, title, or label)
- Power users need access to complex CQL queries
- The CLI should follow acon's existing patterns (resource operations with flags)
- Search can target multiple content types (pages, blogposts, attachments, etc.)
- CQL has many fields, but only a subset are commonly used

## Decision

Implement a top-level `acon search` command with:

1. Optional positional string for full-text search
2. Simple flags for common search fields (title, label, creator)
3. Standard modifier flags (space, limit, type, json)
4. Raw CQL flag for advanced queries
5. All search criteria combine with AND logic
6. Default behavior searches pages only (type=page)
7. Case-insensitive searches (CQL default behavior)

Command structure:

```bash
acon search [QUERY] [FLAGS]
```

## Why

Top-level command (not subcommand):

- Search is cross-cutting - can search pages, blogposts, attachments, etc.
- Making it `acon page search` would be semantically wrong when searching non-page content
- Search is common enough to warrant top-level placement

Positional string for text search:

- Makes simple searches intuitive: `acon search "api docs"`
- Follows common CLI patterns (grep, find, git log --grep)
- More ergonomic than requiring `--text` flag for most common case

Simple flags for common fields:

- Covers 90% of use cases without learning CQL
- Title, label, and creator searches are the most common needs
- Date searches are complex (relative vs absolute, format parsing) so excluded

Raw CQL flag for power users:

- Provides escape hatch for complex queries (date ranges, ancestor searches, OR logic)
- Avoids bloating the CLI with rarely-used flags
- Respects that some users will want full CQL power

Flags can appear in any order:

- Standard CLI behavior (Cobra framework default)
- Feels natural to users: `acon search --label urgent "critical"` or `acon search "critical" --label urgent`

AND logic for combining criteria:

- Most intuitive behavior: narrow down results progressively
- Matches common search UX (filters reduce results)
- OR logic available via raw CQL if needed

Default to pages only:

- acon is primarily a page management tool
- Reduces noise in results
- Users can override with `--type` flag

## Trade-offs

Accept:

- Users must learn CQL for complex queries (date ranges, OR logic, ancestor searches)
- Cannot do body-only search (CQL `text` field searches title + body + labels together)
- Date filtering requires raw CQL (too complex for simple flags)
- Limited to AND logic without raw CQL

Gain:

- Simple interface for 90% of use cases
- No need to learn CQL for basic searches
- Consistent with acon's existing flag patterns
- Clear escape hatch (--cql) when simple interface insufficient
- Flexible flag ordering feels natural

## Alternatives

Alternative 1: CQL-only interface

```bash
acon search --cql "text ~ \"api docs\""
acon search --cql "type=page and space=DEV"
```

Pro:
- No complexity in flag design
- Full CQL power exposed directly

Con:
- Requires learning CQL for even simple searches
- Verbose for common cases
- Quotes and escaping are error-prone

Rejected: Too much friction for simple searches

Alternative 2: Subcommand under page

```bash
acon page search [FLAGS]
```

Pro:
- Follows existing pattern (page is resource, search is operation)
- Consistent with page create, page list, etc.

Con:
- Semantically wrong when searching non-page content
- Users would expect `acon page search --type blogpost` to fail
- Search is cross-cutting, not page-specific

Rejected: Semantic mismatch with multi-type search capability

Alternative 3: Positional type argument

```bash
acon search <type> <query>
acon search page "api docs"
acon search blogpost "meeting notes"
```

Pro:
- Makes type explicit
- Clear what content type you're searching

Con:
- Doesn't handle queries without text (label-only, all attachments in space)
- Doesn't handle title-only or field-specific searches
- Can't search multiple types without CQL
- More verbose for default case

Rejected: Too rigid, doesn't handle edge cases

Alternative 4: Separate --text flag

```bash
acon search --text "api docs"
acon search --text "api" --title "REST"
```

Pro:
- More explicit about what's being searched

Con:
- More verbose for most common case
- Less intuitive than positional argument
- Doesn't match common CLI patterns (grep, find, etc.)

Rejected: Unnecessary verbosity for common case

Alternative 5: Include date flags

```bash
acon search "api" --created-after "7d"
acon search "bug" --modified-after "2024-01-01"
```

Pro:
- Date filtering is useful

Con:
- Complex to implement (relative vs absolute, format parsing, validation)
- Two date fields (created, modified) multiplied by operators (before, after, between)
- Edge cases (timezone handling, date format ambiguity)
- Adds significant complexity for moderate value

Rejected: Complexity outweighs benefit, use raw CQL instead

## Usage Examples

Simple full-text search:

```bash
acon search "api documentation"
```

Search with space filter:

```bash
acon search "meeting notes" -s DEV
```

Search different content type:

```bash
acon search "diagram" --type attachment -s TEAM
```

Title and body combination:

```bash
acon search --title "Security Review" "critical"
```

Multiple filters:

```bash
acon search "api" --label important --creator me -s DEV
```

Label-only search:

```bash
acon search --label urgent
```

Creator filter:

```bash
acon search --creator me --label todo
acon search "refactor" --creator user@example.com
```

Advanced CQL:

```bash
acon search --cql "type=page and ancestor=123456 and created>=startOfDay('-7d')"
```

JSON output:

```bash
acon search "api" -s DEV -j
```

Limit results:

```bash
acon search "bug" -l 10
```

## Implementation Notes

API Version:

- Use v1 API for search (v2 does not have search endpoint)
- Add `doRequestV1` method to API client alongside existing `doRequest`
- v1 endpoint: `/wiki/rest/api/search?cql={cql_query}`

CQL Query Construction:

- Build CQL from flags: combine with AND operator
- URL-encode the complete CQL query
- Default to `type=page` unless overridden
- Positional string maps to `text ~ "query"`
- `--title` maps to `title ~ "value"`
- `--label` maps to `label = "value"` (exact match)
- `--creator me` maps to `creator = currentUser()`
- `-s/--space` maps to `space = KEY`

Pagination:

- Use cursor-based pagination (not offset)
- Extract cursor from `_links.next` in response
- Default limit: 25 (API default)

Response Parsing:

- v1 search returns different structure than v2
- Extract: title, space, url, excerpt, lastModified
- For non-JSON output, format as: Title (Space)\nURL\nExcerpt\nModified: date\n---

Error Handling:

- Fail fast on 429 rate limiting (no automatic retry)
- Include Retry-After header value in error message if present
- Validate mutually exclusive flags (positional string with --cql)

Testing:

- Comprehensive table-driven tests (15-20 cases)
- Mock v1 endpoint with httptest.NewServer
- Test CQL construction for all flag combinations
- Test URL encoding of special characters
- Test pagination with cursors
- Test error responses (400, 401, 404, 429, 500)
- Test empty results
- Test response parsing

Output Format (non-JSON):

```
Title (SPACE_KEY)
https://example.atlassian.net/wiki/spaces/DEV/pages/123456/Title
...excerpt with matching text...
Modified: 2024-01-15

Another Page (DEV)
https://example.atlassian.net/wiki/spaces/DEV/pages/789012/Another+Page
...another excerpt...
Modified: 2024-01-10
---
```
