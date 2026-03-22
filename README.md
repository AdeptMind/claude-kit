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

> **Legend**
>
> **Agents** = Specialized AI personas (backend, security, PO...) with the right model, tools, and skills for their role.
> **Skills** = Reusable capabilities (code review, TDD, pentest...) that agents invoke.
> **Rules** = Guardrails enforced automatically (pattern-first coding, honest tests, IaC-only...).
> **BMAD** = Break → Model → Act → Deliver — the phased workflow from idea to production.
> **Ralph** = The autonomous team lead that orchestrates parallel agent teams.
> **Roles** = Working methods (dev, po, qa...) that adapt how the workflow behaves.
> **Cowork** = Claude Desktop's agentic mode — Claude Kit configures it, Cowork executes.

---

## Why Claude Kit?

Claude Code is powerful but raw. Out of the box: no roles, no workflow, no quality gates. You babysit an agent that drifts off-scope.

Claude Kit adds **three layers**:

**A team of 40+ specialists** — backend, devops, security, finops, architect, PO, business analyst, project manager, QA, HR... Each activates in the right context with the right model and skills.

**A structured workflow with quality gates** — BMAD takes a project from idea to shipped feature through phases. TDD is enforced. Every story gets a readiness check before implementation and a done check after. No more "just start coding and hope for the best."

**Role-based orchestration** — Switch between working methods with `ck user`. A PO gets validation gates, product-language reports, and client challenge loops. A dev gets zero friction. A QA gets user journey verification. In all-roles mode, each BMAD phase activates the right perspective automatically.

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
4. Each agent implements following TDD — code without a failing test gets deleted
5. Double review: spec conformity + code quality before acceptance
6. Validate every story against acceptance criteria
7. Run the full test suite and report results

You review the PR. That's your job now.

### Phase by phase

```
/bmad-break       # Define the problem → problem.yaml
/bmad-model       # Design architecture → architecture.yaml + backlog.yaml
/ralph            # Implement everything with agent teams
```

### Quick fixes

```
/review           # Code review with severity levels
/test-gen         # Generate tests for your code
/security-check   # OWASP audit
/pentest          # Simulated penetration test
```

### Switch perspectives

```bash
ck user po        # Product owner — validation gates, product-language reports
ck user qa        # QA — user journey verification, visual proof
ck user dev       # Developer — fast cycles, zero friction (default)
ck user all       # Full orchestration — right role at each BMAD phase
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
ck init              # Interactive picker — choose your agents + security profile
# or
ck init --plan       # Let Claude recommend components for your stack
```

### Add agents later

```bash
ck add backend devops security    # Add agents + their skills and rules
ck add business-analyst           # Management agents too
ck list                           # See what's available vs installed
ck sync                           # Update everything
```

---

## What's inside

### 40+ Agents

**Dev**: backend, frontend, devops, SRE, security, pentester, DBA, data scientist, ML engineer, Go, Python, TypeScript, Rust, mobile (iOS/Android/RN/Flutter)...

**Management**: product owner, business analyst, project manager, scrum master, architect, tech lead, UX/UI designer, FinOps, change manager, HR/people ops, data governance, SOC2 compliance, access review...

Each has the right model, tools, and skills for its role.

### Workflow (BMAD)

Principles → Break → Clarify → UX Spec → Model → Analyze → Checklist → GSD Prep → Act → Deliver. Run the full pipeline with `/bmad-run` or pick individual phases.

### Role-Based Working Methods

| Role | What changes |
|------|-------------|
| `dev` | Fast cycles, no gates — current behavior (default) |
| `po` | Validation gates, product-language reports, client challenge loops, so_that proof |
| `qa` | User journey verification, visual evidence (screenshots/output), e2e readiness |
| `all` | Full orchestration: PO for Break, Architect+TL+SRE for Model, Dev+DevOps+Security for Act, QA+PO for Deliver |

### Quality Enforcement

- **TDD enforced** — code without a failing test gets deleted and restarted
- **Double review** — spec conformity + code quality review before acceptance
- **Readiness check (DoR)** — stories validated before implementation begins
- **Done check (DoD)** — business value proof required, not just passing tests
- **Traceability** — every task traces to a user story to a pain point to a persona

### Security & Policy

```bash
ck policy init       # Pick a security profile (permissive / moderate / strict)
ck policy apply      # Merge policy into settings.json with diff preview
ck audit view        # See what agents did — structured JSONL log
ck audit export      # Export to CSV/JSON for SIEM
ck sandbox create    # Run Claude Code in an isolated container
```

Profiles are policy-as-code (`.claude/policy.yaml`) — committed, shareable, inheritable.

### Desktop App (macOS)

A Wails-based `.app` for non-technical users:
- Visual project dashboard with BMAD phase progress
- File manager with preview (yaml, markdown, excel, pdf)
- Cowork profile editor — form-based, no markdown editing
- Management roles prominent, dev roles collapsible
- One-click "Open in Cowork" for workflow execution

### Guardrails

- **Pattern-first coding** — agents scan existing code before implementing
- **Honest test pairing** — tests never weakened to hide failures
- **IaC-only infrastructure** — no manual cloud CLI mutations
- **Pre-commit scope check** — surfaces file leaks before commit

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
| `/quick-dev` | Quick implementation from spec |
| `/brainstorm` | Creative ideation |
| `/party` | Multi-agent debate |

---

## Inspiration

- [BMAD Method](https://github.com/bmadcode/BMAD-METHOD) — the phased workflow approach
- [Superpowers](https://github.com/obra/superpowers) — TDD enforcement and structured development
- [spec-kit](https://github.com/nicobailey-llc/spec-kit) — structured specification engineering

---

<details>
<summary><strong>Full CLI Reference</strong></summary>

### Core commands

| Command | Description |
|---------|-------------|
| `ck init` | Interactive setup — agents + security profile |
| `ck init --plan` | AI-guided setup via Claude session |
| `ck add [names...]` | Add agents by name (auto-installs deps) |
| `ck remove [names...]` | Remove components |
| `ck list` | Available vs installed |
| `ck sync` | Update + refresh docs-index |
| `ck docs` | Generate stack-aware docs-index.md |
| `ck user [role]` | View/set working method (dev, po, qa, all...) |
| `ck user <role> --global` | Set role globally |
| `ck policy init` | Pick a security profile |
| `ck policy apply` | Merge policy into settings.json |
| `ck policy show` | Show resolved policy |
| `ck audit view` | View agent activity log |
| `ck audit export` | Export log to CSV/JSON |
| `ck sandbox create` | Generate container config from policy |
| `ck sandbox run` | Launch sandboxed Claude Code |
| `ck sandbox stop` | Stop container |
| `ck dep install` | Install recommended dependencies |
| `ck profile list\|use\|add\|remove` | Manage Claude profiles |
| `ck version` | Print version |

### Eval & Quality

| Command | Description |
|---------|-------------|
| `ck skill eval <skill-dir>` | Test trigger accuracy |
| `ck skill optimize <skill-dir>` | Eval-improve loop |
| `ck skill grade <skill-dir>` | Grade output against assertions |
| `ck skill benchmark <results-dir>` | Compare grading stats |
| `ck skill validate <skill-dir>` | Validate structure |
| `ck skill report <results.json>` | HTML report |

### Packaging

| Command | Description |
|---------|-------------|
| `ck skill package <skill-dir>` | Package as .skill |
| `ck agent validate <agent.md>` | Validate agent |
| `ck agent package <agent.md>` | Package as .agent |
| `ck agents registry` | Generate registry |
| `ck package <template-dir>` | Bundle as .claude-kit |
| `ck install <archive>` | Install archive |

</details>

<details>
<summary><strong>All slash commands</strong></summary>

**BMAD Workflow:**
`/bmad-run`, `/bmad-break`, `/bmad-model`, `/bmad-act`, `/bmad-deliver`

**Spec & Quality Gates:**
`/principles`, `/clarify`, `/analyze`, `/checklist`, `/ux-spec`

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

</details>

<details>
<summary><strong>Skills reference</strong></summary>

### Dev Skills

| Skill | What it does |
|-------|-------------|
| `code-reviewer` | Code review with severity levels |
| `test-generator` | Test generation with framework detection |
| `test-check` | Per-function coverage — never weakens assertions |
| `tdd-enforced` | RED → GREEN → REFACTOR cycle enforcement |
| `api-documenter` | OpenAPI/Swagger generation |
| `git-commit-helper` | Conventional commit messages |
| `dependency-auditor` | Vulnerability scanning |
| `performance-audit` | N+1 queries, bundle size, caching |
| `accessibility-audit` | WCAG 2.1 AA compliance |
| `database-review` | Schema, indexing, query optimization |

### Quality Gates

| Skill | What it does |
|-------|-------------|
| `readiness-check` | Definition of Ready — validates before implementation |
| `done-check` | Definition of Done — requires business value proof |
| `traceability-check` | Task → story → pain point matrix |
| `spec-reviewer` | Spec conformity review (subagent) |
| `plan-writer` | Granular implementation plans for complex stories |

### Security

| Skill | What it does |
|-------|-------------|
| `code-security-audit` | OWASP Top 10, injection, XSS, secrets |
| `infra-security-audit` | Cloud config, permissions, encryption |
| `auth-review` | OAuth/JWT, RBAC, token policies |
| `secret-rotation` | Secret storage and rotation |
| `pentest-web` | Auth bypass, IDOR, SSRF, JWT attacks |
| `threat-model` | STRIDE threat modeling |

### FinOps

| Skill | What it does |
|-------|-------------|
| `cost-optimization` | Rightsizing, auto-scaling, reserved instances |
| `tagging-audit` | Cost allocation tag compliance |
| `waste-detection` | Idle resources, orphaned volumes |
| `budget-forecast` | Cost estimation from IaC |

### Management

| Skill | What it does |
|-------|-------------|
| `requirements-elicitation` | Structured requirements discovery |
| `process-mapping` | As-is/to-be business process flows |
| `gap-analysis` | Current vs desired state comparison |
| `risk-assessment` | Risk identification and scoring |
| `milestone-tracking` | Timeline tracking and forecasting |
| `sprint-planning` | Sprint capacity and story selection |
| `retrospective` | Structured retro facilitation |
| `policy-drafting` | Organizational policy documents |
| `impact-assessment` | Change impact across people/process/tech |
| `communication-planning` | Stakeholder communication plans |

</details>

<details>
<summary><strong>Guardrails detail</strong></summary>

### TDD Enforced (`rules/testing.md` + `skills/tdd-enforced`)
RED → GREEN → REFACTOR. Code written before a failing test exists gets deleted. No exceptions for code stories. Anti-rationalization table handles 12 common excuses.

### Double Review (Ralph Phase D)
After a teammate reports done: spec-reviewer (conformity) + code-quality-reviewer (quality) run in parallel as fresh subagents. Both must PASS before acceptance-validator runs. Skipped in dev mode.

### Pattern-first coding (`rules/code-style.md`)
Before implementing, agents scan the codebase for similar features and follow existing patterns. Before creating a file, they find where similar files live — never inventing directories.

### Honest test pairing (`rules/testing.md`)
Every function has a test. When a function changes, the test updates if the contract changed — but never weakens assertions to hide failures.

### IaC-only infrastructure (`rules/infrastructure.md`)
Cloud CLI is read-only (`describe`, `get`, `list`). All mutations go through Terraform/Helm.

### Pre-commit visibility (`settings.json`)
A hook fires before `git commit`, printing staged files to catch scope leaks.

</details>

<details>
<summary><strong>Build from source</strong></summary>

### Prerequisites
- Go 1.21+
- Make
- Wails v2 (for desktop app only)

### Commands

```bash
make build              # Compile CLI binary
make install            # Build + install to /usr/local/bin + templates
make clean              # Remove build artifacts
make uninstall          # Remove everything
```

### Desktop app

```bash
cd cmd/claude-kit-ui
wails build             # Produces .app bundle
```

### Template resolution order
1. `$BMAD_TEMPLATE_DIR` environment variable
2. `~/.bmad/templates/`
3. Adjacent `project-template/.claude/` (development)

</details>

<details>
<summary><strong>Supported stacks</strong></summary>

`ck docs` auto-detects your stack and generates framework-specific notes.

**Languages:** JavaScript, TypeScript, Python, Go, Ruby, Rust, Java, Kotlin, PHP

**Frameworks:** Next.js, React, Vue, Nuxt, Svelte, Angular, Express, Fastify, NestJS, Hono, Django, Flask, FastAPI, Rails, Sinatra, Laravel, Symfony

**Tools:** Docker, Terraform, Kubernetes, Helm, GitHub Actions, Prisma, Drizzle, Tailwind

</details>

<details>
<summary><strong>Claude Kit vs Superpowers</strong></summary>

Both enhance Claude Code with structured workflows. Here's how they differ:

| | **Claude Kit** | **Superpowers** |
|---|---|---|
| **Scope** | Full project lifecycle — idea to delivery | Code-focused — spec to implementation |
| **Workflow** | 10-phase BMAD (Break → Model → Analyze → Act → Deliver) with quality gates between each | 7-phase (Brainstorm → Spec → Plan → TDD → Subagent → Review → Finalize) |
| **Agents** | 40+ specialized agents (backend, security, finops, PO, BA, PM, HR, QA...) | Fresh subagents per task (no persistent roles) |
| **Roles** | Role-based working methods — `ck user po/qa/dev/all` adapts the entire workflow | Single workflow, same for everyone |
| **Non-dev users** | PO, PM, HR, QA get their own agents, profiles, and desktop app | Developer-only |
| **Desktop app** | macOS .app with dashboard, file manager, profile editor | No GUI |
| **TDD** | Enforced via skill + rule + Ralph integration | Enforced — deletes code written before tests |
| **Review** | Double review (spec + quality) as parallel subagents | Single review phase |
| **Architecture** | Full architecture design phase (ADRs, data model, API surface) | No architecture phase |
| **Security** | Dedicated agents + skills (pentest, threat model, secret rotation) + policy profiles + container sandbox | Basic security in review |
| **FinOps** | Dedicated agent + 4 skills (cost optimization, tagging, waste, budget) | None |
| **Quality gates** | DoR (readiness) + DoD (business value proof) + traceability matrix | Spec validation |
| **Client engagement** | Challenge loops at every milestone, visual round reviews | None |
| **Cowork integration** | Profile editor + push to Cowork + launch from app | None |
| **Distribution** | CLI (`ck`) + desktop app + Homebrew | CLAUDE.md configuration |
| **Packaging** | `.skill`, `.agent`, `.claude-kit` archives | None |
| **Eval framework** | `ck skill eval/grade/benchmark/optimize` | None |

**When to use which:**
- **Superpowers** if you want TDD-first coding with minimal setup and you're a developer working solo
- **Claude Kit** if you want a full project lifecycle tool that works for developers AND non-technical stakeholders, with role-based workflows, security policies, and team orchestration

They're not mutually exclusive — Claude Kit's TDD enforcement was inspired by Superpowers, and both can coexist in the same project.

</details>

---

[GitHub](https://github.com/adrien-barret/claude-kit)
