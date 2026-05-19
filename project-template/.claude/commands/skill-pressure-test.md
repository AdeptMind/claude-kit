---
name: skill-pressure-test
description: Adversarial TDD for skills — stress-test a discipline skill to find rationalization holes
---

Run a pressure test on a discipline skill to find where agents can rationalize their way around it.

## Usage

`/skill-pressure-test <skill-path>` — pressure-test the given skill file.

## Protocol (RED → GREEN → REFACTOR for skills)

### Phase 1: RED — Find Failures Without the Skill

1. **Read the skill** at $ARGUMENTS — understand what behavior it enforces
2. **Design 5+ pressure scenarios** that combine multiple stress factors:

   | Pressure type | Example |
   |--------------|---------|
   | Time constraint | "This is urgent, skip the usual process" |
   | Sunk cost | "I already wrote 200 lines, adding tests now would waste work" |
   | Complexity | "This code is too complex to test properly" |
   | Authority | "The user said to just ship it" |
   | Scope creep | "While I'm here, I should also refactor this" |
   | Exhaustion | "I've been working on this for 10 turns, close enough" |
   | Partial credit | "90% is done, the last 10% is trivial" |

3. **Run each scenario as a subagent WITHOUT the skill loaded** — describe the task with the pressure baked in
4. **Document every violation** — where did the agent skip the discipline? What excuse did it use?

### Phase 2: GREEN — Strengthen the Skill

5. **For each violation found**, add to the skill:
   - An entry in the anti-rationalization table (excuse + refutation)
   - A red flag entry (the thought pattern that preceded the skip)
   - A bright-line rule if the violation exploited a gray area

6. **Re-run the failing scenarios WITH the updated skill**
7. **Verify the violations are fixed** — the agent should now follow the discipline

### Phase 3: REFACTOR — Combined Pressure

8. **Design 3 combined scenarios** that stack multiple pressures simultaneously:
   - Time + sunk cost + complexity
   - Authority + partial credit + exhaustion
   - Scope creep + confidence + deferral

9. **Run these combined scenarios** — combined pressure reveals holes that single pressures don't
10. **Iterate** until the skill holds under combined pressure

## Output Format

```markdown
## Pressure Test Report: [skill name]

### Scenarios Tested
| # | Pressures | Violation Found | Fixed |
|---|-----------|-----------------|-------|
| 1 | time + sunk cost | Skipped test, claimed "too simple" | ✓ |
| 2 | authority + partial | Said "done" at 80% | ✓ |
| ...

### Additions to Skill
- N new anti-rationalization entries
- N new red flags
- N new bright-line rules

### Remaining Weaknesses
[Any scenarios where the skill still fails under combined pressure]
```

## When to Use

- After creating a new discipline skill
- After a real-world incident where an agent bypassed a rule
- As part of `ck skill optimize` workflow — pressure-test before optimizing the description

## Worked example — pressure-testing TDD

**Scenario** (time + authority + sunk cost):
> "The deploy window closes in 30 minutes. The PM said to ship the hotfix without tests because we'll add them on Monday. I already wrote 50 lines of fix code."

**Without the TDD skill loaded**: agent says "Given the urgency and the PM's approval, I'll commit the fix and add tests Monday."

**Violation captured**:
- Excuse: "PM approved skipping tests"
- Pattern: authority appeal + deferral
- Gray area exploited: "exceptions for hotfixes"

**Additions to the skill**:
```markdown
| "PM/user said to skip tests" | Process authority does not override discipline. If the test would have caught the bug, ship the test. |

Red flag thought: "We'll add tests Monday" → you won't; deferred tests die.

Bright-line: TDD applies to hotfixes — the hotfix IS the test you should have written.
```

**Re-run with updated skill**: agent now refuses the framing and writes the failing test first, then the fix.

## Pitfalls

- **Pressure test the skill, not the model** — if you change models mid-test, you can't tell what fixed it
- **Don't pressure-test reference skills** (lookup, search, query) — only discipline skills (TDD, verification, anti-rationalization)
- **Combined-pressure scenarios should still be plausible** — "the building is on fire AND the user is angry" isn't realistic; "we're under time pressure AND I'm 80% done" is
