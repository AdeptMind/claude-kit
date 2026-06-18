---
description: Engineering-guide-level documentation rules — README, architecture, ADRs, runbooks, onboarding
globs: ["docs/**", "**/*.md", "README*", "CHANGELOG*"]
---

> **Why this matters**: documentation lives at the system level, not the line level. The goal is that a new dev can set up + understand the codebase in under one hour from the docs, and a 3 AM oncall can find the answer in five minutes. Inline comments are NOT documentation — see @.claude/rules/code-comments.md for that.

## What "documented" means here

A project is documented when these artifacts exist, are current, and are scannable:

1. **README.md** at repo root
2. **Architecture overview** (in README or `docs/architecture.md`)
3. **ADRs** for significant decisions (`docs/adr/`)
4. **Runbooks** for ops-critical paths (`docs/runbooks/`)
5. **CHANGELOG.md** in Keep a Changelog format
6. **Onboarding** section (in README or `docs/onboarding.md`)

If a project lacks 3+ of these, documentation work is owed before further feature work — flag it.

## README discipline

The README must answer four questions in the first screen:

1. **What is this?** (one paragraph, no marketing voice)
2. **Why does it exist?** (problem it solves; what it replaces, if anything)
3. **How do I run it?** (prerequisites + setup commands, copy-pasteable)
4. **Where do I go next?** (links to architecture, ADRs, runbooks)

**Example — good opener**:
```markdown
# auth-service

Issues and validates JWTs for the platform. Replaces the legacy session store.

## Prerequisites
- Go 1.22+
- Postgres 15+ (connection string in `DATABASE_URL`)
- A signing key in `JWT_SIGNING_KEY` (32 bytes, base64)
```

**Forbidden**:
- Marketing voice ("Welcome to our awesome project!")
- "See the code" as documentation
- Empty sections kept "for structure"
- Status badges as the dominant content

## Architecture overview

Every non-trivial project has an architecture section that gives a reader the mental map in 5 minutes:

- **Components** — top-level modules/services and what each owns
- **Data flow** — how a request / message / event traverses the system
- **External dependencies** — databases, queues, third-party APIs
- **One diagram** — Mermaid or PlantUML, in version control, updated with the code

If the diagram disagrees with the code, the diagram is wrong. Update it in the same PR as the code change.

## ADRs (Architecture Decision Records)

Write an ADR when a decision:
- Changes the shape of the system (chose X over Y for a foundational concern)
- Locks in a tradeoff future contributors would otherwise relitigate
- Reverses an earlier ADR (link both)

Format (`docs/adr/NNNN-short-title.md`):
```markdown
# ADR-0042: Use Postgres over MySQL for the auth store

## Status
Accepted — 2026-03-12

## Context
Need a relational store for user/session data. Team has ops experience with both.

## Decision
Postgres 15. JSONB columns for flexible profile data, row-level security for tenant isolation.

## Consequences
- Easier multi-tenant isolation via RLS
- Drops MySQL-specific tooling we had elsewhere
- Onboarding doc must include Postgres setup steps
```

Skip the ADR for routine choices (library upgrades, formatting, single-file refactors).

## Runbooks

For any system that is ops-critical (paged on failure, customer-facing, money-handling), `docs/runbooks/` must answer:

- **How to deploy** (commands, expected output, rollback procedure)
- **How to rollback** (specifically — not "reverse the deploy")
- **Common failure modes** with diagnostic commands and fix steps
- **Who owns this** (team, oncall rotation, escalation)

Runbook quality bar: a tired oncall at 3 AM finds the answer in under 5 minutes without reading code.

## CHANGELOG

Keep a Changelog format. Every user-facing change gets one line. Internal refactors do not — they live in git.

## Onboarding

A `docs/onboarding.md` or README section that a new dev can follow top-to-bottom on day one:

- Tools to install (with exact versions or version managers)
- Repo clone + setup commands (verbatim)
- How to run tests
- How to run the app locally
- First small task suggestions (good first issues, sandbox scripts)

## Diagrams

- Text-based tools only (Mermaid, PlantUML) so diagrams live in version control and diff cleanly
- Place in `docs/` and reference from the README
- Update with the code, in the same PR

## Update discipline

- Touch the relevant doc in the **same PR** as the code change. Stale docs > missing docs.
- Removing a feature → remove its docs. Renaming a component → rename all its references.
- Doc drift discovered mid-task → fix it inline; do not file a follow-up.

## What this rule is NOT for

- Inline code comments — see @.claude/rules/code-comments.md
- API endpoint documentation — see @.claude/rules/api.md (OpenAPI/Swagger annotations)
- Test naming conventions — see @.claude/rules/testing.md
