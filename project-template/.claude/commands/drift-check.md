---
name: drift-check
description: Compare actual code to spec.md — detect extra commands, missing files, test mismatches (Hybrid SDD/BMAD workflow)
---

Compare the implementation against `.claude/output/spec.md` on three mechanical axes. Produce `.claude/output/drift-check.md` and report blockers loudly.

This is a **read-only** check. Do not modify code. Do not modify the spec. The user decides which side to fix.

## Prerequisites

1. Read `.claude/output/spec.md`. If absent, stop and tell the user to run `/spec` first.
2. Find the implementation directory:
   - If `$ARGUMENTS` is provided, use it as the code path
   - Else infer from section 5 of the spec (look at top-level entry like `cmd/...`)
   - If still unclear, ask the user

## Three checks

### Check 1 — Public surface (commands / endpoints)

Extract the public surface list from spec section 2. Scan the code to enumerate what actually exists:

- **CLI**: look at the dispatch in `main.go` or `cli.go` — switch/case branches on `os.Args[1]` or equivalent
- **HTTP service**: look at route registrations (`mux.HandleFunc`, `r.Route`, etc.)
- **Library**: look at exported symbols (`go doc`, `grep '^func [A-Z]'`)

Compare. Two failure modes:
- **Extra** in code, not in spec → drift, severity BLOCKER (silent feature creep)
- **Missing** in code, listed in spec → drift, severity BLOCKER (unfinished)

### Check 2 — File layout

Extract the file tree from spec section 5. Run `find . -name '*.go' | sort` (or the language equivalent) in the implementation root. Compare:

- **Extra file** present, not in spec → drift, severity MINOR (might be a helper the spec failed to declare) or BLOCKER (if it's a new package not anticipated)
- **Missing file** listed in spec → drift, severity BLOCKER

### Check 3 — Test count and naming

Extract the test naming convention from spec section 6 and the expected count (one per AC, count from section 7 mapping).

Run `grep -c '^func Test' tests/` (adjust per language). Compare:

- **Test count mismatch** → drift, severity BLOCKER if fewer, MINOR if more
- **Test name does not match convention** → drift, severity MINOR

## Output format

Write `.claude/output/drift-check.md`:

```markdown
# Drift Check — <timestamp>

## Summary
- Public surface: <PASS|N drift(s)>
- File layout: <PASS|N drift(s)>
- Test count: <PASS|N drift(s)>

## Findings

### Public surface
| Severity | Type | Spec says | Code has | Notes |
|---|---|---|---|---|
| BLOCKER | extra | — | `shorty stats` | not in spec section 2 |

### File layout
| Severity | Type | Path | Notes |
|---|---|---|---|

### Tests
| Severity | Type | Found | Expected | Notes |
|---|---|---|---|---|

## Verdict
**<CLEAN|N MINOR drifts|N BLOCKERS>**

If BLOCKERS: list the choice — update spec OR update code — for each one. Do not act, just describe the choice.
```

## Report back to user

- Verdict line first (clean / N minor / N blockers)
- If blockers: enumerate them with the file:line where the drift was detected
- Recommend the next action: `/spec` to update spec, or code fix

Exit posture: if there are blockers, say so explicitly. Do not soften. The point of the drift check is to make divergence visible — not to negotiate it away.
