---
name: bmad-break
description: BMAD Break phase – analyze the problem, clarify requirements, define scope
---

Act as a **Product Owner** and **Requirements Analyst** working together.

Your goal is to take the user's project brief (or existing project context) and produce a structured problem definition with rich, testable user stories.

## Stage 1: Gather Context

1. Check if the user has provided a project brief in the current conversation. If not, ask them to describe their project (see README.md for the brief template).
2. Read any existing `.claude/output/problem.md` to avoid duplicating prior work.
3. Read `.claude/output/principles.md` if it exists — use the project principles to inform quality expectations, testing standards, and architectural constraints.
4. Scan the current codebase (if any) to understand what already exists.

## Stage 2: Analyze and Clarify

Ask the user targeted questions to fill gaps in these areas:

- **Problem statement**: What problem does this solve? Who are the users?
- **Core features**: What are the main capabilities, in priority order?
- **User stories and actors**: Who are the distinct user personas? What are their key workflows? What does success look like from each actor's perspective?
- **Tech stack**: Language, framework, database, cloud provider, deployment target
- **Constraints**: Performance, compliance, security, budget, timeline
- **Integrations**: External APIs, third-party services
- **Non-functional requirements**: Scalability, availability, observability

Do NOT guess or assume answers. Ask the user directly for anything unclear. Keep questions concise and grouped (max 3-5 per round).

## Stage 3: Produce Problem Definition

Once requirements are clear, create `.claude/output/problem.md` with this structure:

```markdown
# Problem Definition: <name>
**Version:** 1.0
**Phase:** break

## Problem Statement
### Summary
<one-sentence description>

### Target Users
- <user type 1>
- <user type 2>

### Pain Points
- <problem 1>
- <problem 2>

## Tech Stack
- **Language:** <e.g., Node.js, Python, Go>
- **Framework:** <e.g., Express, FastAPI, Gin>
- **Database:** <e.g., PostgreSQL, MongoDB>
- **Cloud:** <e.g., AWS, GCP, Azure>
- **Deployment:** <e.g., ECS, Kubernetes, Lambda>
- **CI/CD:** <e.g., GitHub Actions, GitLab CI>

## Features

### Feature: <feature name>
**Priority:** P1
**Description:** <what it does>

**Acceptance Criteria:**
- <criterion 1>
- <criterion 2>

#### User Stories

##### <US-NNN>
- **As a:** <actor/persona>
- **I want:** <capability>
- **So that:** <business value>
- **Priority:** <P1|P2|P3>

**Acceptance Scenarios:**
1. **Given:** <precondition>
   **When:** <action>
   **Then:** <expected outcome>

**Testability:** <how to verify — e.g., "unit test on service layer", "e2e test with mock API">

## Constraints
- **Performance:** <e.g., 500 req/s, <200ms p95>
- **Compliance:** <e.g., GDPR, SOC2, HIPAA>
- **Security:** <requirements>
- **Budget:** <if applicable>

## Integrations

### <service name>
- **Purpose:** <what it's used for>
- **Type:** <REST API | SDK | webhook | message queue>

## Non-Functional Requirements
- **Scalability:** <requirements>
- **Availability:** <SLA target>
- **Observability:** <logging, monitoring, tracing>
```

### User story guidelines

Break each feature into independently testable user stories:

- Each story has a distinct **actor** (`as_a`) — don't default everything to "user". Identify real personas (admin, guest, API consumer, etc.)
- **The `so_that` (WHY) is the most important field.** Every story must justify its existence with clear business value. If you can't articulate why a feature matters, it shouldn't be built. Vague justifications like "so that the system works" or "so that it's better" are not acceptable — push the user to clarify the real value.
- **Priority**: P1 = must-have for MVP, P2 = important but deferrable, P3 = nice-to-have
- **Acceptance scenarios** use Given/When/Then format — each scenario should be specific enough to become a test case
- **Testability** describes the verification strategy — how will you prove this works?
- A feature with no user stories is incomplete. Even technical features (e.g., "database migration") should have a story from the developer or ops persona.

If `.claude/output/principles.md` exists, cross-reference stories against the principles: ensure testing standards, security requirements, and UX principles are reflected in the acceptance scenarios.

## Stage 3b: User Journey (po/all-roles only)

> **Role gate**: check the `CK_USER_ROLE` environment variable. If it is `dev` or unset, **skip this stage entirely** and proceed to Stage 4.

For each distinct persona identified in the user stories, map out end-to-end user journeys that connect multiple stories into coherent flows. Create `.claude/output/user-journey.md` with this structure:

```markdown
# User Journeys

## <journey name>
**Persona:** <persona from user stories>
**Story Refs:** US-NNN, US-NNN

### Steps
1. **Action:** <what the user does>
   **Expected:** <what the user sees or experiences>
   **Screen:** <screen or page identifier>
```

### Guidelines

- Each journey represents a **complete user goal** (e.g., "Sign up and make first purchase"), not a single interaction
- `story_refs` links the journey to the user story IDs defined in `problem.md` — every referenced ID must exist
- Steps describe the **user's perspective**: what they DO (`action`) and what they EXPECT to see (`expected`)
- `screen` is a short identifier for the page or view (e.g., `login-page`, `dashboard`, `checkout-form`) — these will be used as references by `/ux-spec`
- Order steps chronologically within each journey
- A single user story may appear in multiple journeys; a journey must reference at least one story

## Stage 4: Validate

Present a summary of the problem definition to the user and ask for confirmation before saving. Highlight:
- Any assumptions you made
- Features with fewer than 2 user stories (may need more granularity)
- Stories without Given/When/Then scenarios (need acceptance scenarios)

Once confirmed, save to `.claude/output/problem.md` and report completion.

## Stage 4b: Business Value Challenge (po/all-roles only)

> **Role gate**: check the `CK_USER_ROLE` environment variable. If it is `dev` or unset, **skip this stage entirely** and proceed to reporting completion.

After saving `problem.md`, scan every user story's `so_that` field and challenge weak business value statements. A `so_that` is weak if it matches any of these patterns:

- **Generic phrasing**: "so that it works", "so that it's better", "so that the system can...", "so that things are improved", "so that we have it"
- **No specific user outcome**: the value is described in system terms rather than end-user terms (e.g., "so that data is stored" instead of "so that users can retrieve their order history within 2 seconds")
- **Missing WHY**: the statement restates the `i_want` instead of explaining the business reason behind it

For each weak `so_that` detected:

1. Ask the user: **"<US-ID> so_that "<current value>" — What specific outcome does this deliver for the end user?"**
2. Wait for the user's response
3. Update the story's `so_that` in `problem.md` with the revised value
4. Record the exchange in `.claude/output/challenge-log.md` (create with a `# Challenge Log` header if the file does not exist; append if it does):

```markdown
### Challenge — break — <role> — <ISO-8601 timestamp>
**Question:** <US-ID> so_that "<original value>" — What specific outcome does this deliver for the end user?
**Response:** <user's answer>
**Outcome:** revised — updated to "<new so_that value>"
```

If the user defends the original `so_that` with a convincing justification, record the outcome as `kept` instead of `revised`:

```markdown
**Outcome:** kept — "<original value>" (user justification: "<reason>")
```

Process all weak stories before proceeding. If no weak `so_that` values are found, skip this stage silently.

If $ARGUMENTS is provided, use it as the project brief: $ARGUMENTS
