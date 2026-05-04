## Verification Before Completion (non-negotiable)

Never claim a task is done, complete, shipped, or working without **executing proof**.

### Protocol

1. **Identify** the verification command (test suite, build, curl, browser check, etc.)
2. **Execute** the command — do not skip, do not guess the output
3. **Read** the COMPLETE output — not just the exit code, not just the last line
4. **Verify** the output matches the expected result — green tests, 200 status, correct behavior
5. **THEN and only then** declare completion

### Red Flags — You Are About to Lie

| Thought | What to do instead |
|---------|-------------------|
| "This should work based on the code I wrote" | Run the test. "Should" is not proof. |
| "The logic looks correct" | Execute it. Correct logic with a typo still fails. |
| "I'll just say it's done and move on" | Stop. Run the verification. |
| "The test passed earlier, this change is minor" | Run it again. Minor changes break things. |
| "I'm confident this works" | Confidence is not evidence. Execute the command. |
| "It compiled, so it works" | Compilation ≠ correctness. Run the tests. |

### Forbidden Language (before verification)

Never use these phrases unless you have execution output proving them:

- "Done", "Livré", "Complete", "Shipped", "Implemented"
- "This works", "This should work", "This will work"
- "Tests pass" (without showing output)
- "Build succeeds" (without showing output)
- "The feature is ready"

### Partial Completion

If 3 out of 5 items are done, say exactly that:
> "3/5 done — items X and Y still need [specific thing]."

Never round up. Never imply completion. State facts.

### Applies To

- Story completion reports to the lead
- Task status updates
- PR descriptions
- Any claim about code behavior
