# Project Completion Task

This document guides AI agents through completing a project and transitioning to the next one.

---

## When to Use

Use this task when all success criteria in the active project are met and the project is ready to be closed.

---

## Workflow

1. Verify completion:
   - All success criteria are checked off
   - All deliverables exist
   - No open Decision Points remain

2. Move the project document:
   - Move from `.ai/projects/p-NNN-*.md` to `.ai/projects/completed/`

3. Update the project index:
   - Update `.ai/projects/README.md` with completion date
   - Update status from Active to Completed

4. Review related documents:
   - Check if any DRs need status updates
   - Verify deliverables are documented where expected
   - Clean up any temporary notes

5. Assign the next project:
   - Identify the next project (usually next numbered pending project)
   - Update `AGENTS.md` to set the new Active Project
   - Set the new project status to Active and add start date

6. Inform the user:
   - Summarise what was completed
   - Confirm the next active project

---

## Checklist

Before marking complete:

- [ ] All success criteria met
- [ ] All deliverables created
- [ ] Project moved to completed/
- [ ] Project index updated
- [ ] AGENTS.md points to next project
- [ ] Next project status set to Active
