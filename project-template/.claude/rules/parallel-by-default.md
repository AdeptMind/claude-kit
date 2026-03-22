---
description: Enforce parallel agent usage for independent tasks — debugging, investigation, builds, research
---

## Parallel by Default

When you encounter any of these situations, **spawn an Agent** instead of doing it sequentially:

- **Build/test failure**: spawn an agent to investigate and fix while you continue with the next task
- **Research needed**: spawn an agent to search/read/explore while you work on something else
- **Independent subtasks**: if two things don't depend on each other, run them in parallel agents
- **Debugging**: spawn an agent to debug the issue — don't iterate step-by-step in the main conversation

This is not optional. Sequential investigation when parallel is possible is a waste of the user's time.

## How to apply

- Use `run_in_background: true` for truly independent work
- Use named agents (`name: "fix-build"`) so you can follow up with `SendMessage`
- Use `isolation: "worktree"` when the agent needs to modify files without conflicts
- Report the agent's result when it completes — don't make the user ask

## Exceptions

- Tasks that require user input mid-execution (ask first, then spawn)
- Tasks where the result determines the next step (foreground agent, not background)
- Simple one-shot commands (a single `git status` doesn't need an agent)
