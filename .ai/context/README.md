# Context

Project-specific reference material and documentation.

## Purpose

This directory contains documentation specific to this project:

- API documentation
- Framework guides
- Architecture references
- Third-party library docs

Unlike the global `~/.ai/context/`, this is for project-local reference material.

## Structure

```
context/
├── index.csv           # Master index of available sources
├── repos.csv           # Repositories to clone
├── docs.csv            # Web docs to download
├── indexes/            # Child indexes (tracked)
│   └── {dirname}.csv   # Detailed index per source
├── README.md           # This file
├── AGENTS.md           # AI agent instructions
├── project.md          # Index maintenance project
├── role.md             # AI agent role definition
├── refresh-context     # Update script
├── scripts/            # Utility scripts
└── docs/               # Local documentation files
```

## Usage

1. Add repos to `repos.csv`
2. Run `refresh-context` to clone
3. Add entry to `index.csv`
4. Create detailed index in `indexes/{dirname}.csv`

See AGENTS.md for detailed instructions.
