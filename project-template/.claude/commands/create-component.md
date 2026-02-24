---
name: create-component
description: BMad Builder — create a new agent, command, skill, or rule from templates and existing patterns
---

You are a **component builder**. Create a new BMAD component (agent, command, skill, or rule) by studying existing components of the same type and generating a properly structured file.

## Stage 1: Parse Arguments

$ARGUMENTS format: `<type> <name> [description]`

Where `type` is one of: `agent`, `command`, `skill`, `rule`

Examples:
- `agent devops-engineer Infrastructure and deployment specialist`
- `command deploy Automate deployment to staging/production`
- `skill api-validator Validate API contracts against OpenAPI spec`
- `rule performance Performance optimization guidelines`

If $ARGUMENTS is provided: $ARGUMENTS

If no arguments or incomplete, ask:
1. "What type of component? (agent / command / skill / rule)"
2. "What name? (kebab-case, e.g., `my-component`)"
3. "Brief description?"

## Stage 2: Study Existing Patterns

Read 2-3 existing components of the same type for pattern reference:

- **Agents**: Read from `.claude/agents/` — study structure (role, tools, skills, principles, execution sequence)
- **Commands**: Read from `.claude/commands/` — study structure (frontmatter, stages, prerequisites, output)
- **Skills**: Read from `.claude/skills/` — study structure (purpose, inputs, outputs, steps)
- **Rules**: Read from `.claude/rules/` — study structure (categories, guidelines, examples)

Identify common patterns: frontmatter format, section structure, naming conventions, cross-references.

## Stage 3: Gather Requirements

Based on the component type, ask targeted questions:

### Agent
- What is this agent's primary role and expertise?
- What tools does this agent need access to? (Bash, Read, Write, Edit, Glob, Grep, etc.)
- What skills should this agent reference?
- What principles govern this agent's behavior?
- Does this agent work solo or as part of a team?

### Command
- What are the stages of this command? (3-5 typical)
- What prerequisites does it need? (prior artifacts, codebase state)
- What output artifacts does it produce? (files in `.claude/output/`)
- Does it need user confirmation gates?
- Does it accept $ARGUMENTS?

### Skill
- Delegate to the existing skill-creator pattern if available in `.claude/skills/`
- Otherwise: What inputs does this skill take?
- What output does it produce?
- What are the step-by-step instructions?

### Rule
- What category does this rule belong to? (code-style, testing, security, infrastructure, etc.)
- What are the key guidelines? (5-10 concrete rules)
- Are there anti-patterns to call out?
- Does it reference other rules?

## Stage 4: Generate Component

Create the component file in the correct location:

| Type | Location | Format |
|------|----------|--------|
| Agent | `.claude/agents/{name}.md` | Role definition with tools, skills, principles |
| Command | `.claude/commands/{name}.md` | YAML frontmatter + staged instructions |
| Skill | `.claude/skills/{name}.md` | Skill definition with inputs/outputs/steps |
| Rule | `.claude/rules/{name}.md` | Categorized guidelines |

Follow the exact patterns observed in existing components of the same type.

## Stage 5: Validate and Save

Present the generated component to the user.

Check:
- Frontmatter is valid (name, description)
- Structure matches existing components of the same type
- Cross-references to other components are valid
- No placeholder content remains

Ask for confirmation before saving.

If the component is a command, suggest: "Add this command to your CLAUDE.md command reference table."
