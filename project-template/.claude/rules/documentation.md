---
description: Documentation rules applied when writing or updating docs and README files
globs: ["docs/**", "**/*.md", "README*"]
---

> **Why this matters**: stale or hidden docs are worse than missing docs — they actively mislead. Docs are read when something breaks at 3 AM; optimize for that reader, not the author.

## Structure and Placement

- Keep documentation close to the code it describes — prefer co-located README files over a monolithic docs folder
- Use clear headings, short paragraphs, and bullet points; documentation should be scannable
- Document prerequisites, setup steps, and required environment variables in the project README

**Example — good README opener**:
```markdown
# auth-service

Issues and validates JWTs for the platform. Replaces the legacy session store.

## Prerequisites
- Go 1.22+
- Postgres 15+ (connection string in `DATABASE_URL`)
- A signing key in `JWT_SIGNING_KEY` (32 bytes, base64)
```

**Example — bad README opener** (what to avoid):
```markdown
# auth-service
Welcome to the auth-service repository! This service handles authentication.
```
(no prerequisites, no setup, no purpose differentiation, marketing voice)

## Content Standards

- Include code examples for non-obvious usage; examples should be copy-pasteable and tested
- Update docs when the related code changes — stale docs are worse than no docs
- For API endpoint documentation, follow the standards in `rules/api.md`

## Inline Code Comments

- Comment "why" (intent, constraints, trade-offs), not "what" (the code already says what)
- Use `TODO(author):` for planned work and `HACK:` for intentional workarounds that need cleanup
- Do not leave commented-out code — use version control instead

## Changelog and Decision Records

- Maintain a `CHANGELOG.md` using Keep a Changelog format; update it with every user-facing change
- Record significant architecture decisions in `docs/adr/` using: Title, Status, Context, Decision, Consequences

## Diagrams

- Use text-based diagram tools (Mermaid, PlantUML) so diagrams live in version control
- Place architecture diagrams in `docs/` and reference them from the README
