---
description: Coordination rules for agent team sessions — file ownership, messaging, contracts, and quality gates
---

## Contract-First Development

When multiple teammates work in parallel within the same round:
- **Define interfaces before implementation**: shared types, API contracts, DB schemas, and module boundaries must be committed BEFORE teammates start coding
- Teammates implement AGAINST the committed contracts, not their own assumptions
- If a contract needs to change, the teammate messages the lead — the lead updates the contract and notifies affected teammates

## Worktree Isolation

Each teammate MUST work in an isolated git worktree — never on the shared working tree.

### Setup (lead responsibility)

Before dispatching a teammate, the lead creates the worktree:
```bash
git worktree add .worktrees/<story-id> -b <story-id>
```

The teammate receives the worktree path and works exclusively within it. This prevents:
- File conflicts between parallel teammates
- Dirty working tree breaking other teammates' builds
- Merge conflicts from simultaneous edits

### Cleanup (lead responsibility)

After a story is validated and merged:
```bash
git worktree remove .worktrees/<story-id>
git branch -d <story-id>
```

### Rules

- `.worktrees/` MUST be in `.gitignore` — add it immediately if missing
- Each worktree has its own `node_modules`/`vendor` — run the setup command (npm install, go mod download, etc.) after creation
- Run the test suite in the worktree BEFORE starting work to establish a green baseline
- If a worktree's tests fail at baseline, do NOT start coding — message the lead

### Merge Strategy

The lead merges completed worktrees into the main branch in dependency order:
1. Merge story A (no deps) → run full test suite
2. Merge story B (depends on A) → run full test suite
3. If conflicts arise, the lead resolves them — teammates do NOT rebase each other's branches

## File Ownership

- Each teammate OWNS a distinct set of files — no two teammates edit the same file in the same round
- Shared contracts (types, interfaces, schemas) are owned by the lead and committed before the round starts
- If a teammate needs to modify a file owned by another, they message the lead to coordinate

## Messaging Protocol

- **Before implementing**: if unclear about a requirement or interface, message the lead — do NOT guess
- **When blocked**: message the lead immediately with what you need and from whom — do NOT wait silently
- **When done**: message the lead with a completion report (files, tests, issues)
- **When finding issues**: if you discover a problem in another teammate's code during integration, message the lead — do NOT fix it yourself

## Plan Before Code

- Teammates MUST describe their implementation plan before writing code
- The plan should cover: files to create/modify, approach, how acceptance criteria will be met
- The lead reviews and approves the plan — only then does the teammate start coding
- This prevents wasted work from wrong assumptions or duplicate effort

## Quality Gates

- A story is NOT passed until the lead validates it with the acceptance-validator skill
- Tests must pass before marking a story as complete
- The full test suite must pass after each round (not just individual story tests)
- If validation fails, the teammate fixes the issues before the lead marks the story as passed

## Dependency Management

- Stories with `dependsOn` NEVER start until all dependencies are validated and passed
- Within a round, if story B needs output from story A (both in the same round), they must be moved to sequential execution — B runs after A
- The lead detects these hidden dependencies during the contract phase and adjusts the round grouping
