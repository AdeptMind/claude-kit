# BMAD Project Template

## Intent Detection — Read This First

When the user sends a message **without a slash command**, default to `/ralph`. Ralph handles everything — from a vague description to a full backlog. This is the most common path.

**Exceptions** — override the Ralph default when the user's intent clearly matches one of these:

| Signal | Action |
|--------|--------|
| "go", "run it", "full workflow", "bmad" | Run `/bmad-run` |
| "where am I", "what's next", "status", "help", "dashboard" | Run `/bmad-help` |
| Asks about principles, standards, governance | If `.claude/output/principles.md` exists: **read and display it**. If not: run `/principles` to create it. |
| Questions requirements ("unclear", "what about", "edge case") | If `problem.yaml` exists: run `/clarify`. If not: run `/bmad-break`. |
| "Check specs", "ready to code?", "anything missing" | If all 3 artifacts exist: run `/analyze`. If no CRITICAL issues, then run `/checklist`. If artifacts missing: say which phases to run. |
| "debate", "discuss", "opinions on", "argue" | Run `/party` |
| "quick fix", "small feature", "just do", "fast track" | Run `/quick-spec` or `/quick-dev` depending on whether the user wants a spec or direct implementation |
| "brainstorm", "ideas for", "what if", "ideate" | Run `/brainstorm` |
| "split this doc", "shard", "break up this file" | Run `/shard` |
| "UX", "wireframe", "user flow", "screens", "user experience" | Run `/ux-spec` |
| "create agent", "new command", "new skill", "new rule" | Run `/create-component` |
| Asks for code review, test gen, security check | Run `/review`, `/test-gen`, or `/security-check`. |

**When the user uses a slash command** (`/bmad-break`, `/ralph`, etc.): follow that command exactly. The default-to-Ralph rule only applies when there is NO slash command.

## Workflow

- Full workflow: `/bmad-run` (Brainstorm → Principles → Break → Clarify → UX Spec → Model → Analyze → Checklist → GSD Prep → Act → Deliver)
- Individual phases: `/bmad-break`, `/bmad-model`, `/bmad-act`, `/bmad-deliver`
- Spec & quality: `/principles`, `/clarify`, `/analyze`, `/checklist`, `/ux-spec`
- Implementation: `/ralph`, `/ralph-loop`, `/gsd-prep`, `/quick-spec`, `/quick-dev`
- Ideation: `/brainstorm`, `/party`
- Utilities: `/shard`, `/create-component`, `/bmad-help`

## Command Reference

### Standalone commands (work from just a prompt, no prior artifacts needed)

| Command | What it does | Input |
|---------|-------------|-------|
| `/principles` | PO vs TL debate to define project governance | Codebase scan + interactive |
| `/bmad-break` | Define problem, features, rich user stories | Project brief or prompt |
| `/clarify` | Structured ambiguity scan | `problem.yaml` OR a project description as argument |
| `/ralph` | Full implementation — backlog to code | Backlog file, `backlog.yaml`, OR a text description |
| `/brainstorm` | Creative ideation with structured techniques | Topic or challenge |
| `/party` | Multi-agent debate on a topic | Topic to debate |
| `/quick-spec` | Lightweight tech spec — fast track to tasks | Feature description |
| `/quick-dev` | Quick implementation from spec or description | Spec file path OR feature description |
| `/shard` | Split large documents into indexed sections | File path |
| `/create-component` | Create a new agent, command, skill, or rule | `<type> <name> [description]` |
| `/bmad-help` | Project state dashboard and next steps | Optional natural language question |

### Pipeline commands (require prior BMAD artifacts)

| Command | What it does | Requires |
|---------|-------------|----------|
| `/bmad-model` | Architecture + backlog design | `problem.yaml` |
| `/ux-spec` | UX design — flows, wireframes, interactions | `problem.yaml` OR feature description |
| `/analyze` | Cross-artifact consistency check (read-only) | `problem.yaml` + `architecture.yaml` + `backlog.yaml` |
| `/checklist` | Pre-implementation quality gate | `problem.yaml` + `architecture.yaml` + `backlog.yaml` |
| `/gsd-prep` | Codebase mapping + context packs for teammates | `backlog.yaml` + `architecture.yaml` |
| `/bmad-act` | Implement from backlog (delegates to Ralph) | `backlog.yaml` + `architecture.yaml` |
| `/bmad-deliver` | Prepare release | Implemented code |
| `/ralph-loop` | Resume interrupted Ralph session | `.claude/ralph-prd.json` |

### Shortcuts

| Shortcut | Expands to |
|----------|-----------|
| `/r` | `/ralph` |
| `/p` | `/principles` |
| `/c` | `/clarify` |
| `/g` | `/gsd-prep` |
| `/h` | `/bmad-help` |
| `/qs` | `/quick-spec` |
| `/qd` | `/quick-dev` |

## Approach Selection

For any non-trivial implementation (new feature, infra change, refactor):
1. **Scan first**: read how similar things are done in the codebase before deciding on an approach
2. **Propose before implementing**: if there are multiple valid approaches, present 2-3 options with their tradeoffs and wait for the user to choose — do not pick one silently
3. **Prefer existing over new**: reuse existing patterns, modules, and dependencies before introducing new abstractions or tools
4. **Targeted over broad**: when the user asks about a specific thing, answer that specific thing — do not expand scope unless explicitly asked

## Code Principles

All code produced by any agent or skill MUST follow these principles:

- **DRY** (Don't Repeat Yourself): never duplicate logic; extract shared code into reusable functions, modules, or classes
- **KISS** (Keep It Simple): prefer the simplest solution that works; avoid unnecessary abstractions and complexity
- **SOLID**: Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion
- **Least invasive**: make the smallest change possible; do not refactor, rename, or reorganize code outside the scope of the task
- **No over-engineering**: do not add features, configurability, or abstractions that are not explicitly required
- **Separation of concerns**: keep business logic, data access, presentation, and infrastructure in distinct layers
- **Clean code**: descriptive naming, small functions, no dead code, no commented-out code, no magic numbers

## Docs Index

- Stack-specific guidance is in `.claude/docs-index.md` — **read it at session start** for framework conventions
- The docs-index is auto-generated by `ck docs` based on detected project dependencies
- Staleness metadata is in `.claude/.docs-meta.json` (dependency hash + timestamp)
- At session start, check if docs-index is stale: run `/ck-sync` or tell the user to run `ck docs --refresh`
- Prefer retrieval-led reasoning: read the docs-index before making framework-specific decisions

## Build & Test

- Adapt commands below to your stack (Node.js, Python, Go, etc.)
- Build: `npm run build` (or equivalent)
- Test: `npm test` (or equivalent)
- Lint: `npm run lint` (or equivalent)
- After modifying functions, run `/test-check` to verify test coverage and honesty

## Code Style

- Follow language/framework conventions of the project
- Keep functions small and focused on a single responsibility
- Name variables and functions descriptively
- Prefer composition over inheritance when possible
- See @.claude/rules/code-style.md for detailed rules

## Git Workflow

- Use conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`
- **All commit messages and PR titles/descriptions must be written in English** — no exceptions regardless of the project's spoken language
- Before pushing, verify the remote is configured: `git remote -v`
- Before committing, review staged files: `git diff --cached --name-only` — never include files outside the current story or task scope
- Run `/commit-msg` to generate commit messages from staged changes
- Run `/review` before merging PRs

## Security

- Never hardcode secrets in code or config files
- Run `/security-check` before releases
- Run `/pentest` for web application security validation
- See @.claude/rules/security.md for security rules

## Context Management

- When context usage reaches the threshold defined by `CONTEXT_COMPACT_THRESHOLD` (default: 75%), proactively run `/compact` to free up context space before continuing work.
- The threshold is configured in `.claude/settings.json` under `env.CONTEXT_COMPACT_THRESHOLD`.

## Architecture

- See `.claude/agents/` for available agent roles
- See `.claude/skills/` for available skills
- Output artifacts are written to `.claude/output/`
