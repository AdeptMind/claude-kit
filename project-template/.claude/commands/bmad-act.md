---
name: bmad-act
description: BMAD Act phase – implement code from the backlog using Ralph agent teams
---

Act as the **implementation coordinator** for the BMAD Act phase.

This is Phase 3 of the BMAD workflow — all prior phases (Break, Clarify, Model, Analyze, Checklist, GSD Prep) should be complete before this runs.

## Prerequisites

1. Read `.claude/output/backlog.md` and `.claude/output/architecture.md`. If either does not exist, tell the user to run `/bmad-model` first and stop.
2. Read `.claude/output/principles.md` if it exists — pass project-specific standards to Ralph for teammate spawn prompts.
3. Read `.claude/output/checklist.md` if it exists — confirm there are no unresolved FAIL items. If there are, warn the user before proceeding.
4. Check if `.claude/output/gsd/prep-report.md` exists. If not, run `/gsd-prep` first to generate codebase mapping and context packs.

## Execution

Delegate to `/ralph` for the full implementation process:

- **Contract-first development**: shared interfaces committed before teammates start
- **Plan approval**: each teammate plans before coding, lead reviews
- **Parallel implementation**: teammates code in parallel within each round
- **Acceptance validation**: lead validates each story, feedback loop on failure
- **Quality checks**: code review, tests, security scan, dependency audit

Follow ALL of Ralph's steps as defined in `/ralph`. Do NOT skip any phase.

## BMAD Gate

This phase is complete when:
- All stories have `passes: true`
- The full test suite passes
- Quality checks report no critical issues
- `.claude/output/act-report.md` has been produced

If $ARGUMENTS is provided, use it as additional context or task filter: $ARGUMENTS

## Role-Aware Act Phase (po/all-roles only)

Check the `CK_USER_ROLE` environment variable. If it is `dev` or unset, skip this entire section — `/bmad-act` behaves exactly as above (delegate to Ralph, no extra roles).

If `CK_USER_ROLE` is `po` or `all-roles`, instruct Ralph to invoke the following role-specific reviews **after each implementation round**. Each role runs as a subagent using the corresponding agent definition.

### Per-Round Role Reviews

After Ralph completes a round's implementation and before moving to the next round, run these reviews in parallel:

1. **DevOps** (agent: `.claude/agents/devops.md`)
   - Review round changes for CI/CD configuration gaps
   - Check Dockerfile, deploy scripts, and infrastructure-as-code files
   - Flag missing or misconfigured pipeline steps, environment variables, or deployment targets

2. **Security** (agent: `.claude/agents/security.md`)
   - Run code security audit against OWASP top 10, injection vectors, secrets in code
   - Perform pentest checks on any new endpoints or exposed surfaces
   - Flag hardcoded credentials, missing input validation, insecure defaults

3. **FinOps** (agent: `.claude/agents/finops.md`)
   - Review resource choices: compute size, storage type, network configuration
   - Flag cost implications of new infrastructure or service selections
   - Suggest cheaper alternatives where applicable (spot instances, serverless, cold storage)

4. **SRE** (agent: `.claude/agents/sre.md`)
   - Validate observability: structured logging, metrics, distributed tracing
   - Review scaling strategy and auto-scaling configuration
   - Check failover, health checks, and circuit-breaker patterns

5. **QA**
   - Read `.claude/output/user-journey.md` if it exists
   - Prepare end-to-end test scenarios derived from user journeys
   - Write e2e test plan to `.claude/output/e2e-test-scenarios.md` for the Deliver phase

### Integrating Findings

Append each role's findings to the round review file (`round-N-review.md` in `.claude/output/gsd/`). Use the following format per role:

```markdown
### DevOps Review
- [findings or "No issues found"]

### Security Review
- [findings or "No issues found"]

### FinOps Review
- [findings or "No issues found"]

### SRE Review
- [findings or "No issues found"]

### QA Review
- [findings or "No issues found"]
```

If a role finds **critical** issues (security vulnerabilities, missing deployment config, broken observability), Ralph must address them before proceeding to the next round.
