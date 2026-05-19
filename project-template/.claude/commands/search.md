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

## Query tips

- **Semantic queries work best with intent, not keywords**: prefer `"how do we handle JWT expiry"` over `"JWT token"`
- **Quote multi-word phrases** when you want exact match: `'"connection pool"'`
- **Combine with `--limit`** to narrow noisy results: `ck knowledge search --semantic --limit 3 ...`
- **Hybrid search (default `--semantic`)** uses RRF (Reciprocal Rank Fusion) — top results usually score on both keyword and semantic similarity

## Examples

- `/search how does authentication work` — semantic query, finds conceptually related code
- `/search "DATABASE_URL"` — exact-match keyword for an env var
- `/search rate limiting in the api gateway` — natural-language query, ranks by relevance across files

## Troubleshooting

- **Empty results for a phrase you know exists**: the index is stale — run `/index` to refresh
- **Results are off-topic**: try a more specific query or drop `--semantic` to fall back to pure keyword
- **`model.safetensors` not found**: step 1 should have caught this; re-run `ck model download`
