# Claude Kit

**Turn Claude Code into a production team.**

```bash
brew tap adrien-barret/claude-kit && brew install claude-kit
cd my-project && ck init
```

That's it. Your project now has agents, workflows, guardrails, and a team lead that parallelizes work across specialists.

> *"As soon as the meeting finished I was already installing it in my repo. It really worked to the point of solving some issues I haven't even spotted in my code."*
> — Taina Martinez

---

## What it does

Claude Code is powerful but raw. Out of the box, every project starts from scratch: no roles, no workflow, no quality gates. You end up babysitting an agent that drifts off-scope.

Claude Kit fixes this with three things:

**1. A team of specialists** — 20+ agents (backend, devops, security, finops, architect...) that activate in the right context with the right model and skills.

**2. A structured workflow** — BMAD takes a project from idea to shipped feature through phases with quality gates. No more "just start coding and hope for the best."

**3. Ralph** — an autonomous team lead that reads your backlog, spawns parallel agent teams, coordinates contracts between them, and validates every story before moving on.

---

## See it in action

### From idea to implementation in one command

```
> /ralph build a REST API for user management with auth, roles, and rate limiting
```

Ralph will:
1. Break it into stories with acceptance criteria
2. Design the architecture and shared contracts
3. Spawn backend + security + devops agents in parallel
4. Each agent implements its stories following TDD and project conventions
5. Validate every story against acceptance criteria before marking it done
6. Run the full test suite and report results

You review the PR. That's your job now.

### Or go phase by phase

```
/bmad-break       # Define the problem → problem.yaml
/bmad-model       # Design architecture → architecture.yaml + backlog.yaml
/ralph            # Implement everything with agent teams
```

### Just need a quick fix?

```
/review           # Code review with severity levels
/test-gen         # Generate tests for your code
/security-check   # OWASP audit
/pentest          # Simulated penetration test
```

---

## Install

### Homebrew (recommended)

```bash
brew tap adrien-barret/claude-kit
brew install claude-kit
```

### From source

```bash
git clone https://github.com/adrien-barret/claude-kit
cd claude-kit && make install
```

### Setup your project

```bash
cd my-project
ck init              # Interactive picker — choose your agents
# or
ck init --plan       # Let Claude recommend components for your stack
```

### Add agents later

```bash
ck add backend devops security    # Add agents + their skills and rules
ck list                           # See what's available vs installed
ck sync                           # Update everything
```

---

## What's inside

### Agents
Backend, Frontend, Architect, Tech Lead, DevOps, SRE, Security, Pentester, FinOps, DBA, Product Owner, and more. Each has the right model, tools, and skills for its role.

### Workflow (BMAD)
Principles → Break → Clarify → Model → Analyze → Checklist → GSD Prep → Act → Deliver. Run the full pipeline with `/bmad-run` or pick individual phases.

### Guardrails
- **Pattern-first coding** — agents scan existing code before implementing anything
- **Honest test pairing** — tests are never weakened to hide failures
- **IaC-only infrastructure** — no manual `gcloud`/`aws` mutations
- **Pre-commit scope check** — surfaces file leaks before they enter a commit

### Skills
Code review, test generation, security audit, pentest simulation, threat modeling, performance audit, accessibility audit, API documentation, FinOps cost optimization, and 20+ more.

---

## Shortcuts

| You type | What happens |
|----------|-------------|
| `/ralph <description>` | Full implementation with agent teams |
| `/bmad-run` | Complete workflow from idea to delivery |
| `/review` | Code review on current changes |
| `/test-gen` | Generate tests |
| `/security-check` | Security audit |
| `/pentest` | Penetration test simulation |
| `/quick-spec` | Lightweight tech spec |
| `/quick-dev` | Quick implementation from spec or description |

---

## Inspiration

- [BMAD Method](https://github.com/bmadcode/BMAD-METHOD) — the phased workflow approach
- [spec-kit](https://github.com/nicobailey-llc/spec-kit) — structured specification engineering
- Community patterns — TDD enforcement, subagent-driven review, and granular planning from the Claude Code ecosystem

---

<details>
<summary><strong>Full CLI Reference</strong></summary>

### Core commands

| Command | Description |
|---------|-------------|
| `ck init` | Interactive setup — pick components from a categorized list |
| `ck init --plan` | AI-guided setup via Claude session |
| `ck init --global` | Install to `~/.claude` |
| `ck add [names...]` | Add agents by name (auto-installs skills + rules) |
| `ck remove [names...]` | Remove components |
| `ck list` | Available vs installed components |
| `ck sync` | Update installed components + refresh docs-index |
| `ck docs` | Generate stack-aware docs-index.md |
| `ck dep install` | Install recommended dependencies |
| `ck profile list\|use\|add\|remove` | Manage Claude account profiles |
| `ck version` | Print version |

### Eval & Quality

| Command | Description |
|---------|-------------|
| `ck skill eval <skill-dir>` | Test trigger accuracy against eval queries |
| `ck skill optimize <skill-dir>` | Eval-improve loop to optimize skill descriptions |
| `ck skill grade <skill-dir>` | Grade output against assertions |
| `ck skill benchmark <results-dir>` | Aggregate and compare grading stats |
| `ck skill validate <skill-dir>` | Validate skill structure |
| `ck skill report <results.json>` | Generate interactive HTML report |

### Packaging & Distribution

| Command | Description |
|---------|-------------|
| `ck skill package <skill-dir>` | Package as .skill archive |
| `ck agent validate <agent.md>` | Validate agent frontmatter and skills |
| `ck agent package <agent.md>` | Package as .agent archive |
| `ck agents registry` | Generate agent-registry.yaml |
| `ck package <template-dir>` | Bundle full template as .claude-kit archive |
| `ck install <archive>` | Install .skill, .agent, or .claude-kit archive |

### BMAD Eval

| Command | Description |
|---------|-------------|
| `ck bmad eval [output-dir]` | Evaluate BMAD artifacts against phase assertions |
| `ck bmad benchmark <run1> <run2>` | Compare two BMAD eval runs |

</details>

<details>
<summary><strong>All slash commands</strong></summary>

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

**Quick track:**
`/quick-spec` (`/qs`), `/quick-dev` (`/qd`)

**Ideation:**
`/brainstorm`, `/party`

**Utilities:**
`/ck-sync`, `/shard`, `/create-component`, `/bmad-help` (`/h`)

**Roles:**
`/role-backend`, `/role-frontend`, `/role-architect`, `/role-tech-lead`, `/role-devops`, `/role-security`, `/role-pentester`, `/role-product-owner`, `/role-ux-designer`, `/role-ui-designer`

</details>

<details>
<summary><strong>Skills reference</strong></summary>

### Dev Skills

| Skill | What it does |
|-------|-------------|
| `code-reviewer` | Code review with severity levels and auto-fix suggestions |
| `test-generator` | Test generation with framework detection |
| `test-check` | Per-function test coverage — never weakens assertions |
| `api-documenter` | OpenAPI/Swagger generation |
| `git-commit-helper` | Conventional commit messages |
| `dependency-auditor` | Vulnerability scanning and license audit |
| `performance-audit` | N+1 queries, bundle size, caching |
| `accessibility-audit` | WCAG 2.1 AA compliance |
| `database-review` | Schema, indexing, query optimization |

### Security Skills

| Skill | What it does |
|-------|-------------|
| `code-security-audit` | OWASP Top 10, injection, XSS, secrets |
| `infra-security-audit` | Cloud config, permissions, encryption |
| `auth-review` | OAuth/JWT, RBAC, token policies |
| `secret-rotation` | Secret storage and rotation |
| `pentest-web` | Auth bypass, IDOR, SSRF, JWT attacks |
| `threat-model` | STRIDE threat modeling |

### FinOps Skills

| Skill | What it does |
|-------|-------------|
| `cost-optimization` | Rightsizing, auto-scaling, reserved instances |
| `tagging-audit` | Cost allocation tag compliance |
| `waste-detection` | Idle resources, orphaned volumes |
| `budget-forecast` | Cost estimation from IaC |

</details>

<details>
<summary><strong>Guardrails detail</strong></summary>

### Pattern-first coding (`rules/code-style.md`)
Before implementing, agents scan the codebase for similar features and follow existing patterns. Before creating a file, they find where similar files live — never inventing directories.

### Honest test pairing (`rules/testing.md`)
Every function has a test. When a function changes, the test updates if the contract changed — but never weakens assertions to hide failures. `/test-check` enforces this automatically.

### IaC-only infrastructure (`rules/infrastructure.md`)
Cloud CLI is read-only (`describe`, `get`, `list`). All mutations go through Terraform/Helm. No exceptions.

### Pre-commit visibility (`settings.json`)
A hook fires before `git commit`, printing staged files to catch scope leaks.

### Approach selection (`CLAUDE.md`)
Agents scan first, propose 2-3 options when ambiguous, prefer existing patterns, and stay on scope.

</details>

<details>
<summary><strong>Build from source</strong></summary>

### Prerequisites
- Go 1.21+
- Make

### Commands

```bash
make build              # Compile binary
make install            # Build + install to /usr/local/bin + templates
make clean              # Remove build artifacts
make uninstall          # Remove everything
```

### Template resolution order
1. `$BMAD_TEMPLATE_DIR` environment variable
2. `~/.bmad/templates/`
3. Adjacent `project-template/.claude/` (development)

### Dependencies
[cobra](https://github.com/spf13/cobra), [bubbletea](https://github.com/charmbracelet/bubbletea), [huh](https://github.com/charmbracelet/huh), [lipgloss](https://github.com/charmbracelet/lipgloss)

</details>

<details>
<summary><strong>Supported stacks</strong></summary>

`ck docs` auto-detects your stack and generates framework-specific notes.

**Languages:** JavaScript, TypeScript, Python, Go, Ruby, Rust, Java, Kotlin, PHP

**Frameworks:** Next.js, React, Vue, Nuxt, Svelte, Angular, Express, Fastify, NestJS, Hono, Django, Flask, FastAPI, Rails, Sinatra, Laravel, Symfony

**Tools:** Docker, Terraform, Kubernetes, Helm, GitHub Actions, Prisma, Drizzle, Tailwind

</details>

---

[GitHub](https://github.com/adrien-barret/claude-kit)
