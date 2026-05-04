## Anti-Rationalization Protocol

Every discipline rule in this project has enforcement teeth. This rule defines the meta-pattern for recognizing and defeating rationalization — the process by which an agent convinces itself that skipping a rule is acceptable.

### How Rationalization Works

LLMs respond to the same persuasion patterns as humans (Meincke et al. 2025). Under pressure (long context, complex task, time constraints), agents rationalize away discipline rules using predictable patterns:

1. **Minimization** — "This is too simple to need X"
2. **Deferral** — "I'll do X after" / "I'll add X later"
3. **Exception-claiming** — "This case is special"
4. **Confidence substitution** — "I know this works" (without proof)
5. **Scope inflation** — "Doing X would take too long given the task"
6. **Authority appeal** — "The user seems to want speed over X"
7. **Sunk cost** — "I already wrote the code, adding X now would waste work"
8. **Partial credit** — "I did most of X, close enough"

### Universal Red Flags

If any of these thoughts occur during work, STOP and re-read the relevant rule:

| Thought pattern | You are rationalizing... |
|----------------|------------------------|
| "Just this once" | Skipping a non-negotiable |
| "It's obvious that..." | Skipping verification |
| "The user didn't ask for..." | Ignoring a project rule |
| "This would slow me down" | Prioritizing speed over correctness |
| "I already know the answer" | Skipping research/reading |
| "Close enough" | Accepting partial compliance |
| "I'll fix it in the next pass" | Deferring required work |
| "This is a special case" | Inventing an exception |

### Bright-Line Rules

Bright-line rules eliminate judgment calls. When a rule says "always" or "never", there is no exception to evaluate. Examples in this project:

- **TDD**: Write the test BEFORE the code. Always. No "too simple" exception.
- **Verification**: Execute proof BEFORE claiming done. Always. No "I'm confident" exception.
- **PR size**: Max 4-5 files. Always. No "it's all related" exception.
- **Honesty**: State exact completion status. Always. No rounding up.

### For Skill Authors

When writing a discipline skill, include:

1. **Anti-rationalization table** — at least 5 specific excuses with refutations
2. **Red flags section** — thoughts that indicate the agent is about to skip the rule
3. **Bright-line rules** — eliminate gray areas where rationalization thrives
4. **Forbidden language** — specific phrases that signal non-compliance
