## Receiving Code Review — Anti-Sycophancy Protocol

When receiving feedback from a reviewer (human or subagent), respond with technical substance, not social performance.

### Forbidden Responses

Never reply to review feedback with:

- "You're absolutely right!"
- "Great catch!"
- "Good point!"
- "That's a great suggestion!"
- "I completely agree!"

These add zero information. They signal social compliance, not understanding.

### Required Response Pattern

For each review item, respond with:

1. **Technical restatement** — restate the issue in your own words to prove understanding
2. **Assessment** — agree with evidence, or push back with technical reasoning
3. **Action** — what you will change (or why you won't)

Example:
> **Reviewer**: "This function doesn't handle the case where `user` is nil."
>
> **Bad**: "Great catch! I'll fix that right away."
>
> **Good**: "The `GetUser` call on line 42 can return nil when the ID doesn't exist in the store. I'll add a nil check before accessing `user.Email` and return a 404."

### When the Reviewer Is Wrong

Push back. Technical correctness matters more than social comfort.

- If the suggestion would introduce a bug: say so with evidence
- If the suggestion contradicts project conventions: cite the convention
- If the suggestion is based on a misunderstanding: clarify the actual behavior
- If the suggestion is a style preference with no functional impact: say "no change needed — style preference, current code follows project convention"

Do not implement a change you believe is wrong just to avoid disagreement.

### When You're Uncertain

If a review item is unclear:

- **STOP** — do not implement a partial interpretation
- **Ask** — request clarification with a specific question
- **Wait** — do not guess and proceed

One clear question saves more time than two rounds of wrong fixes.

### Applies To

- `/review` output
- `/pr-review` output
- Double review feedback from spec-reviewer and code-quality-reviewer
- Human code review comments
- Any feedback on implementation choices
