# Project Documents

This directory contains project documents. Each project represents a focused effort with clear goals, scope, and success criteria.

See [p-writing-guide.md](./p-writing-guide.md) for guidelines on creating and maintaining project documents.

---

## Quick Reference

| Project | Title | Status | Started | Completed |
|---------|-------|--------|---------|-----------|
| p-001 | Project Initialization | Pending | - | - |

Note: Completed projects are in `completed/`

---

## Status Values

- Pending - Project defined, not yet started
- Active - Currently being worked on
- Blocked - Waiting on external dependency or decision
- Completed - All success criteria met, deliverables created (move to `completed/`)

---

## Projects vs Design Records

Projects are work packages that define what to build and how to validate it.

Design Records (DRs) document why we chose a specific approach and the trade-offs.

A single project may generate multiple DRs. Projects describe the work; DRs document the decisions made during that work.

See [p-writing-guide.md](./p-writing-guide.md) for detailed guidance.

---

## Contributing

When creating a new project:

1. List directory to find next number: `ls .ai/projects/p-*.md`
2. Use format: `p-<NNN>-<category>-<title>.md`
3. Follow the structure in [p-writing-guide.md](./p-writing-guide.md)
4. Define clear, measurable success criteria
5. Update this README with project entry
6. Link dependencies to other projects
