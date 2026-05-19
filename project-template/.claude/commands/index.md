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

## When to use `--full`

- After bumping the embedding model version (`ck model download` returns a new digest)
- When chunk-boundary logic has changed in CocoIndex itself
- When you suspect index corruption (search returns nothing or wildly wrong results)

Otherwise, **incremental is the default** — re-running `/index` only processes changed files, which is fast.

## What gets excluded

CocoIndex respects `.gitignore` and skips by default:
- `node_modules/`, `vendor/`, `.git/`, `__pycache__/`, `dist/`, `build/`
- Binary files (images, compiled artifacts, lock files)
- Files over 1 MB unless `--include-large` is set

## Troubleshooting

- **`cocoindex: command not found`**: pipx ran in a different shell — restart your terminal or `pipx ensurepath`
- **`model.safetensors not found`**: `ck model download` was interrupted — re-run it
- **S3 access denied**: verify `AWS_PROFILE` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` are set
- **Zero chunks imported**: the source has no indexable files OR they're all excluded — try `--full` and check the file types
