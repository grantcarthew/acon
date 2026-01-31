# Context Directory

Project-specific reference material and cloned documentation for AI agent consumption.

## Directory Structure

```
.ai/context/
├── index.csv           # Master index of available sources (for agents)
├── repos.csv           # Repositories to clone (for scripts)
├── docs.csv            # Web docs to download (for scripts)
├── indexes/            # Child indexes for each source (tracked)
│   └── {dirname}.csv   # Detailed file index for each source
├── README.md           # Human-readable overview
├── AGENTS.md           # This file - AI agent instructions
├── project.md          # Index maintenance project
├── role.md             # AI agent role definition
├── refresh-context     # Main update script
├── .gitignore          # Ignore cloned repos
├── docs/               # Local documentation (tracked)
├── scripts/            # Utility scripts (tracked)
│   ├── lib.sh          # Shared functions
│   ├── refresh-repos    # Clone/update repos from repos.csv
│   └── refresh-docs     # Download web docs from docs.csv
└── [cloned-repos]/     # Cloned reference repos (gitignored)
```

## Discovery Flow

1. Read `index.csv` to see available sources with descriptions and topics
2. Read `indexes/{dirname}.csv` for detailed file listings within a source
3. Navigate directly to the files you need

## Adding a New Reference Source

1. Add entry to repos.csv:

```csv
url,directory,sparse_paths,ref
https://github.com/org/repo.git,repo-name,,
```

2. Run refresh-context to clone:

```bash
.ai/context/refresh-context
```

3. Add entry to index.csv:

```csv
directory,description,topics,source_url,source_type,last_updated
repo-name,Brief description,topic1;topic2,https://github.com/org/repo,official-tool-repo,YYYY-MM-DD
```

4. Create child index at `indexes/{dirname}.csv` with appropriate schema

5. Apply token efficiency principle (see project.md):
   - Include: READMEs, API docs, usage guides, key source files
   - Exclude: CONTRIBUTING, CI/CD configs, test infrastructure

## Maintenance

Update all cloned repositories:

```bash
.ai/context/refresh-context
```

## Search and Navigation

Read indexes directly (most token-efficient):

```bash
# See what sources are available
cat .ai/context/index.csv

# Search for topics across all indexes
rg -i "topic" .ai/context/indexes/
```

Full-text search:

```bash
rg -i "search-term" .ai/context/
```

## Related Files

- [index.csv](index.csv) - Master index of available sources
- [README.md](README.md) - Human-readable overview
- [project.md](project.md) - Index maintenance instructions
- [role.md](role.md) - AI agent role definition
