## Spec-Driven Hybrid Workflow

The hybrid SDD/BMAD workflow uses BMAD phases (Break, Clarify) as an entry ramp to build a declarative spec, then executes the Act phase strictly against that spec — no invention, no scope drift.

This rule applies only when the project was initialized with `ck init` choosing **Hybrid SDD/BMAD**. It augments the standard BMAD workflow with three discipline points: spec crystallization, ambiguity surfacing, and drift checking.

### When this rule applies

- After `/bmad-break` and `/clarify` have produced a stable `problem.md`
- Before `/bmad-act` or `/ralph` starts implementation
- The `model` phase must produce a **`spec.md`** (executable specification) in addition to `architecture.md`

### Spec format (minimum viable)

`spec.md` lives next to `architecture.md` in the output directory. It must contain:

1. **Identity** — binary name, language, stack, entry point (3 lines max)
2. **Commands / endpoints / public API** — exhaustive list. If a name is not here, it must not exist in the code.
3. **Data model** — types + persistence format
4. **Cross-cutting rules** — error codes, output discipline, atomicity invariants
5. **File layout** — exact directory tree. Files not listed = drift.
6. **Test naming convention** — agents must name tests so each AC maps 1:1 to a test function

### Three discipline points

#### 1. Spec crystallization (in `/bmad-model`)

The Model phase output is no longer just architecture. It produces a `spec.md` that:

- States the file layout the implementation MUST follow
- Names every public function/command that should exist
- References each AC by ID and the test function name that proves it

If `spec.md` is absent at the end of Model, the workflow stops. No Act phase without a spec.

#### 2. Ambiguity surfacing (in `/bmad-act` and `/ralph`)

When the executing agent reads the spec, if any point is ambiguous (multiple reasonable interpretations), the agent MUST:

1. Write the question to `.claude/output/spec-questions.md`
2. Pick the most **conservative interpretation** (the smallest, least surprising option)
3. Continue execution

The agent does NOT silently choose, invent a new feature, or add scaffolding "for completeness". If `spec-questions.md` is non-empty after the run, the lead reviews each question and updates the spec — then re-runs if needed.

#### 3. Drift check (mandatory after `/bmad-act` and `/ralph`)

Before declaring an Act phase complete, run a drift check that compares actual code to `spec.md`:

- **Commands**: enumerate dispatch cases vs spec command list. Mismatch → drift.
- **File layout**: actual `find . -name '*.go'` (or equivalent) vs spec layout. Extra files = drift. Missing files = drift.
- **Test count**: actual test functions vs AC count from `acceptance-criteria.md`. Mismatch → drift.

Drift is documented in `.claude/output/drift-check.md` with severity (blocker, minor). Blockers prevent the lead from accepting the story.

### Forbidden in hybrid mode

- Implementing a feature, file, or command not in the spec
- Choosing among ambiguous interpretations silently
- Skipping the drift check at the end of Act
- Accepting a story while drift-check.md has unresolved blockers

### Why this exists

Pure BMAD lets the architecture emerge during Act. Pure SDD requires the user to write a tight spec upfront, which most non-PM users find cognitively heavy. The hybrid uses BMAD's Break + Clarify phases to construct the spec collaboratively, then locks the spec for execution.

Result: the user keeps a fuzzy entry point, the IA does not invent, and drift is detected mechanically rather than discovered weeks later in code review.
