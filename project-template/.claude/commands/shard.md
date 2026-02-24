---
name: shard
description: Document sharding — split a large document into indexed, navigable sections
---

You are a **document organizer**. Split a large document into smaller, well-indexed sections for easier navigation and context management.

## Stage 1: Identify Source

$ARGUMENTS must be a file path to the document to shard: $ARGUMENTS

If no arguments, ask: "Which file should I shard? Provide the path."

Read the file and determine its format.

## Stage 2: Parse and Split

**Markdown files** (`.md`):
- Split on `## ` (h2) boundaries
- Each section becomes its own file
- Content before the first `## ` becomes `_intro.md`
- Promote the section heading to `# ` (h1) in each shard

**YAML files** (`.yaml`, `.yml`):
- Split on top-level keys
- Each top-level key becomes its own file
- Preserve the key as a comment header in each shard

**Other formats**:
- Inform the user: "Sharding supports Markdown and YAML files. Convert to one of these formats first."

## Stage 3: Generate Shards

Derive a base name from the source file (e.g., `architecture` from `architecture.yaml`).

Create output directory: `.claude/output/shards/{base-name}/`

Generate:

**`index.md`** — Table of contents:
```markdown
# {Original filename} — Shards Index

Source: `{original-path}`
Sharded: {date}
Sections: {count}

| # | Section | File | Size |
|---|---------|------|------|
| 1 | {section title} | [{slug}.md](./{slug}.md) | {lines} lines |
| 2 | ... | ... | ... |
```

**`{section-slug}.md`** — Each section:
- Slugified from the section title (lowercase, hyphens, no special chars)
- Contains the section content with heading promoted to `# `
- Includes a navigation footer: `← Previous | [Index](./index.md) | Next →`

## Stage 4: Report

Present the results:
```
## Sharding Complete

**Source**: {path}
**Output**: `.claude/output/shards/{base-name}/`
**Sections**: {count}
**Files created**: {count + 1} (including index)

| Section | File | Lines |
|---------|------|-------|
| ... | ... | ... |
```

Ask for confirmation before saving.
