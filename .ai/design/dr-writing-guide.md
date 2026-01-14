# DR Writing Guide

Creating and maintaining Design Records.

Location: `.ai/design/design-records/dr-NNN-title.md`

Read when: Writing/updating DRs or reconciling docs.

---

## Markdown Formatting Note

IMPORTANT: Avoid using bold Markdown (`**text**`) in design records. Bold formatting adds no semantic value to large language models but significantly increases token count. Use section headers, lists, and clear structure instead.

Good: Use `## Section Name` and bullet points
Bad: Use `**Section Name:**` with bold

Note: This guidance applies to DR and Project documents consumed by AI agents. Human-facing documentation may use bold for readability.

---

## DR Schema

DR structure:

### Header

```markdown
# DR-NNN: Title

- Date: YYYY-MM-DD
- Status: Accepted | Superseded
- Category: (technology/component/area)
```

### Required Sections

Problem:

- What constraint or issue drove this decision?
- What forces are at play?
- What problem does this solve?

Decision:

- Clear, specific statement of what was decided
- Should be implementable and testable

Why:

- Core reasoning behind this choice
- Why is this the right solution for our context?
- Supporting details that explain the decision

Trade-offs:

- Accept: What costs, limitations, or complexity are we accepting?
- Gain: What benefits, simplicity, or capabilities do we get?

Alternatives:

- What other options were considered?
- Why were they rejected?
- What were their trade-offs?

### Optional Sections

Add as needed:

- Structure - Schema definitions, field descriptions
- Scope - Where/how this applies (global vs local, etc.)
- Usage Examples - How to use the decision in practice
- Validation - Rules for correctness
- Execution Flow - Step-by-step behavior
- Implementation Notes - High-level guidance for implementers (not code)
- Security - Security considerations and threat model
- Breaking Changes - Updates from previous versions
- Updates - Historical changes with dates

---

## What Belongs in DRs

### Configuration Examples

Config structure and schema (TOML, JSON, etc.):

```toml
[database]
host = "localhost"
port = 5432
timeout = 30
pool_size = 10

[cache.redis]
enabled = true
ttl = 3600

  [cache.redis.endpoints]
  primary = "redis-01.example.com:6379"
  secondary = "redis-02.example.com:6379"
```

This is NOT implementation code - it's the schema/structure being defined.

### Usage Examples

Behavior and command usage:

```bash
myapp deploy --environment prod --verbose
myapp run migration --dry-run "add user index"
```

### Field Descriptions

Field meanings and constraints:

status (string, required):

- Current state of the resource
- Valid values: `"pending"`, `"active"`, `"completed"`, `"failed"`
- Default: `"pending"`
- Immutable once set to `"completed"` or `"failed"`

### Validation Rules

Validation requirements:

At resource creation:

- Resource name matches pattern: `/^[a-z][a-z0-9-]*[a-z0-9]$/`
- Name length between 3 and 63 characters
- At least one of `endpoint`, `config_file`, or `inline_config` must be present
- If `ttl` is specified, must be between 60 and 86400 seconds

### Execution Flows

Step-by-step algorithms:

When `myapp process <job-id>` is executed:

1. Determine retry strategy:
   - CLI `--retry-policy` flag - use specified policy
   - Else job `retry_policy` field - use job-specific policy
   - Else global `default_retry_policy` - use system default
   - If all fail - use built-in exponential backoff

2. Execute processing:
   - Load job data from storage
   - Apply retry policy if previous attempts exist
   - Process according to job type
   - Update status atomically

### Tables and Matrices

Scope and behavior matrices:

| Feature      | Free Tier | Pro Tier | Enterprise |
| ------------ | --------- | -------- | ---------- |
| API Access   | Yes       | Yes      | Yes        |
| Rate Limit   | 100/hour  | 1000/hr  | Unlimited  |
| Webhooks     | -         | Yes      | Yes        |
| SLA          | -         | 99.9%    | 99.99%     |
| Support      | Community | Email    | 24/7 Phone |

### Breaking Changes Notes

Historical changes:

Breaking Changes from v1.x:

1. Changed: `timeout` field now measured in seconds (previously milliseconds)
2. Removed: `legacy_mode` flag (superseded by compatibility layer)
3. Added: `retry_strategy` field for configurable retry behavior
4. Renamed: `endpoint_url` to `endpoint` for consistency

### Updates Section

Dated updates:

Updates:

- 2025-01-15: Changed from fixed connection pool to dynamic sizing based on load
- 2025-01-20: Added support for both environment variables and config files
- 2025-02-01: Deprecated string-based status in favor of enum type

---

## What Does NOT Belong in DRs

### Implementation Code

Do not include actual source code (Go, Python, etc.):

```go
// Bad: Implementation code
func LoadConfig(path string) (*Config, error) {
    // ... implementation ...
}
```

Why: DRs capture decisions, not implementation.

Exception: Pseudocode in Why section for complex logic.

### Cross-Links to Other DRs

Do not link to other DR documents:

Why: Links break when DRs change.

Instead: Use `design-records/README.md` index.

Exception: Link when status changes (superseding/superseded):

```markdown
Status: Superseded by [dr-042](../dr-042-new-approach.md)
```

Note: Superseded DRs are moved to the `superseded/` directory, so they link back up to the main directory using `../`

### User Documentation Duplication

Do not duplicate content from user docs:

Bad: Duplicating full CLI docs

Good: Brief example + reference to full docs

---

## When to Create a DR

Always Create a DR for:

- Architectural decisions (component structure, data flow)
- Algorithm specifications (resolution order, search, matching)
- Breaking changes or deprecations
- Data formats, schemas, or protocols (TOML structure, field definitions)
- Public API or CLI command structure
- Security or performance trade-offs
- Major UX decisions (flag names, command organization)
- Multi-file vs single-file config decisions

Never Create a DR for:

- Simple bug fixes
- Documentation corrections (typos, clarifications)
- Code refactoring without behavior change
- Cosmetic changes
- Internal implementation details that don't affect external behavior

When Unsure, Ask:

Would a future developer need to understand WHY we made this choice?

- Yes - Create a DR
- No - Just fix it

---

## DR Numbering and Lifecycle

### File Naming Convention

Format: `dr-<NNN>-<category>-<title>.md`

- `NNN` = Three-digit number with leading zeros (001, 002, 003...)
- `category` = Technology/component/area (config, api, data, cli, etc.)
- `title` = KISS description of the decision
- All lowercase kebab-case (words separated by hyphens, no underscores or spaces)

Examples:

- `dr-001-config-file-format.md`
- `dr-002-api-authentication-strategy.md`
- `dr-003-data-storage-structure.md`
- `dr-004-cli-command-organization.md`

### Numbering

- Sequential: dr-001, dr-002, dr-003, etc.
- Gaps are acceptable (superseded DRs)
- Never reuse numbers
- Get next number from `design-records/README.md` index

### Status Values

- Accepted - Decision is in effect
- Superseded - Replaced by newer DR, link in header, move to `superseded/`

---

## Writing a Good DR

### Focus on "Why" Not "How"

Include detailed reasoning:

```markdown
## Decision

Use TOML for all configuration files

## Why

- Human-readable and editable
- No whitespace sensitivity (unlike YAML)
- Excellent Go support via BurntSushi/toml
- Supports comments and complex nested structures
```

### Be Specific

Include concrete details and behavior:

```markdown
## Decision

Single configuration file with global + local merge strategy

## Merge Behavior

- Local config merges with global
- Same keys in local override global values
- New keys in local are added
- Omitted keys use global defaults
```

### Document Trade-offs Honestly

List both costs and benefits:

```markdown
## Trade-offs

Accept:

- Users must learn TOML syntax
- More complex than simple key=value files
- Parsing requires external library

Gain:

- Comments support (critical for user guidance)
- Nested structures for complex config
- No whitespace errors (unlike YAML)
```

### Consider Alternatives Seriously

Analyze with pros, cons, and rejection reasoning:

```markdown
## Alternatives

YAML:

- Pro: Widely known, standard in DevOps
- Pro: Native Go support
- Con: Whitespace-sensitive, error-prone for hand-editing
- Con: Complex spec with surprising edge cases
- Rejected: Error-prone editing outweighs familiarity

JSON:

- Pro: Simple, universal
- Con: No comments (users can't document their config)
- Con: Less human-friendly (trailing commas, quoted keys)
- Rejected: Lack of comments is a dealbreaker
```

---

## Reconciliation Process

After design changes:

1. Remove deprecated references: `rg "old-pattern" docs/`
2. Update DR index in `design-records/README.md` with current status
3. Check DR status accuracy (Accepted, Superseded?)
4. Remove stale TODOs: `rg "TODO|TBD|to be written" docs/`
5. Verify examples match current schema
6. Remove any "Related Decisions" sections from DRs
