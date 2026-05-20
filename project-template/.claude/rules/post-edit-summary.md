---
description: After each batch of edits, summarize touched files and their role in the system
globs: ["**/*"]
---

## After-Edit Infrastructure Summary (non-negotiable)

After every batch of edits in a single response (one or more `Edit` / `Write` operations producing related changes), produce a short summary that helps the reader hold a mental map of the system.

A "batch" = the edits made before the next user turn, on the same logical concern. One summary per batch — not one per tool call.

### What the summary must contain

For each file changed in the batch, give:

- **Path** from repo root
- **Role** — what this file IS in the architecture (e.g. "service entry point", "config loader", "CI gate", "slash command definition")
- **This change** — one short line: what changed and why

A compact table is the right format:

| File | Role | This change |
|------|------|-------------|
| `cmd/server/main.go`        | HTTP entry point             | Wired the new `/healthz` handler         |
| `internal/auth/jwt.go`      | Token issuance + validation  | Rotated signing-key load path            |
| `.github/workflows/ci.yml`  | CI gate: lint + test         | Bumped Go toolchain to 1.22              |

### When to skip the summary

- One-character or whitespace-only edits
- Pure renames with no content change
- Documentation typo fixes
- Anything where the path itself already says everything (e.g. a single test rename)

If in doubt: include it. The cost of an extra table is low; the cost of losing the user's mental model is high.

### Why this matters

Edits accumulate fast. After 3–4 file touches, the reader loses track of "where did we change things and how does it shape the system." A 30-second summary:

- Keeps the user's mental model fresh
- Surfaces cross-cuts (a change touching service + proxy + config = a coordination smell)
- Doubles as raw material for the PR description
- Forces the editor (you) to step back and check that the changes actually belong together

### Connection to other rules

- This is finer-grained than the end-of-turn summary; one summary per edit batch, not per turn
- Feeds the PR description verbatim — keep wording terse and copy-paste-ready
- Pairs with `verification.md`: the summary states what changed; verification states what was tested
