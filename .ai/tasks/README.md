# Task Documents

This directory contains task documents that guide AI agents through different types of work.

Task documents provide focused instructions for specific activities like design work or implementation work. They're meant to be read by agents when starting a particular type of task.

For an overarching interactive workflow that coordinates these tasks, see [.ai/workflow.md](../workflow.md).

---

## Available Tasks

### [design-phase.md](./design-phase.md)

Guides agents through design work - asking questions, exploring alternatives, making decisions, and creating Design Records.

When to use: Starting design work, making architectural decisions, documenting trade-offs.

Key activities:

- Ask probing design questions
- Explore alternatives and trade-offs
- Create Design Records (DRs)
- Maintain DR index

### [implementation-phase.md](./implementation-phase.md)

Guides agents through implementation work - translating design decisions into code while following DRs.

When to use: Writing code, implementing features, building according to design decisions.

Key activities:

- Implement according to DRs
- Reference DRs in code comments
- Write tests
- Flag design gaps if discovered

### [code-review.md](./code-review.md)

Guides agents through comprehensive code review, ensuring correctness, maintainability, and alignment with Design Records.

When to use: Reviewing code for quality, verifying implementation matches design, identifying issues before deployment.

Key activities:

- Review code against Design Records
- Analyze correctness, design, and quality
- Check error handling and testing
- Create rectification project for issues found

### [project-review.md](./project-review.md)

Guides agents through reviewing a project document before implementation begins.

When to use: Starting a new project, ensuring project is well-defined and ready for work.

Key activities:

- Analyse codebase for relevant context
- Research dependencies and best practices
- Add Current State section
- Add Decision Points for unresolved questions

### [project-completion.md](./project-completion.md)

Guides agents through completing a project and transitioning to the next one.

When to use: All success criteria met, project ready to close.

Key activities:

- Verify all criteria and deliverables complete
- Move project to completed/
- Update project index and AGENTS.md
- Set next project as active

---

## How to Use Task Documents

In AGENTS.md or project documents, reference the appropriate task when you want agents to focus on specific work:

```markdown
## Current Task

Read .ai/tasks/design-phase.md and help make design decisions for the authentication system.
```

```markdown
## Implementation

Follow .ai/tasks/implementation-phase.md to implement the features defined in p-001.
```

As standalone instructions, task documents are self-contained - an agent can read just the task document and understand what to do.

---

## Creating New Task Documents

If you identify other types of work that would benefit from focused agent guidance, create new task documents here.

Format: `task-name.md`

Each task document should:

- Define the objective
- Explain the agent's role
- Provide workflow steps
- Include specific instructions for agents
- Reference relevant guides and resources
