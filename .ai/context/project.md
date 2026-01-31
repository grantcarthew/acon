# Context Index Maintenance

## Project Overview

Maintain the context directory's index system. The directory contains cloned documentation sources and reference material.

Project Type: Repository maintenance and indexing

Status: Recurring maintenance task

Critical Constraints:

- Read-only/Create/Update operations only - NEVER delete files
- Root index.csv contains child directories only - NEVER include root-level files

## Index Files

| File | Purpose | Audience |
|------|---------|----------|
| `index.csv` | Master index of available sources | Agents |
| `repos.csv` | Repositories to clone | Scripts |
| `indexes/{dirname}.csv` | Detailed file index per source | Agents |

## Goals

1. Discover all cloned repositories and local directories
2. Ensure `index.csv` has entries for all sources
3. Ensure each source has a corresponding `indexes/{dirname}.csv`
4. Keep indexes accurate and up-to-date

## Technical Approach

### Discovery

Find all git repositories:

```bash
fd --type d --hidden --no-ignore --min-depth 2 '^\.git$' .ai/context/
```

List existing indexes:

```bash
ls .ai/context/indexes/
```

### Master Index (index.csv)

Schema:

```csv
directory,description,topics,source_url,source_type,last_updated
```

Fields:

- `directory`: Child directory name
- `description`: Brief description of the content
- `topics`: Semicolon-separated topics
- `source_url`: Original URL (or N/A for local)
- `source_type`: git-sparse-checkout, official-docs, official-tool-repo, local-docs
- `last_updated`: YYYY-MM-DD format

### Child Indexes (indexes/{dirname}.csv)

Each source gets its own index in `indexes/`. Choose schema based on content type:

- Documentation: `file,description,topics`
- Source code: `component,description,path,language`
- Tools: `area,description,path,topics`

Apply token efficiency principle:

- Include: Content that helps LLMs USE the tool/library
- Exclude: Contributor-focused content (CONTRIBUTING, CI/CD, tests)

Use relative paths from source root.

### Validation

1. Every directory in `index.csv` should exist
2. Every directory should have an `indexes/{dirname}.csv`
3. Paths in child indexes should be valid
4. NO FILES DELETED

## Success Criteria

- `index.csv` lists all available sources
- Each source has `indexes/{dirname}.csv`
- Index schemas fit the content type
- All indexes are valid CSV
- Topics are meaningful

## Related Documentation

- [AGENTS.md](AGENTS.md) - AI agent instructions
- [role.md](role.md) - AI agent role definition
