---
name: bmad-deliver
description: BMAD Deliver phase – prepare release with deployment scripts, docs, and release notes
---

Act as a **DevOps Engineer** and **Tech Lead** working together, using the agents defined in `.claude/agents/devops.md` and `.claude/agents/tech-lead.md`.

Your goal is to prepare the project for release: deployment configuration, documentation, and release notes.

## Prerequisites

1. Read `.claude/output/architecture.md` for infrastructure and deployment design.
2. Read `.claude/output/act-report.md` for what was implemented.
3. Read `.claude/output/principles.md` if it exists — use security and deployment principles to guide the release process.
4. Scan the project source tree to understand the current state of the code.

If no code exists in the project, tell the user to run `/bmad-act` first and stop.

## Stage 1: Deployment Configuration

Based on the architecture, create or verify:

1. **Containerization** (if applicable):
   - Dockerfile with multi-stage build, non-root user, pinned base image
   - .dockerignore excluding unnecessary files

2. **Infrastructure as Code** (if applicable):
   - Terraform, CloudFormation, or Pulumi configs for cloud resources
   - Proper variable extraction (no hardcoded values)
   - Least-privilege IAM roles and policies
   - Resource tagging (environment, team, service, cost-center)

3. **CI/CD Pipeline**:
   - Build, test, lint, security-scan stages
   - Environment-specific deployment (dev, staging, production)
   - Secrets injected from environment/secret manager (never hardcoded)

4. **Configuration Management**:
   - Environment-based configuration (env vars or config files)
   - Separation between secrets and non-sensitive config
   - Health check endpoints

## Stage 2: Documentation

Create or update:

1. **README.md**: Project overview, setup instructions, development guide, deployment guide.
2. **API documentation**: If an API exists, generate or update OpenAPI/Swagger docs.
3. **Architecture diagram description**: Text-based description of the system architecture.
4. **Runbook**: Basic operational procedures (deploy, rollback, monitor, troubleshoot).

## Stage 3: Security Review

Run a final security check:

1. No hardcoded secrets in source code or config files.
2. Dependencies are pinned and free of known critical vulnerabilities.
3. Authentication and authorization are properly configured.
4. HTTPS/TLS is enforced for external communication.
5. Logging does not expose sensitive data.

## Stage 4: Release Notes

Create `.claude/output/release-notes.md`:

```markdown
# Release Notes - <project name> v<version>

## Summary
<one-paragraph description of what this release includes>

## Features
- <feature 1>: <description>
- <feature 2>: <description>

## Architecture
- <key architectural decisions>

## Infrastructure
- **Compute**: <what's used>
- **Database**: <what's used>
- **CI/CD**: <pipeline summary>

## Security
- <authentication method>
- <key security measures>

## Known Limitations
- <limitation 1>

## Deployment
1. <step 1>
2. <step 2>
3. <step 3>

## Configuration
| Variable | Description | Required |
|----------|-------------|----------|
| <VAR> | <description> | yes/no |
```

## Stage 5: Validate

Present the release checklist to the user:

- [ ] Deployment config is complete and tested
- [ ] Documentation is up to date
- [ ] Security review passed
- [ ] Release notes are accurate
- [ ] All acceptance criteria from the backlog are met

Ask the user to confirm the release is ready.

If $ARGUMENTS is provided, use it as additional context: $ARGUMENTS

## Role-Aware Delivery Validation (po/all-roles only)

Check `CK_USER_ROLE`. If it is `dev` or unset, **skip this entire section** — the command behaves exactly as above.

If `CK_USER_ROLE` is `po` or `all-roles`, execute the following multi-role validation before presenting the final checklist. Each role produces a **sign-off block** that is appended to `.claude/output/act-report.md` under a `## Delivery Sign-offs` section.

### QA — User Journey Verification

1. Read `.claude/output/user-journey.md` (if it exists).
2. For each journey step, execute a manual walkthrough or automated check against the running application.
3. Capture evidence (command output, screenshots, or log excerpts) for each step.
4. Report **pass/fail per step** with evidence summary.

Sign-off format:

```
### QA Sign-off
- **Status**: PASS | FAIL
- **Journey steps**: X/Y passed
- **Evidence**: <summary of captured evidence per step>
- **Blocking issues**: <list or "None">
```

### PO — Business Value Confirmation

1. Read `.claude/output/problem.md` — extract each feature's `so_that` clause.
2. For each feature, confirm the business value is delivered with concrete evidence (demo output, test results, or user-facing behavior).
3. Flag any feature whose `so_that` is not demonstrably met.

Sign-off format:

```
### PO Sign-off
- **Status**: APPROVED | REJECTED
- **Features validated**: X/Y
- **Evidence per feature**:
  - <feature>: <evidence of business value delivered>
- **Gaps**: <list or "None">
```

### DevOps — Deploy Readiness

1. Verify CI/CD pipeline is green and produces deployable artifacts.
2. Confirm staging environment has been tested (or can be tested).
3. Verify a rollback plan is documented (in runbook or release notes).
4. Check that environment-specific configs and secrets are externalized.

Sign-off format:

```
### DevOps Sign-off
- **Status**: READY | NOT READY
- **CI/CD pipeline**: GREEN | RED — <link or summary>
- **Staging tested**: YES | NO
- **Rollback plan**: DOCUMENTED | MISSING
- **Config externalized**: YES | NO — <details>
```

### FinOps — Cost Report

1. Read `.claude/output/architecture.md` for estimated resource costs (if present).
2. Compare estimated costs against actual or projected resource usage.
3. Report budget compliance status.

Sign-off format:

```
### FinOps Sign-off
- **Status**: COMPLIANT | OVER BUDGET | NO ESTIMATE
- **Estimated cost**: <from architecture or "not specified">
- **Actual/projected cost**: <computed or "not measurable yet">
- **Variance**: <percentage or "N/A">
- **Recommendations**: <list or "None">
```

### Security — Final Vulnerability Report

1. Scan dependencies for known vulnerabilities (`npm audit`, `pip-audit`, or equivalent).
2. Review code for hardcoded secrets, SQL injection, and XSS vectors.
3. Check OWASP Top 10 compliance for web-facing components.
4. Verify TLS, security headers, and least-privilege IAM (if applicable).

Sign-off format:

```
### Security Sign-off
- **Status**: PASS | FAIL
- **Dependency vulnerabilities**: <count critical/high/medium/low>
- **Code audit findings**: <list or "None">
- **OWASP compliance**: <summary>
- **Blocking issues**: <list or "None">
```

### Collecting Sign-offs

After all five roles have produced their sign-off blocks:

1. Append a `## Delivery Sign-offs` section to `.claude/output/act-report.md` containing all five blocks.
2. Add a summary verdict at the top of the section:
   - **RELEASE APPROVED** — all roles passed/approved/ready/compliant
   - **RELEASE BLOCKED** — one or more roles reported blocking issues (list them)
3. Present the summary to the user and ask for final confirmation before proceeding.
