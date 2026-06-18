## Inline Code Comments — Default: None

A descriptive name and clear code structure carry the meaning. Comments are the exception, not the routine.

### Bright-line

If removing the comment would not confuse a future reader, the comment should not exist.

### Allowed (rare, deliberate)

A comment earns its place only when ONE of these is true:

- **Framework hack** — code that bypasses or surprises the framework's normal pattern (`# bypass urllib SSL verification — see issue 'lib-name' #321`)
- **Hidden constraint** — invariant invisible at the call site (`// ordering matters: must run before X loads`)
- **Counter-intuitive behavior** — when the obvious reading is wrong
- **Subtle invariant** — a bug-prone edge case worth flagging
- **`SAFETY:` blocks** — `unsafe` (Rust), unchecked casts, manual lifetime management
- **`TODO(author):`** — planned work with an owner; `HACK:` — intentional workaround that needs cleanup

### Forbidden patterns

- Explaining WHAT the code does — the identifier and the lines already say
- Restating or paraphrasing the function name
- Referencing tickets, PRs, or "added for the X flow" — these rot fast and belong in git/PR history
- Docstrings written "for coverage" or to satisfy linter rules
- Block comments that paraphrase the next 3 lines
- "Header banners" with author/date/changelog — git tracks all of that

### Naming carries the weight

If you feel you need a comment to explain a function, the function name is wrong. Rename it. Same for variables: `cutoff` over `c // cutoff timestamp`, `retryCount` over `r // number of retries`.

### Anti-rationalization

| Excuse | Refutation |
|--------|-----------|
| "Future me will thank present me" | Future you reads the code, not the comment. Make the code clearer. |
| "Reviewers want context" | Put it in the PR description, not the source. |
| "Stack tradition says docstring everything" | The tradition is broken; the codebase should not pay for it. |
| "It's a public function — needs a docstring" | Public functions need clear names + typed signatures + tests, not prose. |
| "Just one comment, harmless" | One becomes ten. The discipline is the rule, not the size. |
| "It's helpful context for newcomers" | Newcomer context belongs in the README / engineering guide, not on line 47. |

### Forbidden language

Never start a comment with:

- "This function..." / "This method..."
- "Helper to..." / "Utility that..."
- "We do X because we need to..." (instead: do X, name it after the need)

### Why this rule exists

A line-level comment burns reader attention every time someone scans the file. Multiply by every code reader, every year. Comments that don't earn their place are net-negative attention drain — they also drift from the code they describe, becoming misleading. Architecture, intent, and "why" live at the system level (see @.claude/rules/documentation.md). The line level is for code.
