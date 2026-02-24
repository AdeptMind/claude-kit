---
name: bmad-help
description: Project state dashboard — show progress, artifacts, and recommended next steps
---

You are a **project navigator**. Scan the project state and present a clear dashboard showing where the user is in the BMAD workflow, what's done, what's missing, and what to do next.

## Stage 1: Scan Project State

Check for the existence and content of these artifacts:

**BMAD artifacts** (in `.claude/output/`):
- `principles.md` — Project principles (Phase 0)
- `problem.yaml` — Problem definition (Phase 1: Break)
- `architecture.yaml` — Architecture design (Phase 2: Model)
- `backlog.yaml` — Implementation backlog (Phase 2: Model)
- `checklist.md` — Pre-implementation readiness (Phase 2.3)
- `gsd/prep-report.md` — GSD preparation (Phase 2.5)
- `gsd/gap-analysis.md` — Gap analysis (Phase 2.5)
- `gsd/context-packs/` — Context packs (Phase 2.5)
- `act-report.md` — Implementation report (Phase 3: Act)
- `release-notes.md` — Release notes (Phase 6: Deliver)
- `brainstorm-*.md` — Brainstorm sessions
- `party-*.md` — Party debate sessions
- `quick-spec-*.md` — Quick specs
- `ux-spec.md` — UX specification
- `shards/` — Sharded documents

**Session state**:
- `.claude/ralph-prd.json` — Check for interrupted Ralph sessions

**Existing code** — scan project root for:
- `src/`, `lib/`, `app/`, `cmd/`, `pkg/`, `internal/` directories
- `main.*`, `index.*`, `app.*` entry points
- `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml` — detect stack
- Test files and test directories

## Stage 2: Determine Current Phase

Based on artifacts found, determine the furthest completed phase:

| Phase | Required Artifact | Status |
|-------|------------------|--------|
| Phase -1: Brainstorm | `brainstorm-*.md` | Optional |
| Phase 0: Principles | `principles.md` | Optional |
| Phase 1: Break | `problem.yaml` | Required |
| Phase 1.5: Clarify | clarifications in `problem.yaml` | Auto |
| Phase 1.75: UX Spec | `ux-spec.md` | Optional |
| Phase 2: Model | `architecture.yaml` + `backlog.yaml` | Required |
| Phase 2.25: Analyze | (read-only check) | Auto |
| Phase 2.3: Checklist | `checklist.md` | Required |
| Phase 2.5: GSD Prep | `gsd/prep-report.md` | Required |
| Phase 3: Act | `act-report.md` | Required |
| Phase 6: Deliver | `release-notes.md` | Required |

## Stage 3: Display Dashboard

Present the dashboard in this format:

```
## Project Status

**Stack**: [detected stack or "not detected"]
**Current Phase**: [phase name]
**Next Step**: [recommended action]

### Artifacts

| Phase | Artifact | Status |
|-------|----------|--------|
| ... | ... | Done / Missing / N/A |

### Interrupted Sessions
[If ralph-prd.json exists: "Ralph session interrupted — resume with /ralph-loop"]
[If none: "No interrupted sessions"]

### Recommended Next Step
[Based on state, recommend ONE clear action]
```

## Stage 4: Recommend Next Step

Based on the current state:

- **No artifacts at all**: "Start with `/bmad-break` to define your problem, or `/ralph` for direct implementation."
- **Only `problem.yaml`**: "Run `/bmad-model` to design architecture and backlog."
- **`problem.yaml` + `architecture.yaml` + `backlog.yaml`**: "Run `/analyze` and `/checklist` to verify readiness, then `/gsd-prep`."
- **All prep done, no act-report**: "Run `/ralph` to start implementation."
- **`act-report.md` exists**: "Run `/bmad-deliver` to prepare the release."
- **Ralph interrupted**: "Resume with `/ralph-loop`."
- **Everything done**: "Project complete! Consider running `/review` or `/security-check` for final validation."

## Stage 5: List Available Commands

Group commands by category:

**Full Workflow**: `/bmad-run`
**Planning**: `/bmad-break`, `/clarify`, `/principles`, `/brainstorm`
**Design**: `/bmad-model`, `/ux-spec`, `/analyze`, `/checklist`
**Implementation**: `/gsd-prep`, `/ralph`, `/ralph-loop`, `/quick-spec`, `/quick-dev`
**Quality**: `/review`, `/test-gen`, `/security-check`
**Utilities**: `/party`, `/shard`, `/create-component`
**Release**: `/bmad-deliver`

If $ARGUMENTS is provided, interpret it as a natural language question and answer accordingly: $ARGUMENTS
