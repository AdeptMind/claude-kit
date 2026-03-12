# Claude Kit (ck)

> **Ship production-quality features, not scaffolding.**

Claude Code is powerful — but out of the box, every project starts from scratch: no agent roles, no workflow, no guardrails. Teams waste time re-inventing prompts, fighting hallucinations, and babysitting agents that drift off-scope.

`claude-kit` solves this. It gives Claude a **structured workflow** (BMAD), a **team of specialists** (backend, devops, security, finops, and more), and **guardrails** that prevent the most common failure modes — before you write a single line of code.

**What you get:**
- **Agent team** — 20+ role-optimized agents (architect, tech-lead, security, golang, terraform, k8s…) that activate automatically in the right context, each with the right model and skill set
- **BMAD workflow** — a phased approach (Break → Model → Act → Deliver) with structured gates that take a project from idea to shipped feature with minimal drift
- **Ralph** — an autonomous team lead that parallelizes implementation across agents, coordinates contracts, and validates each story before moving on
- **Guardrails** — rules that prevent manual cloud mutations, pattern-first coding, honest test pairing, and pre-commit scope visibility
- **Smart discovery** — `ck add new "go microservices"` asks Claude to find the right agent from local templates, upstream repos, or the broader ecosystem

One `ck init` turns an empty project into a Claude-powered team with a coherent workflow.

> *"As soon as the meeting finished I was already installing it in my repo. It really worked to the point of solving some issues I haven't even spotted in my code."*
> — **Tainã Martinez**

---

A Go CLI for managing Claude Code project templates — interactive TUI setup, component management, stack-aware docs generation, and template synchronization.

Built with [Charm](https://charm.sh): Bubble Tea + huh + lipgloss.

---

## Quick Start

### Install via Homebrew

```bash
brew tap adrien-barret/claude-kit
brew install claude-kit
```

### Install from source

```bash
cd claude-cli
make install
```

This builds the `claude-kit` binary (aliased as `ck`), copies it to `/usr/local/bin`, and installs templates to `~/.bmad/templates/`.

### Initialize a project

```bash
cd my-project

# Interactive TUI — pick components from a categorized multi-select
ck init

# AI-guided setup — Claude recommends components based on your project
ck init --plan

# Install to global ~/.claude
ck init --global
```

### Add agents interactively

```bash
# Interactive agent picker — auto-installs skills + rules
ck add

# Add specific agents by name (with their deps)
ck add backend devops

# Add a specific component type
ck add skill code-reviewer
ck add command review
ck add rule testing
```

### Other commands

```bash
ck list                      # See available vs installed components
ck remove                    # Interactive removal picker
ck remove backend            # Remove an agent
ck sync                      # Update installed components from templates
ck docs                      # Generate stack-aware docs-index.md
```

---

## CLI Reference

| Command | Description |
|---------|-------------|
| `ck init` | Interactive setup — categorized multi-select of components |
| `ck init --plan` | AI-guided setup via Claude session |
| `ck init --global` | Install to `~/.claude` |
| `ck add` | Interactive agent picker (auto-installs skills + rules) |
| `ck add <name> [name...]` | Add agents by name with their dependencies |
| `ck add <type> <name>` | Add a specific component (skill, command, rule) |
| `ck remove` | Interactive removal picker |
| `ck remove <name>` | Remove an agent |
| `ck remove <type> <name>` | Remove a specific component |
| `ck list` | Available vs installed side-by-side table |
| `ck list --available` | Available components only |
| `ck list --installed` | Installed components only |
| `ck sync` | Update installed components + refresh docs-index |
| `ck docs` | Generate docs-index.md via stack detection |
| `ck docs --refresh` | Force regenerate even if fresh |
| `ck dep install` | Install recommended dependencies interactively |
| `ck profile list\|use\|add\|remove` | Manage Claude account profiles |
| `ck teammate-mode` | View or change teammate display mode |
| `ck version` | Print version |

### Eval & Quality

| Command | Description |
|---------|-------------|
| `ck skill eval <skill-dir>` | Test trigger accuracy against evals.json queries (`--model`, `--workers`, `--runs`, `--threshold`) |
| `ck skill optimize <skill-dir>` | Eval→improve loop to optimize skill descriptions (`--max-iterations`, `--train-ratio`, `--report`) |
| `ck skill grade <skill-dir>` | Grade output against grading.json assertions (`--output-file` required, `--model`) |
| `ck skill benchmark <results-dir>` | Aggregate grading stats, compare with/without skill (`--output`) |
| `ck skill validate <skill-dir>` | Validate skill structure and frontmatter |
| `ck skill report <results.json>` | Generate interactive HTML eval report (`-o`, `--previous`, `--open`) |

### Packaging & Distribution

| Command | Description |
|---------|-------------|
| `ck skill package <skill-dir>` | Package a skill as .skill archive (`-o`, `--skip-validation`) |
| `ck agent validate <agent.md>` | Validate agent frontmatter, skill refs, and tools |
| `ck agent package <agent.md>` | Package an agent as .agent archive (`-o`, `--skip-validation`) |
| `ck agents registry` | Generate agent-registry.yaml with collaboration maps (`--update`) |
| `ck package <template-dir>` | Bundle full template as .claude-kit archive (`-o`) |
| `ck install <archive>` | Install .skill, .agent, or .claude-kit archive (`--force`) |

### BMAD Eval

| Command | Description |
|---------|-------------|
| `ck bmad eval [output-dir]` | Evaluate BMAD artifacts against phase assertions (`--phase`, `--model`) |
| `ck bmad benchmark <run1> <run2>` | Compare two BMAD eval runs |

### How `add` works

When you add an agent, it automatically installs:
- The agent definition itself
- All skills listed in the agent's frontmatter
- Related rules based on the agent's role (e.g. backend → code-style, testing, security, api)

```bash
ck add security
# ✓ Added agent: security
#   Installing 4 skill dependencies:
#     ✓ Added skill: security/code-security-audit
#     ✓ Added skill: security/infra-security-audit
#     ✓ Added skill: security/auth-review
#     ✓ Added skill: security/secret-rotation
#   Installing 1 related rules:
#     ✓ Added rule: security
```

### Component types

For explicit type prefixes (`ck add <type> <name>`):

- `agent` / `agents`
- `skill` / `skills`
- `command` / `commands`
- `rule` / `rules`

---

## What's Included

| Category | Components |
|----------|-----------|
| **BMAD Workflow** | Principles → Break → Clarify → Model → Analyze → Checklist → GSD Prep → Act → Deliver |
| **Agents** | Backend, Tech Lead, DevOps, Security, Pentester, FinOps |
| **Dev Skills** | Code review, test generation, **test-check**, API docs, commit helper, README updater, dependency audit |
| **Security Skills** | Code audit, infra audit, auth review, secret rotation, pentest simulation, threat modeling |
| **FinOps Skills** | Cost optimization, tagging audit, waste detection, budget forecasting |
| **Other Skills** | Performance audit, accessibility audit, database review, Terraform review, skill creator |
| **Rules** | Code style (pattern-first, placement discovery), testing (function-test pairing), security, API, frontend, infrastructure (IaC-only), documentation, FinOps |

### Slash Commands

**BMAD Workflow:**
`/bmad-run`, `/bmad-break`, `/bmad-model`, `/bmad-act`, `/bmad-deliver`

**Spec & Quality Gates:**
`/principles`, `/clarify`, `/analyze`, `/checklist`

**Implementation:**
`/ralph`, `/ralph-loop`, `/ralph-cancel`, `/gsd-prep`

**Dev Skills:**
`/review`, `/pr-review`, `/test-gen`, `/test-check`, `/docs-gen`, `/commit-msg`, `/code-only`

**Security & FinOps:**
`/security-check`, `/pentest`, `/cost-review`

**Roles:**
`/role-backend`, `/role-tech-lead`, `/role-devops`, `/role-security`, `/role-pentester`, `/role-finops`

**Utilities:**
`/ck-sync`

---

## Docs Index

The docs-index system generates compressed, stack-specific notes that stay in Claude's context.

### How it works

1. `ck docs` scans your project root for dependency files (package.json, go.mod, requirements.txt, etc.)
2. Detects your tech stack (languages, frameworks, tools)
3. Generates `.claude/docs-index.md` with framework-specific directives
4. Stores metadata in `.claude/.docs-meta.json` for staleness tracking

### Auto-sync

The docs-index is considered stale when:
- Dependency files have changed (hash mismatch)
- More than 14 days since last generation

`ck sync` automatically refreshes the docs-index after updating components.

### Supported stacks

Languages: JavaScript, TypeScript, Python, Go, Ruby, Rust, Java, Kotlin, PHP
Frameworks: Next.js, React, Vue, Nuxt, Svelte, Angular, Express, Fastify, NestJS, Hono, Django, Flask, FastAPI, Rails, Sinatra, Laravel, Symfony
Tools: Docker, Terraform, Kubernetes, Helm, GitHub Actions, Prisma, Drizzle, Tailwind

---

## Build & Development

### Prerequisites

- Go 1.21+
- Make

### Build

```bash
cd claude-cli
make build              # Compile binary to ./claude-kit
make install            # Build + copy to /usr/local/bin (+ ck alias) + install templates
make install-templates  # Copy templates to ~/.bmad/templates/ only
make clean              # Remove build artifacts
make uninstall          # Remove binary, alias, and templates
```

### Template directory resolution

The binary resolves the template directory in this order:
1. `$BMAD_TEMPLATE_DIR` environment variable
2. `~/.bmad/templates/` (installed via `make install-templates`)
3. Adjacent `project-template/.claude/` (for development from source)

### Go dependencies

- [cobra](https://github.com/spf13/cobra) — subcommand structure
- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [huh](https://github.com/charmbracelet/huh) — forms, multi-select, confirm dialogs
- [lipgloss](https://github.com/charmbracelet/lipgloss) — styling, tables, colors

---

## Legacy Installer

The `install.sh` script still works as a fallback. If `ck` / `claude-kit` is available, it delegates automatically:

```bash
# These are equivalent:
bmad-setup --plan           →  ck init --plan
bmad-setup --global         →  ck init --global
bmad-setup                  →  ck init
```

If `ck` is not installed, `install.sh` falls back to the original bash-based installer.

---

## Project Structure

```
claude-cli/
├── cmd/claude-kit/         # Go CLI source
│   ├── main.go             # Cobra root command + version
│   ├── init.go             # ck init — huh multi-select + --plan mode
│   ├── add.go              # ck add — interactive agent picker + auto-deps
│   ├── remove.go           # ck remove — interactive removal + warnings
│   ├── list.go             # ck list — lipgloss table
│   ├── sync.go             # ck sync — update + docs refresh
│   ├── docs.go             # ck docs — stack detection + generation
│   ├── eval.go             # ck skill eval — trigger accuracy testing
│   ├── optimize.go         # ck skill optimize — eval→improve loop
│   ├── grade.go            # ck skill grade — LLM grading
│   ├── benchmark.go        # ck skill benchmark — stats aggregation
│   ├── report.go           # ck skill report — HTML report generation
│   ├── validate.go         # ck skill/agent validate
│   ├── package.go          # ck skill/agent/template package
│   ├── install.go          # ck install — archive installation
│   ├── registry.go         # ck agents registry
│   └── bmadeval.go         # ck bmad eval/benchmark
├── internal/
│   ├── catalog/            # Template scanning + component operations
│   ├── stack/              # Stack detection from dependency files
│   ├── docsindex/          # Docs-index generation + staleness
│   ├── config/             # Path resolution + defaults
│   ├── eval/               # Eval runner, trigger detection, improve loop
│   ├── claude/             # Claude subprocess wrapper
│   ├── grading/            # LLM grading engine
│   ├── improve/            # Description improvement with retry
│   ├── benchmark/          # Stats aggregation + delta comparison
│   ├── report/             # HTML report generation
│   ├── packaging/          # Skill/agent/template packaging + install
│   └── bmadeval/           # BMAD phase assertions
├── project-template/.claude/  # Template files
│   ├── CLAUDE.md           # Project memory + approach-selection guardrails
│   ├── settings.json       # Permissions + PreToolUse hook (staged-file review)
│   ├── agents/             # 15 agent role definitions
│   ├── skills/             # 25+ skill directories (incl. test-check)
│   ├── commands/           # 53 slash commands
│   └── rules/              # 9 project rules
├── go.mod / go.sum
├── Makefile                # build, install, install-templates, clean
├── install.sh              # Legacy wrapper → delegates to ck
├── prompts.sh              # AI-guided setup prompts (used by --plan)
└── README.md
```

---

## Skills Reference

### Dev Skills

| Skill | Description |
|-------|-------------|
| `code-reviewer` | Code review with severity levels (critical/warning/info), auto-fix suggestions |
| `test-generator` | Test generation with framework detection and coverage gap analysis |
| `test-check` | Per-function test coverage: finds or creates tests, updates them when contracts change, reports failures honestly — never weakens assertions to hide bugs |
| `api-documenter` | OpenAPI/Swagger documentation generation |
| `git-commit-helper` | Conventional commit message generation |
| `readme-updater` | Keep README in sync with code |
| `dependency-auditor` | Vulnerability scanning, license compatibility matrix, supply-chain risk scoring |

### Security Skills

| Skill | Description |
|-------|-------------|
| `security` | Orchestrator — runs all security sub-skills |
| `security/code-security-audit` | OWASP Top 10, injection, XSS, hardcoded secrets |
| `security/infra-security-audit` | Cloud config, permissions, encryption |
| `security/auth-review` | OAuth/JWT, RBAC, token policies |
| `security/secret-rotation` | Secret storage and rotation policies |
| `security/pentest-web` | Auth bypass, IDOR, SSRF, rate-limit bypass, JWT attacks |
| `security/threat-model` | STRIDE threat modeling |

### FinOps Skills

| Skill | Description |
|-------|-------------|
| `finops` | Orchestrator — runs all FinOps sub-skills |
| `finops/cost-optimization` | Rightsizing, auto-scaling, reserved instances |
| `finops/tagging-audit` | Cost allocation tag compliance |
| `finops/waste-detection` | Idle resources, orphaned volumes |
| `finops/budget-forecast` | Cost estimation from IaC |

### New Skills

| Skill | Description |
|-------|-------------|
| `performance-audit` | N+1 queries, bundle size, caching, lazy loading, connection pooling |
| `accessibility-audit` | WCAG 2.1 AA, ARIA, keyboard nav, contrast, screen reader |
| `database-review` | Schema, indexing, query optimization, migration safety |
| `terraform-review` | Module structure, state management, provider versioning |
| `skill-creator` | Meta-skill to generate new SKILL.md files |

---

## How to Describe Your Application

To use the BMAD workflow, provide a **project brief**:

```
1. PROJECT NAME
2. PROBLEM STATEMENT — what it solves, who it's for
3. TECH STACK — language, framework, database, cloud
4. CORE FEATURES — prioritized list
5. CONSTRAINTS — performance, compliance, multi-tenancy
6. INTEGRATIONS — external APIs, payment, notifications
7. INFRASTRUCTURE — deployment, CI/CD, containers
```

Then run `/bmad-run` for the full workflow, or phase by phase:

```bash
/principles       # (optional) PO vs TL debate → principles.md
/bmad-break       # Define the problem with rich user stories → problem.yaml
/clarify          # Structured ambiguity scan → updates problem.yaml
/bmad-model       # Design architecture → architecture.yaml, backlog.yaml
/analyze          # Cross-artifact consistency check (read-only)
/checklist        # Pre-implementation quality gate → checklist.md
/gsd-prep         # Codebase mapping + context packs for teammates
/ralph            # Agent team implementation with numbered branches
/bmad-deliver     # Prepare release → release-notes.md
```

### Standalone vs pipeline commands

Some commands work from just a prompt — no prior artifacts needed:

| Command | Input |
|---------|-------|
| `/principles` | Codebase scan + interactive debate |
| `/bmad-break` | Project brief or prompt |
| `/clarify` | `problem.yaml` OR a project description as argument |
| `/ralph` | Backlog file, `backlog.yaml`, OR a text description |

Others require artifacts from earlier phases:

| Command | Requires |
|---------|----------|
| `/bmad-model` | `problem.yaml` |
| `/analyze` | `problem.yaml` + `architecture.yaml` + `backlog.yaml` |
| `/checklist` | `problem.yaml` + `architecture.yaml` + `backlog.yaml` |
| `/gsd-prep` | `backlog.yaml` + `architecture.yaml` |

Common standalone patterns:

```bash
# Just want Ralph to implement from a description
/ralph build a REST API for user management

# Define principles before anything else
/principles

# Clarify a problem description without running break first
/clarify I'm building a CLI tool for managing dotfiles...
```

---

## Agent Guardrails

The template ships with a set of behavioral constraints designed to prevent the most common failure modes observed in real multi-session usage.

### Approach Selection (`CLAUDE.md`)

Before implementing anything non-trivial, agents must:
1. **Scan first** — read how similar features are done in the codebase
2. **Propose before implementing** — present 2-3 options with tradeoffs and wait for a choice when multiple valid approaches exist
3. **Prefer existing over new** — reuse established patterns, modules, and dependencies
4. **Stay targeted** — answer what was asked, don't expand scope unilaterally

### IaC-Only Infrastructure (`rules/infrastructure.md`)

Cloud CLI commands (`gcloud`, `aws`, `az`, `kubectl`) are **read-only** tools:
- Allowed for investigation: `describe`, `get`, `list`
- **Banned for mutations**: creating, updating, or deleting cloud resources manually
- All infrastructure changes must go through Terraform, Helm, or the project's IaC tool
- Exceptions must be flagged as `TODO` comments in IaC files, never applied silently

### Pattern First + Placement Discovery (`rules/code-style.md`)

Before creating a new file or implementing a feature:
- **Pattern first**: find an existing feature of the same type and follow its exact conventions
- **Placement discovery**: locate where similar files live before creating new ones — never invent a directory that doesn't exist in the project

### Honest Test Pairing (`rules/testing.md` + `test-check` skill)

Every non-trivial function must have a corresponding test. When a function changes:
- Update the test if the **contract** changed intentionally
- **Never update a test just to make it pass** — a failing test means the function is broken
- Never weaken assertions (e.g. replacing `assertEqual(x, 42)` with `assertNotNil(x)`) to hide a failure

Use `/test-check` after modifying functions to run this automatically.

### Pre-Commit Visibility (`settings.json`)

A `PreToolUse` hook fires before every `git commit` or `git push`, printing the staged file list. This surfaces scope leaks (files from other stories or tasks) before they enter the commit.

---

## Inspiration

The BMAD workflow and its components draw from several methodologies:

- **[BMAD](https://github.com/bmadcode/BMAD-METHOD)** — Break, Model, Act, Deliver. The phased approach to taking a project from idea to implementation with structured gates.
- **[spec-kit](https://github.com/nicobailey-llc/spec-kit)** — Structured specification engineering. Inspired the `/principles` (project governance via structured debate), `/clarify` (ambiguity scanning), `/analyze` (cross-artifact consistency), and `/checklist` (pre-implementation quality gate) commands.
- **Ralph + ralph-loop** — Autonomous implementation lead pattern. Ralph parses a backlog into parallel rounds, spawns agent teammates with bounded context packs, and coordinates contract-first development. `/ralph-loop` enables session-resilient execution via stop hooks.

---

## Rules

Rules are modular project instructions loaded based on file patterns:

| Rule | Globs | What it enforces |
|------|-------|-----------------|
| `code-style.md` | `src/**`, `lib/**`, `app/**` | DRY, KISS, SOLID, pattern-first (scan before implementing), placement discovery (find where similar files live before creating new ones) |
| `testing.md` | `tests/**`, `**/*.test.*`, `**/*.spec.*` | Test-first, edge cases, independent tests, function-test pairing, no weakening assertions to hide failures |
| `security.md` | _(all files)_ | No secrets, input validation, least privilege |
| `api.md` | `src/routes/**`, `src/api/**`, `src/controllers/**` | REST conventions, pagination, error format |
| `frontend.md` | `src/components/**`, `**/*.tsx`, `**/*.jsx` | Small components, accessibility, state handling |
| `infrastructure.md` | `infra/**`, `*.tf`, `Dockerfile*`, `k8s/**` | IaC-only changes (no manual `gcloud`/`aws`/`kubectl` mutations), least-privilege IAM, non-root containers |
| `documentation.md` | `docs/**`, `**/*.md` | Close to code, examples, keep updated |
| `finops.md` | `infra/**`, `*.tf`, `k8s/**`, `helm/**` | Tagging, rightsizing, lifecycle, scheduling |

---

Upstream repository: [adrien-barret/claude-kit](https://github.com/adrien-barret/claude-kit)
