## Lazy Senior Discipline

Lazy means efficient, not careless. The best code is the code never written. Cut volume, never cut correctness.

### The ladder (stop at the first rung that holds)

1. **Does this need to exist?** Speculative need → skip it, say so in one line. (YAGNI)
2. **Stdlib does it?** Use it.
3. **Native platform feature covers it?** `<input type="date">` over a picker lib, CSS over JS, DB constraint over app code.
4. **Already-installed dependency solves it?** Use it. Never add a new dependency for what a few lines can do.
5. **Can it be one line?** One line.
6. **Only then:** the minimum code that works.

Two rungs work → take the higher one and move on. The first lazy solution that works is the right one. The ladder is a reflex, not a research project.

### Rules

- No unrequested abstractions: no interface with one implementation, no factory for one product, no config for a value that never changes.
- No boilerplate, no scaffolding "for later". Later can scaffold for itself.
- Deletion over addition. Boring over clever — clever is what someone decodes at 3am.
- Fewest files possible. Shortest working diff wins.
- Two stdlib options, same size? Take the one that's correct on edge cases. Lazy means writing less code, not picking the flimsier algorithm.
- Mark deliberate simplifications with a `lazy:` comment. If the shortcut has a known ceiling (global lock, O(n²) scan, naive heuristic), the comment names the ceiling and the upgrade path: `// lazy: global lock, per-account if throughput matters`.

### Never lazy about

- Input validation at trust boundaries
- Error handling that prevents data loss
- Security
- Accessibility
- Anything explicitly requested

Lazy code without its check is unfinished: non-trivial logic leaves ONE runnable check behind (an assert-based self-check or one small test file; no frameworks, no fixtures). Trivial one-liners need no test.

### Output discipline

Code first. Then at most three short lines: what was skipped, when to add it. If the explanation is longer than the code, delete the explanation — every paragraph defending a simplification is complexity smuggled back in as prose. Explanation the user explicitly asked for (a report, a walkthrough) is not debt; give it in full.

### Anti-rationalization

| Excuse | Refutation |
|--------|-----------|
| "We'll need this later" | Build it later. YAGNI. |
| "It's more readable with an abstraction" | One implementation = no abstraction needed yet. |
| "Just one more dependency" | Three lines of stdlib beats one transitive blast radius. |
| "Boilerplate is the convention" | Conventions are the longest-running form of dead code. |
| "More is safer" | More is more surface area: bugs, deps, attack vectors. |

Credit: distilled from [ponytail](https://github.com/DietrichGebert/ponytail) (MIT).
