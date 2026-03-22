---
name: analyze
description: Cross-artifact consistency analysis — read-only check for gaps, conflicts, and drift across all BMAD outputs
---

Act as a **QA Analyst** performing a cross-artifact consistency check. This command is **read-only** — it modifies nothing, only reports findings.

## Prerequisites

All three are required:
- `.claude/output/problem.md` — if missing, tell the user to run `/bmad-break` first and stop
- `.claude/output/architecture.md` — if missing, tell the user to run `/bmad-model` first and stop
- `.claude/output/backlog.md` — if missing, tell the user to run `/bmad-model` first and stop

Optional (used if present):
- `.claude/output/principles.md` — project principles for violation checks

Read all artifacts before starting the analysis.

## Analysis Categories

### 1. Duplications
Detect requirements, features, or tasks that appear in multiple places with different wording but the same intent. Flag cases where:
- Two backlog tasks cover the same functionality
- A feature in `problem.md` maps to multiple overlapping tasks
- Architecture components duplicate responsibilities

### 2. Ambiguities
Find requirements or tasks that are vague enough to produce different implementations depending on interpretation:
- User stories missing acceptance scenarios
- Tasks without clear acceptance criteria
- Architecture decisions that reference undefined components
- "TBD", "TODO", or placeholder values

### 3. Underspecification
Identify areas where critical detail is missing:
- Features without user stories
- User stories without `so_that` (missing business value / WHY)
- Tasks without `depends_on` that clearly need prior work
- Architecture components without defined interfaces
- Integrations without error handling strategy

### 4. Principles Violations (if `principles.md` exists)
Cross-reference artifacts against the project principles:
- Testing standards not reflected in acceptance criteria
- Security principles not addressed in relevant features
- Architecture principles contradicted by design decisions
- Performance targets without corresponding non-functional requirements

### 5. Coverage Gaps
Check that the full chain is complete: every requirement maps to architecture, every architecture component maps to backlog tasks:
- Requirements in `problem.md` not covered by any backlog task
- Architecture components not exercised by any task
- Backlog tasks that don't trace back to any requirement (orphan tasks)

### 6. Inconsistencies
Detect contradictions between artifacts:
- `problem.md` says REST but `architecture.md` defines GraphQL
- Priority P1 in problem but task is marked low priority in backlog
- Tech stack mismatch between problem definition and architecture
- Conflicting non-functional requirements (e.g., "real-time" + "batch processing" for same data)

### 7. Terminology Drift
Find cases where the same concept uses different names across artifacts:
- "user" vs "customer" vs "account" for the same entity
- "order" vs "purchase" vs "transaction" for the same flow
- Component names that don't match between architecture and backlog

## Role-Aware Analysis (po/all-roles only)

> **Gate**: check `CK_USER_ROLE`. If the value is `dev` or unset, skip this entire section — the analysis stops at category 7 above and proceeds directly to Severity Levels.

When `CK_USER_ROLE` is `po` or `all`, run the four additional analysis categories below. Findings use the same severity levels and output format as the core categories.

### 8. Traceability
Invoke the `traceability-check` skill and include its matrix output in the report. The matrix maps every requirement in `problem.md` to its architecture component(s) and backlog task(s). Flag:
- Requirements with no architecture mapping (orphan requirements)
- Requirements with no backlog task (unplanned work)
- Backlog tasks with no requirement origin (orphan tasks)
- Architecture components not referenced by any requirement (dead components)

### 9. Business Value Quality
Review each user story's `so_that` clause in `problem.md`. Flag:
- Missing `so_that` — no business justification provided
- Weak justifications that restate the action instead of the value (e.g., "so that I can click the button" instead of "so that I can track my spending")
- Generic justifications that could apply to any feature (e.g., "so that the system works better")
- Duplicate business value across unrelated stories (copy-paste smell)

### 10. SRE Operability
Review `architecture.md` for production-readiness gaps. Flag:
- **Observability**: missing or incomplete logging, monitoring, or tracing strategy
- **Scaling**: no scaling strategy defined, or scaling strategy that doesn't match expected load from `problem.md`
- **Failover**: no failover or disaster recovery plan for stateful components
- **Circuit breakers**: external integrations without circuit breaker or retry/backoff strategy
- **Health checks**: services without defined health check endpoints

Critical SRE findings (e.g., no observability strategy, no failover for stateful data) use CRITICAL severity and block progression — same gate behavior as existing critical findings.

### 11. Security Threat Surface
Review `architecture.md` for security gaps. Flag:
- **Authentication**: endpoints or services without auth coverage; missing auth strategy
- **Input validation**: user-facing interfaces without input validation strategy
- **OWASP concerns**: architecture patterns susceptible to OWASP Top 10 (injection, broken access control, security misconfiguration, etc.)
- **Secrets exposure**: hardcoded credentials, missing secrets management strategy, or secrets passed in environment without encryption
- **Attack vectors**: publicly exposed services without rate limiting, WAF, or DDoS protection

Critical Security findings (e.g., no auth on public endpoints, no secrets management) use CRITICAL severity and block progression — same gate behavior as existing critical findings.

## Severity Levels

- **CRITICAL** — Will cause implementation failure or major rework. Blocks `/bmad-run` progression.
- **HIGH** — Likely to cause bugs or incorrect implementation. Should be fixed before coding.
- **MEDIUM** — May cause confusion or suboptimal implementation. Fix recommended.
- **LOW** — Minor inconsistency or style issue. Fix at convenience.

## Output Format

Report to stdout (do NOT write files):

```
Cross-Artifact Analysis Report
═══════════════════════════════

Artifacts analyzed:
  - problem.md (version X, N features, N user stories)
  - architecture.md (N components, N decisions)
  - backlog.md (N tasks across N rounds)
  - principles.md (present/absent)

Findings: N total (N critical, N high, N medium, N low)

─── CRITICAL ───────────────────────────────────────

[C-001] Category: Coverage Gap
  Requirement "payment processing" (problem.md, feature #3) has no
  corresponding backlog task. This feature will not be implemented.
  → Fix: Add tasks for payment processing to the backlog.

─── HIGH ───────────────────────────────────────────

[H-001] Category: Inconsistency
  problem.md specifies PostgreSQL but architecture.md references
  MongoDB in the data layer.
  → Fix: Align database choice across artifacts.

─── MEDIUM ─────────────────────────────────────────

[M-001] Category: Underspecification
  User story US-003 has no acceptance scenarios (Given/When/Then).
  → Fix: Add acceptance scenarios to problem.md.

─── LOW ────────────────────────────────────────────

[L-001] Category: Terminology Drift
  "user" in problem.md, "account" in architecture.md,
  "customer" in backlog.md — all refer to the same entity.
  → Fix: Standardize on one term.

─── Coverage Summary ───────────────────────────────

| Requirement          | Architecture Component | Backlog Tasks | Status   |
|----------------------|------------------------|---------------|----------|
| User auth            | auth-service           | T-001, T-002  | Covered  |
| Payment processing   | —                      | —             | MISSING  |
| Notifications        | notification-service   | T-005         | Covered  |

─── Verdict ────────────────────────────────────────

{PASS — no critical issues, safe to proceed}
{BLOCK — N critical issues must be resolved before implementation}
```

## Gate Behavior

When run as part of `/bmad-run`:
- **CRITICAL issues block progression** — the workflow stops until they are resolved
- HIGH/MEDIUM/LOW issues are reported but don't block

When run standalone (`/analyze`):
- All findings are reported; the user decides what to act on

If $ARGUMENTS is provided, use it as focus area or additional context: $ARGUMENTS
