# Kaizen Feedback Loop Project

## Overview

Kaizen is an optional augmentation layer for pokayokay that enables:
- Automatic capture of failure patterns
- Confidence-based fix task creation
- Learning from failures over time
- Reduced cost through scoped fix tasks vs full re-dispatch

## Architecture

```
┌─────────┐      ┌─────────────┐      ┌─────────┐
│  ohno   │◄────►│  pokayokay  │◄────►│ kaizen  │
│         │      │             │      │         │
│ • tasks │      │ • orchestr. │      │ • grade │
│ • deps  │      │ • agents    │      │ • track │
│ • state │      │ • hooks     │      │ • learn │
└─────────┘      └─────────────┘      └─────────┘
```

## Tech Stack

- **Language**: Go
- **Storage**: SQLite (global + project-local)
- **Configuration**: YAML templates
- **Integration**: CLI-based (ohno, pokayokay hooks)

## Project Status

- **Epics**: 4
- **Stories**: 13
- **Tasks**: 30
- **Estimated Hours**: 68

## Epics

### P0: Kaizen Core - Failure Capture System (epic-c094d920)
Build the core kaizen feedback loop functionality including commands for capturing failures, detecting categories, suggesting fixes, and SQLite storage.

**Stories**:
- SQLite Failure Storage (story-22706cb8)
- Capture Command (story-10fcd7e0)
- Detect Category Command (story-dc0f7348)
- Suggest Command (story-44579502)
- Fix Task Templates (story-a728e24f)
- Init Command with Bootstrap (story-8d104f4b)

### P1: Pokayokay Integration (epic-595024e6)
Integrate kaizen with pokayokay through the post-review-fail hook.

**Stories**:
- Post-Review-Fail Hook (story-301d92fb)
- Work.md Flow Modification (story-23b230ef)
- Pokayokay Integration Documentation (story-ec98cb4e)

### P1: Ohno Enhancement - Source Tracking (epic-c0a411b6)
Add source field to ohno task schema for task origin tracking.

**Stories**:
- Task Source Field (story-f15c1805)
- Source CLI Flag (story-86153cc4)
- Ohno Documentation Updates (story-a400e50e)

### P2: Cross-Repo Documentation (epic-6a9ad408)
Create and link documentation across kaizen, pokayokay, and ohno repositories.

**Stories**:
- Cross-Repo Documentation Links (story-d4ffd5cd)

## Implementation Order

1. **Phase 1**: Kaizen Core (this repo)
   - Start with: `task-3e71a1ae` (Define SQLite schema)
   - Foundation tasks unblock all other commands

2. **Phase 2**: Pokayokay Integration
   - Depends on: suggest command being complete
   - Start with: `task-b47f03bd` (post-review-fail hook)

3. **Phase 3**: Ohno Enhancement
   - Can run in parallel with Phase 2
   - Start with: `task-e9d6123a` (source field schema)

4. **Phase 4**: Cross-Repo Documentation
   - Final polish after all features complete

## Quick Commands

```bash
# View kanban board
npx @stevestomp/ohno-cli serve

# List tasks
npx @stevestomp/ohno-cli tasks

# Get next task to work on
npx @stevestomp/ohno-cli next

# Start work session
/pokayokay:work
```

## Key Files (to be created)

### Kaizen Core
- `cmd/kaizen/capture.go` - capture command
- `cmd/kaizen/detect.go` - detect-category command
- `cmd/kaizen/suggest.go` - suggest command
- `cmd/kaizen/init.go` - init command
- `internal/failures/store.go` - SQLite storage
- `internal/failures/patterns.go` - pattern detection
- `internal/failures/templates.go` - fix task templates
- `internal/integration/ohno.go` - ohno CLI wrapper
- `failures/templates/*.yaml` - fix task templates

## Confidence Thresholds

| Occurrences | Confidence | Action |
|-------------|------------|--------|
| 5+ | High | `auto-create` |
| 2-4 | Medium | `suggest` |
| 1 | Low | `log-only` |

## Related Documents

- [PRD: Feedback Loop Design](../docs/plans/2026-01-28-feedback-loop-design.md)