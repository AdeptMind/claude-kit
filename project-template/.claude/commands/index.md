---
name: index
description: Index sources into the knowledge graph — incremental, supports local dirs and S3
---

Index local directories or S3 buckets into the knowledge graph using CocoIndex.

## Usage

`/index [source]` — indexes the given source (default: current directory).

## Instructions

1. **Check dependencies** — ensure CocoIndex and Model2Vec are ready:
   ```bash
   which cocoindex 2>/dev/null || pip3 install -U 'cocoindex[sqlite]'
   ls ~/.claude-kit/models/potion-code-16M/model.safetensors 2>/dev/null || ck model download
   ```

2. **Run the index** with the user's source ($ARGUMENTS, default `.`):
   ```bash
   ck knowledge index $ARGUMENTS
   ```

3. **Report results** — show how many chunks were imported and suggest next steps:
   - If chunks were imported: suggest `/search <query>` to explore the indexed content
   - If no chunks: check if the source path is correct or if patterns need adjusting

## Examples

- `/index` — index current repo
- `/index ./src` — index only the src directory
- `/index s3://my-bucket/configs/` — index an S3 bucket
- `/index . --full` — force full re-index (ignore incremental cache)
