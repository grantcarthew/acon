# Project Writing Guide

Creating and maintaining Project documents.

Location: `.ai/projects/p-NNN-title.md`

Read when: Writing/updating project documents or planning work.

---

## Markdown Formatting Note

IMPORTANT: Avoid using bold Markdown (`**text**`) in project documents. Bold formatting adds no semantic value to large language models but significantly increases token count. Use section headers, lists, and clear structure instead.

Good: Use `## Section Name` and bullet points
Bad: Use `**Section Name:**` with bold

Note: CLI documentation (docs/cli/) is intended for human readers and should use bold for readability. This guidance applies only to DR and Project documents that are primarily consumed by AI agents.

---

## Project Document Schema

Project structure:

### Header

```markdown
# P-NNN: Title

- Status: Pending | Active | Blocked | Completed
- Started: YYYY-MM-DD (when work begins)
- Completed: YYYY-MM-DD (when finished)
```

### Required Sections

Overview:

- Brief description of what this project accomplishes
- Why this project exists
- How it fits into the larger system

Goals:

- Specific, measurable outcomes
- What we're trying to achieve
- 3-7 concrete goals typically

Scope:

- What's included in this project
- What's explicitly excluded
- Boundaries and constraints

Success Criteria:

- How we know the project is complete
- Measurable, testable outcomes
- Must be unambiguous

Deliverables:

- Concrete outputs (files, documents, code, DRs)
- What gets created or updated
- What gets published or deployed

### Optional Sections

Add as needed:

- Current State - Relevant codebase context and existing implementation details
- Decision Points - Unresolved questions requiring owner decisions (see format below)
- Dependencies - Other projects or resources needed before starting
- Technical Approach - High-level strategy for accomplishing goals
- Research Areas - Topics to investigate
- Design Decisions - DRs that will likely emerge from this project
- Testing Strategy - How we'll validate the deliverables
- Notes - Additional context, learnings, observations
- Updates - Historical changes with dates

---

## Projects vs Design Records

### Projects Are

- Work packages with clear deliverables
- Time-bounded efforts
- Focus on what to build and how to validate it
- May generate multiple DRs
- Can span research, design, and implementation
- Answer: "What are we doing?"

### Design Records Are

- Architectural decisions
- Focus on why we chose a specific approach
- Document trade-offs and alternatives
- Created during project execution
- Permanent record of reasoning
- Answer: "Why did we do it this way?"

### Relationship

A single project may create multiple DRs. For example:

- p-001: API Redesign might generate:
  - dr-001: Authentication Strategy
  - dr-002: Error Response Format
  - dr-003: Versioning Approach

Projects describe the work to be done. DRs document the decisions made while doing that work.

---

## What Belongs in Project Documents

### Goals and Objectives

Clear statements of what we're trying to achieve:

## Goals

1. Implement standardised exit codes
2. Add verbose mode for debugging
3. Support batch operations via stdin
4. Document automation best practices

### Success Criteria

Measurable outcomes that define completion:

## Success Criteria

- [ ] Exit codes are consistent and documented
- [ ] Verbose mode displays HTTP request details
- [ ] Commands accept input from stdin
- [ ] Created DR documenting conventions

### Scope Boundaries

Clear in/out of scope definitions:

## Scope

In Scope:

- Exit code standardisation
- Verbose and quiet modes
- Batch processing support

Out of Scope:

- Interactive prompts (covered in p-005)
- Configuration files (covered in p-004)
- Webhook integration (future work)

### Deliverables

Concrete outputs:

## Deliverables

- `internal/cli/exit_codes.go` - Exit code constants
- `docs/automation.md` - Automation guide
- dr-001: CLI Automation Conventions

### Current State

Relevant codebase context:

## Current State

- Config loading uses `config.Load()` which returns validation errors
- HTTP client in `internal/api/client.go` has no request logging
- Commands return generic errors, no exit code differentiation

### Decision Points

Questions requiring owner decisions. Use numbered questions with lettered options for easy reference (e.g., "1. B"):

## Decision Points

1. Output format for verbose mode

- A: Full wire format (headers, body, timing)
- B: Simplified (method, URL, status, duration)
- C: Structured log format

2. Error handling strategy

- A: Return on first error
- B: Collect all errors and report at end

When resolved, decisions either become a DR (if architectural) or get incorporated into the project (Technical Approach, Scope, etc.). Remove resolved decision points from this section.

### Research Areas

Topics to investigate:

## Research Areas

- API rate limit behaviour and headers
- Best practices for exit codes
- Stdin pipe detection patterns

---

## What Does NOT Belong in Projects

### Implementation Code

Do not include actual source code - that goes in the repository:

```go
// Bad: Implementation code in project doc
func LoadConfig(path string) (*Config, error) {
    // ... implementation ...
}
```

Exception: Pseudocode in Technical Approach for clarity.

### Resolved Design Decisions

Resolved design decisions belong in DRs, not project docs. Use Decision Points for unresolved questions, then move answers to DRs or project context once decided.

Bad: Detailed trade-off analysis with final decision in project doc

Good: Decision Point with options, then dr-001 created when decided

### Step-by-Step Instructions

Projects define what and why, not detailed how:

Bad: "First run this command, then edit this file, then..."

Good: "Implement configuration loading using the standard library"

---

## Project Statuses

### Pending

- Project defined but not started
- Goals and scope documented
- Waiting to begin work

### Active

- Currently being worked on
- Update with notes and learnings as you go
- Track completed goals/criteria

### Blocked

- Cannot proceed due to external dependency
- Document what's blocking progress
- Link to dependency or issue

### Completed

- All success criteria met
- All deliverables created
- Set completion date
- Move to `completed/`

---

## Project Numbering and Naming

### File Naming Convention

Format: `p-<NNN>-<category>-<title>.md`

- `NNN` = Three-digit number with leading zeros (001, 002, 003...)
- `category` = Technology/component/area (cue, cli, assets, distribution, orchestration, etc.)
- `title` = KISS description of the project
- All lowercase kebab-case (words separated by hyphens, no underscores or spaces)

Examples:

- `p-001-cue-foundation-architecture.md`
- `p-002-assets-concrete-tasks.md`
- `p-003-distribution-strategy.md`
- `p-004-cli-minimal-viable.md`

### Numbering

- Sequential: p-001, p-002, p-003, etc.
- Get next number from `.ai/projects/README.md` index
- Never reuse numbers
- Gaps are acceptable (deferred/cancelled projects)

---

## Writing a Good Project Document

### Keep Projects Focused

Good project scope:

- Clear boundaries
- Achievable in reasonable timeframe
- Single area of concern
- Generates clear deliverables

Too broad:

- "Build the entire CLI" - Split into multiple projects

Too narrow:

- "Write one function" - Probably just a task, not a project

### Make Success Criteria Testable

Good criteria:

- [ ] All API endpoints return correct exit codes
- [ ] Verbose mode shows request/response details
- [ ] Created DR documenting conventions

Bad criteria:

- "Understand the API better" (not measurable)
- "Make good error handling" (subjective)

### Use Decision Points

Projects often start with questions that need owner decisions - that's okay!

## Decision Points

1. Caching strategy

- A: In-memory cache
- B: Redis
- C: No caching initially

These get resolved during the project and move to DRs or project details.

### Link to Related Work

## Dependencies

- Requires: p-001 (database setup)
- Blocks: p-005 (needs auth from this project)

## Related DRs

- Will create: dr-003 (session management)
- Builds on: dr-001 (API design)

---

## Project Lifecycle

1. Create - Define goals, scope, success criteria
2. Review - Ensure project is well-scoped and achievable
3. Start - Set status to Active, add start date
4. Execute - Work on deliverables, create DRs, resolve decision points
5. Update - Add notes, learnings as you go
6. Complete - Verify success criteria, set completion date
7. Archive - Move to Completed status, update README index

---

## Examples

### Good Project Document

```markdown
# p-001: User Authentication

- Status: Active
- Started: 2025-12-01

## Overview

Add user authentication to the application. This enables secure access
and personalised features for registered users.

## Goals

1. Implement user registration and login
2. Add session management
3. Secure API endpoints

## Success Criteria

- [ ] Users can register and log in
- [ ] Sessions persist across browser restarts
- [ ] Unauthenticated API requests return 401
- [ ] Created dr-001 documenting auth approach

## Deliverables

- `internal/auth/` - Authentication package
- `docs/auth.md` - Authentication documentation
- dr-001: Authentication Strategy
```

### Bad Project Document

```markdown
# p-001: Add Auth

Make login work.

## TODO

- Read docs
- Write code
- Test stuff
```

Problems:

- Vague goals
- No success criteria
- No clear deliverables
- Not measurable

---

## Contributing

When creating a new project:

1. Get next number from `.ai/projects/README.md`
2. Use template structure above
3. Define clear, measurable success criteria
4. Update README index
5. Link dependencies and related DRs
