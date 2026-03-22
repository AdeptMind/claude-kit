---
name: bmad-model
description: BMAD Model phase – design architecture, produce ADRs, and generate a prioritized backlog
---

Act as an **Architect** and **Tech Lead** working together.

Your goal is to take the problem definition from the Break phase and produce an architecture design and implementation backlog.

## Prerequisites

Read `.claude/output/problem.md`. If it does not exist, tell the user to run `/bmad-break` first and stop.

Read `.claude/output/principles.md` if it exists — use it to inform architectural decisions, tech stack choices, and acceptance criteria in the backlog.

## Stage 1: Architecture Design

Based on the problem definition, design the system architecture:

1. **Component breakdown**: Identify the major components/services and their responsibilities.
2. **Data model**: Define core entities, their relationships, and storage strategy.
3. **API surface**: Define the main endpoints or interfaces between components.
4. **Infrastructure**: Define the deployment topology, networking, and cloud services.
5. **Security model**: Authentication, authorization, data protection approach.
6. **Cross-cutting concerns**: Logging, monitoring, error handling, configuration.

Follow code principles: KISS (simplest architecture that meets requirements), no over-engineering (do not add components or layers that are not justified by the requirements).

## Stage 2: Architecture Decision Records

For each significant decision, document the reasoning:

- What was decided and why
- Alternatives considered
- Trade-offs accepted

## Stage 3: Generate Backlog

Break the architecture into an ordered backlog of implementation tasks. Each task must be:

- **Small enough** to implement in a single session
- **Self-contained** with clear inputs and outputs
- **Ordered** by dependency (foundations first, then features, then polish)
- **Labeled** by component and priority

## Stage 4: Produce Output

Create `.claude/output/architecture.md`:

```markdown
# Architecture: <project name>
**Version:** 1.0
**Phase:** model

## Components

### <component name>
- **Type:** service | library | infrastructure | config
- **Responsibility:** <what it does>
- **Tech:** <specific technology>
- **Depends on:** <other component>

## Data Model

### <entity name>
| Field | Type | Constraints |
|-------|------|------------|
| <field> | <type> | <nullable, unique, indexed, etc.> |

**Relationships:**
- <has_many | belongs_to | has_one> → <target entity>

## API Surface
| Method | Path | Component | Description | Auth |
|--------|------|-----------|-------------|------|
| <GET\|POST\|PUT\|DELETE> | <endpoint path> | <which component> | <what it does> | <public \| authenticated \| admin> |

## Infrastructure
- **Compute:** <ECS, Lambda, K8s, etc.>
- **Database:** <RDS, DynamoDB, etc.>
- **Cache:** <ElastiCache, Redis, etc.>
- **Storage:** <S3, etc.>
- **Networking:** <VPC layout, load balancer>
- **CI/CD:** <pipeline description>

## Security
- **Authentication:** <strategy>
- **Authorization:** <strategy>
- **Encryption:** <at rest, in transit>
- **Secrets Management:** <approach>

## ADRs

### ADR-001: <decision title>
- **Decision:** <what was decided>
- **Rationale:** <why>
- **Alternatives:** <options considered>
- **Trade-offs:** <accepted trade-off>
```

Create `.claude/output/backlog.md`:

```markdown
# Backlog: <project name>
**Version:** 1.0
**Phase:** model

## Round 1: <theme>

### T-001: <short title>
- **Component:** <which component>
- **Priority:** P0
- **Type:** setup | feature | integration | test | infra | docs
- **Description:** <what to implement>
- **Depends on:** none
- **Acceptance Criteria:**
  - <criterion 1>
  - <criterion 2>
- **Files to create:** `<path>`
- **Files to modify:** `<path>`
```

## Stage 5: Validate

Present an architecture summary and the backlog to the user. Ask for confirmation before saving. Highlight:

- Key architectural decisions and their trade-offs
- Total number of tasks by priority
- Any assumptions made

Once confirmed, save both files and report completion.

## Next Step

After the backlog is confirmed, the recommended next steps are:

1. Run `/analyze` to check for cross-artifact inconsistencies
2. Run `/checklist` to verify implementation readiness
3. Run `/gsd-prep` to generate codebase mapping and context packs
4. Run `/ralph` (or `/bmad-act`) to begin implementation

Or simply run `/bmad-run` which orchestrates all of this automatically.

## Role-Aware Behavior (po/all-roles only)

> **Role gate**: check the `CK_USER_ROLE` environment variable. If it is `dev` or unset, **skip this section entirely**.

After the Architect produces `architecture.md` (end of Stage 4) and before presenting results to the user (Stage 5), run the following challenge rounds:

### Challenge Round 1 — Tech Lead ADR Review

Act as a **Tech Lead** and review every ADR in `architecture.md`:

1. **Rationale strength**: is the "why" convincing? Would a senior engineer reading this understand the decision without extra context?
2. **Alternatives considered**: are there obvious alternatives missing? Did the Architect evaluate at least two options?
3. **Trade-offs**: are the accepted trade-offs explicit and reasonable? Are there hidden trade-offs not mentioned?

For each issue found, create a challenge entry.

### Challenge Round 2 — SRE Operability Review

Act as an **SRE** and review the architecture for production readiness:

1. **Observability**: does the architecture define logging, monitoring, and tracing strategies? Are there blind spots?
2. **Scaling strategy**: how does each component scale? Are there bottlenecks or single points of failure?
3. **Failover plan**: what happens when a component fails? Is there a degradation strategy?

For each issue found, create a challenge entry.

### Challenge Log Format

Append each challenge to `.claude/output/challenge-log.md` using this format:

```markdown
### Challenge — model — <role> — <ISO-8601 timestamp>
**Question:** <describe the issue or concern>
**Response:** [architect answers]
**Outcome:** <revised | accepted | deferred> — <brief summary of resolution>
```

Where `<role>` is `tech-lead` or `sre`.

### Resolution Gate

The Architect MUST address ALL flagged challenges before proceeding to Stage 5 (Validate). For each challenge:

1. The Architect provides a response explaining the decision or proposing a change
2. If a change is needed, update `architecture.md` accordingly
3. Record the outcome in the challenge entry (`revised`, `accepted`, or `deferred`)

Only proceed to Stage 5 when every challenge has a recorded outcome.

If $ARGUMENTS is provided, use it as additional context: $ARGUMENTS
