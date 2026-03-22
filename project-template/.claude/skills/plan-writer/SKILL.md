---
name: plan-writer
description: Decompose a story into ultra-granular implementation tasks (2-5 minutes each) with exact file paths, complete code, and verification commands. Use for stories estimated at more than 50 lines of code. Includes a plan-reviewer subagent for validation.
disable-model-invocation: true
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "[story ID or description]"
---

### Granular Plan Writer

You produce implementation plans where each task is a single action taking 2-5 minutes. The plan must be detailed enough for "an enthusiastic junior with no taste or judgment" to execute correctly.

### Plan Structure

#### 1. File Structure Map (mandatory, before tasks)

List ALL files that will be created, modified, or tested:

```markdown
| # | File | Action | Responsibility |
|---|------|--------|----------------|
| 1 | src/services/auth.ts | Create | Authentication service |
| 2 | src/services/auth.test.ts | Create | Auth service tests |
| 3 | src/routes/login.ts | Modify | Add login endpoint |
```

#### 2. Tasks (ordered, with checkboxes)

Each task follows this template:

```markdown
- [ ] **Step N: {action description}**
  - **Files**: {Create|Modify|Test} `path/to/file`
  - **Code**:
    ```{language}
    {COMPLETE code — no placeholders, no "add logic here", no TODOs}
    ```
  - **Verify**: `{exact command to run}`
  - **Expected**: {what success looks like — test output, HTTP status, etc.}
```

#### 3. Task Granularity Rules

- Each task = ONE action (write a test, implement a function, run tests, commit)
- Maximum 2-5 minutes per task
- "Write the failing test" and "implement the code" are SEPARATE tasks
- "Run tests" is its own task after every implementation step
- Code blocks must be COMPLETE — copy-pasteable without modification
- Verification commands must be exact (not "run the tests" but `npm test -- --grep "auth"`)

#### 4. TDD Task Ordering

For code stories, tasks MUST follow RED-GREEN-REFACTOR:
1. Write failing test (RED)
2. Run test, confirm failure (VERIFY RED)
3. Write minimal implementation (GREEN)
4. Run test, confirm pass (VERIFY GREEN)
5. Refactor if needed (REFACTOR)
6. Run tests again (VERIFY REFACTOR)

#### 5. Plan Validation

After writing the plan, dispatch a plan-reviewer subagent to validate it (see plan-reviewer-prompt.md). Maximum 3 iterations of review.

### When NOT to Use Granular Plans

Stories estimated at 50 lines or less use the standard plan format (files + approach + criteria mapping). The lead decides the threshold.
