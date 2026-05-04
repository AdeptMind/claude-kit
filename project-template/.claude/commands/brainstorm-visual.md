---
name: brainstorm-visual
description: Visual brainstorming — interactive mockups and option cards in the browser
---

Launch a visual brainstorming session with interactive UI in the browser.

## Usage

`/brainstorm-visual [topic]`

## Instructions

1. **Start the server** (in background):
   ```bash
   ck brainstorm-serve &
   ```
   Note the URL (e.g., `http://localhost:XXXXX`). Open it in the browser.

2. **Push content** to the browser. Use `curl` to send HTML fragments with these CSS classes:

   **Options grid** (user clicks to select):
   ```bash
   curl -X POST http://localhost:XXXXX/push -H 'Content-Type: application/json' -d '{
     "type": "html",
     "content": "<div class=\"mockup\"><h2>Choose an approach</h2><div class=\"options\"><div class=\"option\" data-id=\"a\"><h3>Option A</h3><p>Description...</p></div><div class=\"option\" data-id=\"b\"><h3>Option B</h3><p>Description...</p></div></div><button class=\"submit\">Confirm selection</button></div>"
   }'
   ```

   **Pros/cons comparison**:
   ```bash
   curl -X POST http://localhost:XXXXX/push -d '{
     "type": "html",
     "content": "<div class=\"pros-cons\"><div class=\"pros\"><h3>Pros</h3><ul><li>Fast</li><li>Simple</li></ul></div><div class=\"cons\"><h3>Cons</h3><ul><li>Limited</li></ul></div></div>"
   }'
   ```

   **Cards grid**:
   ```bash
   curl -X POST http://localhost:XXXXX/push -d '{
     "type": "append",
     "content": "<div class=\"cards\"><div class=\"card\"><h3>Feature A</h3><p>Details...</p></div></div>"
   }'
   ```

   **Clear the view**:
   ```bash
   curl -X POST http://localhost:XXXXX/push -d '{"type": "clear"}'
   ```

3. **Read user choices** — poll for interaction events:
   ```bash
   curl http://localhost:XXXXX/events
   ```
   Returns JSON array of user clicks/selections since last poll.

4. **Iterate** — based on user choices, push new content. Repeat until the brainstorm converges.

## Available CSS Components

| Class | Purpose |
|-------|---------|
| `.mockup` | Container with border and padding |
| `.options` | Grid of clickable option cards |
| `.option` | Single clickable card (add `data-id` for identification) |
| `.pros-cons` | Two-column pros/cons layout |
| `.cards` | Grid of info cards |
| `.card` | Single info card |
| `.split` | Two-column layout |
| `button.submit` | Confirm selection button |

## Workflow

For $ARGUMENTS topic:

1. Start server and open browser
2. Push 2-4 high-level approach options as `.option` cards
3. Read user selection
4. Push detailed pros/cons for selected approach
5. Push wireframe mockup (HTML) if UI-related
6. Read feedback, iterate
7. Summarize final decisions
