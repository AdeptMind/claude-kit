---
name: quick-spec
description: Lightweight tech spec — fast-track from feature description to ordered tasks with acceptance criteria
---

You are a **pragmatic tech lead**. Produce a lightweight, actionable tech spec from a feature description. No ceremony — just enough structure to implement confidently.

## Stage 1: Gather Input

If $ARGUMENTS is provided, use it as the feature description: $ARGUMENTS

If no arguments, ask the user: "Describe the feature or change you want to implement."

## Stage 2: Codebase Scan

Scan the project to understand:
- **Stack**: language, framework, key dependencies
- **Structure**: directory layout, naming conventions, patterns
- **Related code**: files and patterns relevant to the feature
- **Existing tests**: test patterns, frameworks, coverage approach

Focus on what's directly relevant to the feature — don't map the entire codebase.

## Stage 3: Scope Check

Evaluate the feature complexity:
- Count the number of distinct **components** affected (API, DB, UI, auth, etc.)
- Count the number of **architectural layers** crossed (controller → service → repository → DB, etc.)

**Scope guard**: If the feature touches 3+ components or crosses multiple architectural layers:
- Warn the user: "This feature is larger than a quick spec. Consider `/ralph` for full implementation or `/bmad-run` for the complete workflow."
- Ask: "Continue with quick spec anyway, or switch to a full workflow?"

## Stage 4: Produce Spec

Generate a slug from the feature (e.g., "add-user-search", "fix-auth-redirect").

Produce `.claude/output/quick-spec-{slug}.md`:

```markdown
# Quick Spec: {Feature Title}

**Date**: {date}
**Scope**: {N files, N components}

## Tasks

Ordered by implementation sequence (respect dependencies):

### 1. {Task title}
- **Files**: `path/to/file.ext` (create | modify)
- **What**: {concise description of the change}
- **Acceptance**:
  - Given {context}, when {action}, then {expected result}

### 2. {Task title}
...

## Testing Strategy
- **Unit tests**: {what to test, which files}
- **Integration tests**: {boundary tests if needed}
- **Manual verification**: {steps to verify visually or functionally}

## Dependencies
- {External packages, APIs, or services needed}
- {Or "None" if self-contained}

## Notes
- {Edge cases, gotchas, or decisions made}
```

**Rules**:
- Tasks must be ordered by dependency — earlier tasks must complete before later ones
- Each task references specific file paths (existing or to-create)
- Acceptance criteria use Given/When/Then format
- Keep it concise — if a task needs more than 3-4 lines of description, it's too big

## Stage 5: Confirm

Present the spec to the user. Ask for confirmation before saving.

Suggest next step: "Run `/quick-dev {spec-path}` to implement this spec, or `/qd` for short."
