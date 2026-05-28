---
name: spec
description: Crystallize an executable spec.md from problem + architecture (Hybrid SDD/BMAD workflow)
---

Produce `.claude/output/spec.md` — the **executable specification** that locks the implementation contract before `/bmad-act` runs.

This command is meaningful only when the project is in **Hybrid SDD/BMAD mode** (the rule `.claude/rules/spec-driven.md` is installed). In pure BMAD mode it can still run, but the gate in `/bmad-act` will not enforce its output.

## Prerequisites

Read these files. If any is missing, name it and stop.

1. `.claude/output/problem.md` (from `/bmad-break`)
2. `.claude/output/architecture.md` (from `/bmad-model`)
3. `.claude/output/acceptance-criteria.md` if it exists (from `/bmad-break` or `/clarify`)

## Output structure

Write `.claude/output/spec.md` with **exactly these sections**. Do not add extras. Do not invent. If a section has no content because the prior artifacts don't cover it, write the section header followed by `> open question — see spec-questions.md`.

### 1. Identity
- Binary or service name
- Language and minimum version
- Dependencies (or "stdlib only")
- Entry point (file path)
- Persistence/storage choice

### 2. Public surface (exhaustive)
A complete list of commands, endpoints, exported functions, or events. Number them. If a name is not here, the implementation must not create it.

```
S1  shorty add <url>
S2  shorty get <code>
...
```

### 3. Data model
Types (with field names + types) and persistence format. Use code blocks. No prose alternatives.

### 4. Cross-cutting rules
- Error code policy (exit codes for CLI, HTTP statuses for service)
- Output discipline (stdout vs stderr split, formatting)
- Atomicity invariants
- Concurrency expectations

### 5. File layout
Exact tree the implementation MUST produce. Files not listed = drift.

```
project/
├── go.mod
├── cmd/<name>/main.go
├── internal/<pkg>/<file>.go (+ _test.go)
└── tests/acceptance_test.go
```

### 6. Test naming convention
Single line stating the convention. Example: `TestAC_NN_<short_desc>` so each AC maps 1:1 to a test function. Include the count expected (one per AC).

### 7. AC → test mapping
Table with two columns: AC ID, test function name. Built from `acceptance-criteria.md`. If acceptance criteria are not yet in EARS format, convert them in-place (no behavior change, just structure: "When X, the system shall Y").

## Open questions

If during generation you find ambiguity in any prior artifact, append a new entry to `.claude/output/spec-questions.md` (create the file if absent). Format:

```markdown
## Q<N> — <short title>
**Context**: <what the prior artifact said>
**Question**: <what's ambiguous>
**Conservative interpretation chosen**: <the minimal, least-surprising option>
**Recommended spec amendment**: <what would resolve the ambiguity>
```

Do NOT silently choose. Surface every ambiguity.

## Verification

After writing `spec.md`:
1. Confirm all 7 sections are present
2. Confirm the public surface in section 2 is numbered and exhaustive
3. Confirm the file layout in section 5 contains no `...` placeholders
4. If `spec-questions.md` is non-empty, list each question to the user and recommend resolving them before `/bmad-act`

Report back:
- Path: `.claude/output/spec.md`
- Sections present: 7/7 (or which are missing/marked open)
- Open questions count
- Recommended next step
