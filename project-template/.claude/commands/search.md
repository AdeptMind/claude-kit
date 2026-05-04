---
name: search
description: Semantic search in the knowledge graph — find relevant nodes by meaning, not just keywords
---

Search the knowledge graph using hybrid search (semantic + keyword).

## Usage

`/search <query>`

## Instructions

1. **Ensure model is available** — check and auto-download if needed:
   ```bash
   ls ~/.claude-kit/models/potion-code-16M/model.safetensors 2>/dev/null || ck model download
   ```

2. **Run hybrid search** with the user's query ($ARGUMENTS):
   ```bash
   ck knowledge search --semantic --limit 10 $ARGUMENTS
   ```

3. **Show results** — if results are found, display them. For the most relevant hit, show full details:
   ```bash
   ck knowledge show <top-node-id>
   ```

4. **No results?** Fall back to keyword-only:
   ```bash
   ck knowledge search --limit 10 $ARGUMENTS
   ```
   If still nothing, suggest `ck knowledge add` to populate the graph.
