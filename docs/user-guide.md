# Kaizen User Guide

A practical guide to using Kaizen for evaluating AI agent quality.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Core Concepts](#core-concepts)
3. [Common Workflows](#common-workflows)
4. [Command Reference](#command-reference)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)

---

## Getting Started

### Installation

```bash
# Install via go install
go install github.com/srstomp/kaizen/cmd/kaizen@latest

# Or build from source
git clone https://github.com/srstomp/kaizen.git
cd kaizen
go build -o bin/kaizen ./cmd/kaizen
```

### Verify Installation

```bash
kaizen
```

You should see the available commands:

```
Usage: kaizen <command> [options]

Commands:
  grade               Run a single grader on a single input
  grade-skills        Grade all pokayokay skills and generate report
  grade-task          Run code-based graders on task changes
  grade-task-quality  Evaluate task quality based on metadata
  meta                Run meta-evaluations on agents or skills
  eval                Run eval suite against failure cases
  report              View and analyze evaluation reports
  gate                Check if eval/meta results pass threshold (for CI)
  dashboard           Generate HTML dashboard from eval/meta results
```

---

## Core Concepts

### What Kaizen Measures

Kaizen evaluates AI agents at three stages:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agent Evaluation Pipeline                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  BEFORE WORK           DURING WORK           AFTER WORK         │
│  ───────────           ───────────           ──────────         │
│  grade-task-quality    (agent runs)          grade-task         │
│  "Is this task                               "Did it work?"     │
│   well-defined?"                                                │
│                                                                 │
│                   ACROSS SESSIONS                               │
│                   ───────────────                               │
│                   meta / eval                                   │
│                   "Is the agent                                 │
│                    consistent?"                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Metrics

**Pass@k** - What percentage of attempts succeed?
- Higher is better
- Example: Pass@5 = 80% means 4 out of 5 attempts succeeded

**Pass^k (Pass-caret-k)** - Do ALL attempts succeed?
- Measures consistency/reliability
- Example: Pass^5 = 60% means the agent succeeds all 5 times in 60% of test cases
- This is the metric that matters for production reliability

### Grader Types

| Type | Speed | Cost | Use For |
|------|-------|------|---------|
| Code-based | Fast (<1ms) | Free | File existence, test presence, patterns |
| Model-based | Slow (1-5s) | API costs | Semantic quality, clarity, compliance |

---

## Common Workflows

### Workflow 1: Pre-Task Quality Gate

**Use case**: Validate task definitions before an agent starts working.

```bash
kaizen grade-task-quality \
  --task-id "TASK-123" \
  --task-title "Add user authentication" \
  --task-type "feature" \
  --description "Implement login and logout functionality with JWT tokens. Users should be able to sign in with email/password and receive a token that expires after 24 hours." \
  --acceptance-criteria "Users can log in,Users can log out,Tokens expire after 24h,Invalid credentials show error"
```

**What it checks**:
- Description length (minimum 100 characters)
- Acceptance criteria present (required for feature/test/spike)
- No ambiguous keywords ("investigate", "explore", "figure out")

**Example output**:
```json
{
  "task_id": "TASK-123",
  "passed": true,
  "score": 100,
  "issues": [],
  "suggestion": ""
}
```

**If it fails**:
```json
{
  "task_id": "TASK-456",
  "passed": false,
  "score": 50,
  "issues": [
    {"check": "description_length", "message": "Description too short (45 chars, minimum 100)"},
    {"check": "ambiguous_keywords", "message": "Contains ambiguous keyword 'investigate' - task may be too vague"}
  ],
  "suggestion": "Run /pokayokay:brainstorm to refine task requirements"
}
```

---

### Workflow 2: Post-Task Grading

**Use case**: Verify an agent's implementation after it claims to be done.

```bash
kaizen grade-task \
  --task-id "TASK-123" \
  --task-type "feature" \
  --changed-files "src/auth/login.go,src/auth/logout.go,src/auth/token.go" \
  --work-dir "/path/to/project" \
  --format json
```

**What it checks**:
- `file-exists`: Do all claimed files actually exist?
- `test-exists`: Do code files have corresponding test files?

**Example output**:
```json
{
  "task_id": "TASK-123",
  "timestamp": "2026-01-28T10:30:00Z",
  "results": [
    {
      "grader_name": "file-exists",
      "passed": true,
      "score": 100,
      "details": "All 3 files exist"
    },
    {
      "grader_name": "test-exists",
      "passed": false,
      "score": 33.33,
      "details": "Missing tests: src/auth/logout.go, src/auth/token.go"
    }
  ],
  "overall_passed": false,
  "overall_score": 66.67
}
```

---

### Workflow 3: Agent Consistency Testing

**Use case**: Measure how reliably an agent performs the same task.

```bash
# Run meta-evaluations on all agents, 5 times each
kaizen meta --suite agents --k 5

# Run on a specific agent
kaizen meta --suite agents --agent yokay-spec-reviewer --k 10
```

**What it does**:
1. Loads test cases from `meta/agents/<agent-name>/eval.yaml`
2. Runs each test case `k` times
3. Calculates accuracy and consistency (Pass^k)

**Example output**:
```
Meta-Evaluation Results
=======================

Agent: yokay-spec-reviewer
Test Cases: 15
Runs per test: 5

Results:
  - SPEC-001: 5/5 passed (Pass^5: 100%)
  - SPEC-002: 4/5 passed (Pass^5: 0%)
  - SPEC-003: 5/5 passed (Pass^5: 100%)
  ...

Summary:
  Accuracy: 93.3% (14/15 tests passed at least once)
  Consistency (Pass^5): 86.7% (13/15 tests passed all 5 times)
```

---

### Workflow 4: Failure Case Evaluation

**Use case**: Test agents against documented failure patterns.

```bash
# Run all failure case evaluations
kaizen eval --failures-dir failures --k 3

# Filter to specific category
kaizen eval --failures-dir failures --category missing-tests --k 3 --format json
```

**Failure categories**:

| Category | Prefix | What it catches |
|----------|--------|-----------------|
| missed-tasks | MT | Requirements not implemented |
| missing-tests | WT | No tests for implementation |
| wrong-product | WP | Misunderstood requirements |
| regression | RG | Broke existing functionality |
| premature-completion | PC | Claimed done before complete |
| scope-creep | SC | Extra work beyond spec |
| integration-failure | IF | Integration issues |
| session-amnesia | SA | Lost context between sessions |
| hallucinated-deps | HD | Used non-existent dependencies |
| security-flaw | SF | Security vulnerability |
| tool-misuse | TM | Incorrect tool/API usage |
| task-quality | TQ | Poor task execution quality |

---

### Workflow 5: Skill Documentation Grading

**Use case**: Evaluate the quality of skill documentation.

```bash
kaizen grade-skills \
  --skills-dir /path/to/skills \
  --output reports/skill-clarity-report.md
```

**Grading criteria**:
- Clear Instructions (30%): Are instructions unambiguous?
- Actionable Steps (25%): Can users follow step-by-step?
- Good Examples (25%): Are there helpful examples?
- Appropriate Scope (20%): Is the skill focused?

**Passing threshold**: 70/100

---

### Workflow 6: Report Aggregation

**Use case**: View trends and analyze evaluation history.

```bash
# List available reports
kaizen report --type all --list

# Generate aggregated markdown report
kaizen report --type grade --format markdown --output analysis.md

# View as JSON
kaizen report --type eval --format json
```

---

## Command Reference

### grade

Run a single grader on a single input file.

```bash
kaizen grade --grader <name> --input <path> [--spec <text>] [--format text|json]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--grader` | Yes | Grader name (file-exists, test-exists, skill-clarity, etc.) |
| `--input` | Yes | Path to JSON input file |
| `--spec` | No | Specification text (for model-based graders) |
| `--format` | No | Output format: text (default) or json |

### grade-skills

Grade skill documentation for clarity.

```bash
kaizen grade-skills --skills-dir <path> [--output <path>]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--skills-dir` | Yes | Directory containing SKILL.md files |
| `--output` | No | Report output path (default: reports/skill-clarity-YYYY-MM-DD.md) |

### grade-task

Run code-based graders on changed files.

```bash
kaizen grade-task --task-id <id> --changed-files <files> [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-id` | Yes | Task identifier |
| `--task-type` | No | Type: feature, bug, test, spike, chore (default: feature) |
| `--changed-files` | Yes | Comma-separated list of changed files |
| `--work-dir` | No | Working directory (default: .) |
| `--format` | No | Output format: json (default) or text |

### grade-task-quality

Evaluate task definition quality (pre-task gate).

```bash
kaizen grade-task-quality --task-id <id> --task-title <title> --task-type <type> --description <desc> [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--task-id` | Yes | Task identifier |
| `--task-title` | Yes | Task title |
| `--task-type` | Yes | Type: feature, bug, test, spike, chore |
| `--description` | Yes | Task description |
| `--acceptance-criteria` | No | Comma-separated or JSON criteria |
| `--min-description-length` | No | Minimum description length (default: 100) |
| `--format` | No | Output format: json (default) or text |

### meta

Run meta-evaluations on agents or skills.

```bash
kaizen meta --suite <agents|skills> [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--suite` | Yes | Suite to run: agents or skills |
| `--agent` | No | Specific agent to test |
| `--k` | No | Runs per test case (default: 5) |
| `--meta-dir` | No | Path to meta directory (default: meta) |
| `--confirm` | No | Skip confirmation prompt |

### eval

Run evaluation suite against failure cases.

```bash
kaizen eval [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--failures-dir` | No | Path to failures directory (default: failures) |
| `--category` | No | Filter to specific category |
| `--k` | No | Number of runs (default: 1) |
| `--format` | No | Output format: table (default) or json |

### report

View and analyze evaluation reports.

```bash
kaizen report [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--type` | No | Report type: grade, eval, or all (default: grade) |
| `--format` | No | Output format: markdown (default) or json |
| `--list` | No | List reports without aggregating |
| `--output` | No | Write to file instead of stdout |
| `--reports-dir` | No | Reports directory (default: reports/) |
| `--no-trends` | No | Disable trend analysis |

### gate

Quality gate for CI/CD pipelines.

```bash
kaizen gate [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--type` | No | Check type: eval, meta, or all (default: all) |
| `--threshold` | No | Pass threshold 0-100 (default: 95.0) |
| `--reports-dir` | No | Reports directory (default: reports/) |

Returns exit code 0 if passing, 1 if failing.

### dashboard

Generate HTML dashboard.

```bash
kaizen dashboard [options]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--reports-dir` | No | Reports directory (default: reports/) |
| `--output` | No | Output file (default: dashboard.html) |

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Agent Quality Gate

on:
  pull_request:
    branches: [main]

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Kaizen
        run: go install github.com/srstomp/kaizen/cmd/kaizen@latest

      - name: Run Meta Evaluations
        run: kaizen meta --suite agents --k 3 --confirm

      - name: Quality Gate Check
        run: kaizen gate --type meta --threshold 90.0
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Get changed files
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | tr '\n' ',')

if [ -n "$CHANGED_FILES" ]; then
  kaizen grade-task \
    --task-id "pre-commit" \
    --changed-files "$CHANGED_FILES" \
    --format text

  if [ $? -ne 0 ]; then
    echo "Quality gate failed. Please fix issues before committing."
    exit 1
  fi
fi
```

---

## Troubleshooting

### Common Issues

**"No skill files found"**

Ensure your skills directory contains `SKILL.md` files:
```bash
find /path/to/skills -name "SKILL.md"
```

**"Invalid task type"**

Task type must be one of: `feature`, `bug`, `test`, `spike`, `chore`

**"Description too short"**

Provide more detail in your task description. Minimum is 100 characters by default.

**Meta-evaluation returns 0% Pass^k**

This means the agent is inconsistent - it gives different results on repeated runs of the same test. Check:
1. Is the test case deterministic?
2. Does the agent have non-deterministic behavior?
3. Are external dependencies causing variance?

### Getting Help

- GitHub Issues: https://github.com/srstomp/kaizen/issues
- Documentation: See `docs/` directory
- ADRs: See `docs/adr/` for architectural decisions

---

## Next Steps

1. **Start small**: Use `grade-task-quality` on your next task
2. **Add to CI**: Set up the quality gate in your pipeline
3. **Track failures**: Document failure cases in `failures/` directory
4. **Measure consistency**: Run `meta` evaluations weekly
