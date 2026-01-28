# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for the Kaizen project.

ADRs document significant architectural decisions, their context, and consequences. They serve as a historical record of why the system is designed the way it is.

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](001-dual-grader-architecture.md) | Dual-Grader Architecture (Code-Based + Model-Based) | Accepted | 2026-01-28 |

## ADR Template

When creating a new ADR, use this structure:

```markdown
# ADR-NNN: Title

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-NNN
**Deciders:** Names or roles

## Context
What is the issue that we're seeing that motivates this decision?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or harder because of this change?

### Positive
- ...

### Negative
- ...

## Alternatives Considered
What other options were evaluated?

## References
Links to relevant resources
```

## Status Definitions

- **Proposed**: Under discussion, not yet decided
- **Accepted**: Decision made and in effect
- **Deprecated**: No longer applies to new work
- **Superseded**: Replaced by a newer ADR
