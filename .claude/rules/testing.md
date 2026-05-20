---
description: Testing rules applied when generating or reviewing tests
globs: ["tests/**", "test/**", "**/*.test.*", "**/*.spec.*"]
---

## TDD Protocol

All code stories MUST follow the RED-GREEN-REFACTOR cycle. This is non-negotiable.

1. **RED**: Write ONE failing test for ONE behavior. Run it. Confirm it fails for the right reason.
2. **GREEN**: Write the MINIMUM code to pass the test. Run all tests. Confirm they pass.
3. **REFACTOR**: Clean up with all tests green. Run tests after each change.

### Red Flags — Start Over

- Production code written before a failing test exists → **delete the code, write the test first**
- A new test passes immediately → the test is wrong or the feature already exists
- Tests added after implementation "for coverage" → these are not TDD tests

### Anti-Rationalization

| Excuse | Response |
|--------|----------|
| "Too simple to test" | Simple code breaks. The test takes 30 seconds. |
| "I'll test after" | Tests written after code are designed to pass, not to catch bugs. |
| "TDD slows me down" | TDD is faster than debugging. |
| "Just a config change" | If it can break, test it. |
| "I know this works" | Prove it. Write the test. |

### Exceptions

TDD is optional for: config files, infrastructure (Terraform, Helm, Docker), database migrations, static assets, and markdown templates.

---

## Test Design

- Write a failing test before fixing a bug to prove the bug exists and prevent regression
- Cover edge cases and error paths, not just happy paths — test nulls, empty inputs, boundary values
- Use descriptive test names explaining the scenario and expected outcome: `test_returns_404_when_user_not_found`
- Each test must be independent and not rely on execution order or shared mutable state

## Mocking Strategy

- Mock external boundaries (APIs, databases, file systems, clocks) to keep tests fast and deterministic
- Avoid mocking internal code — test real behavior; mocking internals makes tests brittle
- When a test requires complex setup, the code under test likely needs a simpler interface

## Test Data

- Use factories or builders to create test data with sensible defaults; override only what the test cares about
- Do not share mutable test data between tests — each test creates its own state
- Avoid hardcoded fixture files that grow stale; generate test data programmatically

## Test Types and Coverage

- **Unit tests**: test individual functions and classes in isolation; majority of tests
- **Integration tests**: test boundaries between your code and external systems
- **End-to-end tests**: cover critical user flows only; keep count low (slow). See "End-to-End Tests — Real User Simulation" below for the non-negotiable rules.
- Focus coverage on critical paths and complex logic; do not chase arbitrary percentage targets

---

## End-to-End Tests — Real User Simulation (non-negotiable)

**E2E tests MUST simulate what a real human user does.** If a user clicks a button, the test clicks the button. If a user reads output from a terminal, the test reads stdout. Tests that bypass the user-facing surface are NOT E2E tests, regardless of what they are labelled.

### Concrete rules by surface

- **UI / web app**: drive the real interface — clicks, typing into real form fields, navigating between screens (Playwright / Selenium / Cypress against the running app, ideally a real browser). **Never** substitute `curl` against the backend, direct DB queries, or internal API calls "to save time".
- **CLI**: invoke the actual compiled binary as a subprocess and parse its real stdout / stderr / exit code. **Never** import the CLI's internal modules and call functions in-process.
- **Daemon / service**: spawn the actual process, send real signals (SIGTERM, SIGHUP), interact through real config files and real network calls. **Never** mock the signal handler or skip the runtime.
- **Infrastructure**: provision a real (ephemeral) environment when feasible — Docker Compose, testcontainers, kind cluster. The test should be reproducible from a clean machine.

### Litmus test

Before writing an E2E test, ask: **"Would a real user, doing this in production, perform exactly these steps in exactly this order?"**

- If the test calls anything the user never touches (private APIs, internal services, mocked auth) → it is NOT E2E.
- If the test sets up state in a way the user cannot (DB seeds, bypass tokens, internal flags) → the seam being tested is wrong; fix the seam, not the test.

### Tie E2E to user journeys

Each E2E test name corresponds to a named user journey from the product backlog, PRD, or user-flow spec:

```
e2e_user_logs_in_then_lists_their_content_in_dashboard
e2e_admin_imports_csv_through_upload_dialog
e2e_daemon_syncs_team_config_after_install
```

If no journey exists for the behavior under test, write the journey first — then the E2E. Keep E2E count low; depth matters more than breadth.

### Anti-pattern: the "spin-up-and-curl"

- BAD: Start the backend with env vars, run `curl http://localhost:8080/api/login`, assert JSON.
- GOOD: Open the frontend in a real browser, type credentials in the form, click submit, wait for the dashboard, assert what the user actually sees.

The first is an integration test. Calling it E2E creates a false sense of safety — the actual code path that users hit (form rendering, client-side validation, auth flow, redirects, dashboard render) is never exercised, so production-only regressions slip through CI.

### When the real surface is too expensive

If the user-facing flow is genuinely too slow or flaky to be automated end-to-end (e.g. real third-party OAuth provider, payment gateway, vendor portal), document the gap explicitly in the test file and rely on a tighter integration test bundle to compensate — but **do not pretend** an integration test is an E2E.

## CI and Reliability

- All tests must pass before merge; no exceptions
- Fix flaky tests immediately — do not skip, retry, or `@ignore`; flaky tests erode trust
- Keep test setup minimal and close to the assertion

## Function-Test Pairing

- Every non-trivial function must have a corresponding test — use `/test-check` after modifying functions to verify coverage
- When a function's contract changes (signature, return type, behavior), update the test to match the new contract
- **Never update a test just to make it pass** — if the output changed unexpectedly, the function is broken; fix the function, not the test
- Never weaken an assertion (e.g., replacing `assertEqual(x, 42)` with `assertNotNil(x)`) to hide a failure
- If unsure whether a behavior change is intentional, ask before updating the test
