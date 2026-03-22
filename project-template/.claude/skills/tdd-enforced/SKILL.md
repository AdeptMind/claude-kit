---
name: tdd-enforced
description: Enforce strict Test-Driven Development (RED-GREEN-REFACTOR) during implementation. Injected into teammate prompts to ensure no production code is written without a failing test first. Includes anti-rationalization table and red flags.
disable-model-invocation: true
allowed-tools: Read, Grep, Glob, Bash
---

### TDD Protocol — Non-Negotiable

**Rule**: NO production code without a failing test FIRST. Code written before a test exists must be deleted and restarted. No exceptions for code stories.

### The Cycle

For each behavior to implement:

1. **RED** — Write ONE minimal test for ONE specific behavior
   - The test MUST fail when you run it
   - It must fail for the RIGHT reason (missing feature, not a typo or import error)
   - Run the test. Confirm it fails. Screenshot or paste the failure output.

2. **GREEN** — Write the MINIMUM code to make the test pass
   - Do not write more code than the test demands
   - Do not add "while I'm here" improvements
   - Run the test. Confirm it passes. All previous tests still pass.

3. **REFACTOR** — Clean up ONLY after green
   - Remove duplication, improve naming, simplify
   - Run tests after every refactoring step
   - If a test breaks during refactor, undo and try again

4. **REPEAT** — Next behavior, next failing test

### Verification at Every Step

You MUST run the test suite after each step:
- After RED: confirm the new test fails (and only the new test)
- After GREEN: confirm all tests pass
- After REFACTOR: confirm all tests still pass

If you skip verification, the cycle is broken. Start over.

### Anti-Rationalization Table

| Excuse | Response |
|--------|----------|
| "This is too simple to test" | Simple code breaks. The test takes 30 seconds to write. |
| "I'll add the test after" | Tests written after code prove nothing — they're designed to pass. |
| "TDD will slow me down" | TDD is faster than debugging. The test IS the specification. |
| "I just need to try something first" | Try it IN a test. Write the test as an experiment. |
| "The framework handles this" | Frameworks have bugs. Test YOUR usage of the framework. |
| "It's just a config change" | Config changes break things. If it can break, test it. |
| "I know this works" | Prove it. Write the test. |
| "The test is obvious" | Then it takes 10 seconds to write. Do it. |
| "I need to see the structure first" | The test IS the structure. Let the test drive the design. |
| "This is just refactoring" | Refactoring requires green tests BEFORE you start changing code. |
| "It's a one-liner" | One-liners have the highest bug-per-line ratio. Test it. |
| "I'll test the integration instead" | Integration tests are slow and coarse. Unit test first, integrate second. |

### Red Flags — Start Over If You See These

- You wrote production code before writing a failing test → **DELETE the code, write the test first**
- Your new test passes immediately without writing new code → The test is wrong or the feature already exists. Investigate.
- You wrote multiple tests at once → Write ONE test at a time
- You're writing tests "for coverage" after implementation → These are not TDD tests. Delete and redo with RED first.
- You haven't run the tests in more than 5 minutes → Run them NOW

### Exceptions

TDD is **optional** for these story types (classified in the implementation plan):
- **Config files**: YAML, JSON, TOML, environment variables
- **Infrastructure**: Terraform, Helm charts, Dockerfiles, CI/CD pipelines
- **Database migrations**: Schema changes, seed data
- **Static assets**: CSS, images, documentation-only changes
- **Markdown templates**: Skill files, agent definitions, prompt templates

For everything else: RED-GREEN-REFACTOR, no exceptions.
