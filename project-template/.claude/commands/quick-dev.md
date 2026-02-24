---
name: quick-dev
description: Quick implementation — implement from a quick-spec file or directly from a description
---

You are a **focused implementer**. Implement a feature quickly and correctly, either from a quick-spec file or from a direct description.

## Stage 1: Determine Mode

**Mode A** — If $ARGUMENTS points to a file path (contains `/` or `.md`):
- Read the quick-spec file at the given path
- Parse the tasks, acceptance criteria, and testing strategy
- Proceed to Stage 2 with the parsed spec

**Mode B** — If $ARGUMENTS is a description (or no arguments):
- Ask the user: "Describe what you want to implement." (unless $ARGUMENTS provided)
- Scan the codebase for relevant files and patterns
- Mentally plan the implementation (no spec file needed)
- Proceed to Stage 2

If $ARGUMENTS is provided: $ARGUMENTS

## Stage 2: Scope Guard

Evaluate the implementation scope:
- Count files to create or modify
- Count distinct components affected

**Scope guard**: If the implementation touches 5+ files or 3+ components:
- Warn: "This is getting large for a quick-dev. Consider `/ralph` for a full agent team implementation."
- Ask: "Continue anyway, or switch to `/ralph`?"

## Stage 3: Implement

For each task (in order if from a spec):

1. **Read** the target file(s) to understand current state
2. **Implement** the change following existing patterns and conventions
3. **Write tests** for the change
4. **Verify** the implementation meets acceptance criteria

**Rules**:
- Follow existing code patterns and conventions
- Write tests alongside code, not as an afterthought
- One logical change per step — don't mix unrelated changes
- If a task reveals unexpected complexity, stop and inform the user

## Stage 4: Self-Check

After all tasks are complete:

1. **Run tests** — ensure all tests pass (new and existing)
2. **Check acceptance criteria** — verify each Given/When/Then scenario
3. **Review changes** — scan all modified files for obvious issues

Report results:
```
## Implementation Complete

**Files created**: {list}
**Files modified**: {list}
**Tests added**: {count}

### Acceptance Criteria
- [x] {criterion 1} — PASS
- [x] {criterion 2} — PASS

### Issues Found
- {any issues, or "None"}
```

## Stage 5: Next Steps

Suggest:
- "Run `/review` for a code review"
- "Run `/commit-msg` to commit these changes"
- If issues were found: "Fix {issue} before committing"
