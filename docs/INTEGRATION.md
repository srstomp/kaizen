# Kaizen Integration Guide

How kaizen integrates with pokayokay and ohno to create an intelligent task workflow system.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Integration with pokayokay](#integration-with-pokayokay)
3. [Integration with ohno](#integration-with-ohno)
4. [End-to-End Workflow](#end-to-end-workflow)
5. [Configuration](#configuration)
6. [Related Documentation](#related-documentation)

---

## Architecture Overview

Kaizen works as an augmentation layer between pokayokay and ohno, creating a feedback loop for continuous improvement.

```
┌─────────┐      ┌─────────────┐      ┌─────────┐
│  ohno   │◄────►│  pokayokay  │◄────►│ kaizen  │
│         │      │             │      │         │
│ • tasks │      │ • orchestr. │      │ • grade │
│ • deps  │      │ • agents    │      │ • track │
│ • state │      │ • hooks     │      │ • learn │
└─────────┘      └─────────────┘      └─────────┘
     │                  │                   │
     │                  │                   │
     └──────────────────┴───────────────────┘
               Integrated Workflow
```

### Component Roles

| Component | Responsibility | Integration Points |
|-----------|---------------|-------------------|
| **kaizen** | Failure pattern capture, grading, confidence-based suggestions | CLI commands called by pokayokay hooks |
| **pokayokay** | Agent orchestration, review workflow, hook execution | Calls kaizen via `post-review-fail.sh` hook |
| **ohno** | Task management, dependencies, state tracking | Task creation via CLI, queried by pokayokay |

---

## Integration with pokayokay

Kaizen integrates with pokayokay through the **post-review-fail hook**, enabling automatic failure capture and fix task creation.

### How It Works

When a pokayokay review fails (spec-review or quality-review):

```
┌─────────────────┐
│  Review Fails   │
│ (spec/quality)  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ post-review-fail.sh     │
│ ─────────────────       │
│ 1. detect-category      │ ──► kaizen detect-category --details "..."
│ 2. capture              │ ──► kaizen capture --task-id --category --details
│ 3. suggest              │ ──► kaizen suggest --task-id --category
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Action Output          │
├─────────────────────────┤
│ • AUTO                  │ ──► pokayokay auto-creates fix task in ohno
│ • SUGGEST               │ ──► pokayokay prompts user for confirmation
│ • LOGGED                │ ──► pokayokay logs failure, continues normally
└─────────────────────────┘
```

### Kaizen Commands Used

1. **detect-category**: Analyze failure details to classify the issue
   ```bash
   kaizen detect-category --details "Missing test coverage for new API endpoint"
   ```

2. **capture**: Store the failure in kaizen's database for pattern learning
   ```bash
   kaizen capture \
     --task-id "task-123" \
     --category "missing-tests" \
     --details "Missing test coverage..." \
     --source "quality-review"
   ```

3. **suggest**: Get confidence-based action recommendation
   ```bash
   kaizen suggest --task-id "task-123" --category "missing-tests"
   ```

### Action Types

| Kaizen Action | pokayokay Behavior | Trigger Condition |
|---------------|-------------------|------------------|
| `auto-create` → `AUTO` | Automatically creates fix task in ohno | High confidence (5+ occurrences) |
| `suggest` → `SUGGEST` | Prompts user to create fix task | Medium confidence (2-4 occurrences) |
| `log-only` → `LOGGED` | Logs failure for learning, continues | Low confidence (0-1 occurrences) |

### Hook Location

The integration hook is located in pokayokay:
- **Path**: `pokayokay/hooks/post-review-fail.sh`
- **Documentation**: See [pokayokay kaizen integration docs](https://github.com/srstomp/pokayokay/blob/master/docs/integrations/kaizen.md)

### Environment Variables

The hook expects these environment variables (automatically set by pokayokay):

| Variable | Description | Example |
|----------|-------------|---------|
| `TASK_ID` | The ohno task ID that failed review | `task-123` |
| `FAILURE_DETAILS` | Details about why the review failed | `Missing test coverage...` |
| `FAILURE_SOURCE` | Source of failure | `spec-review`, `quality-review` |

---

## Integration with ohno

Kaizen integrates with ohno for fix task creation through the **OhnoClient CLI wrapper**.

### How It Works

When kaizen suggests a fix task with high confidence:

```
┌──────────────────┐
│ kaizen suggest   │
│ returns "AUTO"   │
└────────┬─────────┘
         │
         ▼
┌──────────────────────┐
│ pokayokay receives   │
│ fix task suggestion  │
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ OhnoClient.Create    │
│ Fix Task             │
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ ohno create          │
│ --type fix           │
│ --blocks task-123    │
│ --source kaizen-fix  │
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ Fix task created     │
│ FIX-456              │
└──────────────────────┘
```

### OhnoClient API

Kaizen provides a Go client for creating fix tasks programmatically:

```go
import "github.com/srstomp/kaizen/internal/integration"

client := integration.NewOhnoClient()

taskID, err := client.CreateFixTask(integration.CreateFixTaskParams{
    Title:       "Add tests for task-123",
    TaskType:    "test",
    BlocksTask:  "task-123",
    Description: "Add missing test coverage for API endpoint",
    Source:      "kaizen-fix",
})
```

### Fix Task Parameters

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `Title` | Yes | Task title | `Add tests for task-123` |
| `TaskType` | Yes | Task type | `test`, `fix`, `refactor` |
| `BlocksTask` | Yes | Task ID this fix blocks | `task-123` |
| `Description` | No | Detailed description | `Add missing test coverage...` |
| `Source` | No | Task origin label | `kaizen-fix` |

### Task Types

Kaizen suggests different task types based on failure category:

| Failure Category | Task Type | Example Title |
|-----------------|-----------|---------------|
| `missing-tests` | `test` | `Add tests for task-123` |
| `scope-creep` | `fix` | `Remove scope creep from task-123` |
| `wrong-product` | `fix` | `Fix requirements mismatch in task-123` |
| Other | `fix` | `Fix quality issues in task-123` |

### Dependency Tracking

Fix tasks automatically block the original task:
- Original task: `task-123` (status: in-progress)
- Fix task: `FIX-456` (blocks: task-123)
- Workflow: `task-123` cannot complete until `FIX-456` is resolved

---

## End-to-End Workflow

### Example: Missing Tests Scenario

**Initial State**:
- Agent completes `task-123: Add user authentication`
- Submits for quality review
- Review fails: "Missing test coverage for auth module"

**Step 1: pokayokay calls post-review-fail hook**
```bash
TASK_ID=task-123
FAILURE_DETAILS="Missing test coverage for auth module"
FAILURE_SOURCE="quality-review"

./hooks/post-review-fail.sh
```

**Step 2: kaizen detects category**
```bash
kaizen detect-category --details "Missing test coverage for auth module"
# Output: {"detected_category": "missing-tests", "matched": true}
```

**Step 3: kaizen captures failure**
```bash
kaizen capture \
  --task-id "task-123" \
  --category "missing-tests" \
  --details "Missing test coverage for auth module" \
  --source "quality-review"

# Stored in .kaizen/failures.db or ~/.config/kaizen/failures.db
```

**Step 4: kaizen suggests action**
```bash
kaizen suggest --task-id "task-123" --category "missing-tests"

# If this is the 5th+ occurrence (high confidence):
{
  "action": "auto-create",
  "confidence": "high",
  "fix_task": {
    "title": "Add tests for task-123",
    "description": "Add missing test coverage for auth module",
    "type": "test"
  }
}
```

**Step 5: pokayokay creates fix task in ohno**
```bash
ohno create "Add tests for task-123" \
  --type test \
  --blocks task-123 \
  --source kaizen-fix \
  --description "Add missing test coverage for auth module"

# Output: Created task: TEST-789
```

**Result**:
- `TEST-789` created and blocks `task-123`
- Agent can work on `TEST-789` to add tests
- Once `TEST-789` completes, `task-123` can be re-reviewed
- Pattern reinforced in kaizen database

---

## Configuration

### Storage Location

Kaizen stores failures in SQLite databases:

**Project-local** (recommended for teams):
```bash
cd /path/to/project
kaizen init
# Creates .kaizen/failures.db
```

**Global** (default if not initialized):
```bash
# Uses ~/.config/kaizen/failures.db
```

### Confidence Thresholds

Configure how kaizen determines action confidence:

| Occurrences | Confidence | Action | pokayokay Behavior |
|-------------|------------|--------|-------------------|
| 5+ times | High | `auto-create` | Creates fix task automatically |
| 2-4 times | Medium | `suggest` | Prompts user for confirmation |
| 0-1 times | Low | `log-only` | Logs for learning only |

### Failure Categories

Currently supported categories:

| Category | Description | Fix Task Type |
|----------|-------------|---------------|
| `missing-tests` | No tests for implementation | `test` |
| `scope-creep` | Extra work beyond spec | `fix` |
| `wrong-product` | Misunderstood requirements | `fix` |
| `unknown` | Unclassified failure | `fix` |

### Template Customization

Fix task templates are defined in `failures/templates/*.yaml` (future feature).

---

## Related Documentation

### Kaizen Documentation
- [User Guide](./user-guide.md) - Complete kaizen command reference
- [ADR 001: Dual Grader Architecture](./adr/001-dual-grader-architecture.md) - Grading design
- [Feedback Loop Design](./plans/2026-01-28-feedback-loop-design.md) - System design

### pokayokay Documentation
- [Kaizen Integration Guide](https://github.com/srstomp/pokayokay/blob/master/docs/integrations/kaizen.md) - Detailed integration docs
- [pokayokay Integrations Overview](https://github.com/srstomp/pokayokay/blob/master/docs/integrations/README.md) - All available integrations
- [Hook System](https://github.com/srstomp/pokayokay/blob/master/docs/prompts/hooks.md) - Hook architecture

### ohno Documentation
- [ohno GitHub Repository](https://github.com/srstomp/ohno) - Task management system
- [ohno User Guide](https://github.com/srstomp/ohno/blob/master/README.md) - Getting started

---

## Quick Start

### Prerequisites
```bash
# Install kaizen
go install github.com/srstomp/kaizen@latest

# Verify installation
kaizen

# Initialize in your project (optional)
cd /path/to/project
kaizen init
```

### Test the Integration

Manually test the pokayokay hook integration:

```bash
# Set environment variables
export TASK_ID="test-123"
export FAILURE_DETAILS="Missing test coverage for new feature"
export FAILURE_SOURCE="quality-review"

# Run hook
cd /path/to/pokayokay
./hooks/post-review-fail.sh

# Expected output (first occurrence):
# {"action": "LOGGED"}

# After 5+ similar failures:
# {"action": "AUTO", "fix_task": {...}}
```

### Verify Storage

```bash
# Check if failures are being captured
sqlite3 .kaizen/failures.db "SELECT COUNT(*) FROM failures"

# View failure categories
sqlite3 .kaizen/failures.db "SELECT category, COUNT(*) FROM failures GROUP BY category"
```

---

## Troubleshooting

### kaizen not found
```json
{"action": "LOGGED", "message": "kaizen not installed"}
```

**Solution**: Ensure kaizen is in PATH
```bash
which kaizen
go install github.com/srstomp/kaizen@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### No suggestions returned

Hook always returns `{"action": "LOGGED"}`.

**Possible causes**:
1. Not enough historical data (need 2+ similar failures for suggestions)
2. Category not detected (failure details too generic)
3. Storage issues (check database permissions)

**Debug**:
```bash
# Test category detection
kaizen detect-category --details "Your failure message"

# Check database exists
ls -la .kaizen/failures.db

# Test suggestion directly
kaizen suggest --task-id "test" --category "missing-tests"
```

### ohno CLI not found

When creating fix tasks fails:

**Solution**: Ensure ohno is installed
```bash
which ohno
npm install -g @stevestomp/ohno-cli
```

---

## Contributing

To add new failure categories or improve detection:

1. Add category patterns to `internal/failures/patterns.go`
2. Define fix task template in `failures/templates/*.yaml`
3. Update confidence thresholds in `internal/failures/store.go`
4. Run tests: `go test ./...`
5. Update documentation

For questions or issues:
- GitHub Issues: https://github.com/srstomp/kaizen/issues
- See also: [pokayokay integrations](https://github.com/srstomp/pokayokay/blob/master/docs/integrations/README.md)
