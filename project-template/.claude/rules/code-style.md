---
description: Code style and design principles applied when writing or reviewing code
globs: ["src/**", "lib/**", "app/**"]
---

## Design Principles

- **DRY**: Do not duplicate logic; extract shared behavior into reusable functions, modules, or classes
- **KISS**: Choose the simplest solution that satisfies the requirement; avoid premature abstraction
- **SOLID**:
  - Single Responsibility: each class/module/function does one thing
  - Open/Closed: open for extension, closed for modification
  - Liskov Substitution: subtypes must be substitutable for their base types
  - Interface Segregation: prefer small, focused interfaces over large ones
  - Dependency Inversion: depend on abstractions, not concrete implementations
- **Least invasive**: change only what is necessary for the task; do not refactor surrounding code unless asked
- **No over-engineering**: do not add features, config options, or abstractions beyond what is requested
- **Pattern first**: before implementing, scan the codebase for how similar features are done (same entity type, same layer, same framework pattern) and follow that exact pattern — do not invent a new approach when an existing one is established

## Style

- Follow the language and framework conventions of the project
- Use descriptive names for variables, functions, and classes — names should reveal intent
- Keep functions small and focused on a single responsibility
- Avoid deep nesting; prefer early returns to reduce indentation
- Use consistent formatting (indentation, spacing, line length)
- Prefer explicit over implicit behavior
- Prefer composition over inheritance
- No dead code, no commented-out code, no magic numbers (extract constants with descriptive names)
- Separate business logic from data access and presentation

## Error Handling

- Prefer explicit error returns over thrown exceptions where the language supports it (Go `error`, Rust `Result`)
- In exception-based languages, catch specific exception types — never catch and silently swallow errors
- Always propagate enough context for the caller to understand the failure

## File Organization

- Order imports: standard library first, then external packages, then internal modules — separated by blank lines
- If a file exceeds ~300 lines, consider splitting it; a file should have a single clear purpose

## Placement Discovery

Before creating any new file, discover where it belongs by scanning the existing structure:

1. **Find a similar existing file**: search for a file that does the same kind of thing (another controller, another service, another migration, another test) and note its location
2. **Infer the convention**: use that location as the authoritative answer — do not assume a structure (e.g. `src/`, `lib/`, `internal/`, `app/`) without verifying it exists
3. **When in doubt, ask**: if no similar file exists and the structure is ambiguous, ask the user before creating the file

Common patterns vary widely — examples of what to look for:

| What you're adding | Look for an existing… |
|----|-----|
| API handler / controller | `routes/`, `handlers/`, `controllers/`, `api/` |
| Business logic | `services/`, `usecases/`, `domain/`, `core/` |
| Data access | `repositories/`, `store/`, `db/`, `models/` |
| Types / interfaces | `types/`, `interfaces/`, `schema/`, `pkg/` |
| Tests | file next to the source (`_test.go`, `.test.ts`) or in `tests/`, `__tests__/`, `spec/` |
| Infrastructure | `infra/`, `terraform/`, `deploy/`, `k8s/`, `helm/` |
| Config | `config/`, `configs/`, `.env.example`, `settings/` |

Never invent a directory that doesn't exist. If the project has no `services/` folder, don't create one just because it feels right.
