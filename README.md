```
 _  __     _
| |/ /__ _(_)_______ _ __
| ' // _` | |_  / _ \ '_ \
| . \ (_| | |/ /  __/ | | |
|_|\_\__,_|_/___\___|_| |_|

  Continuous improvement for AI agents
```

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

Evaluation framework for AI agents. Grade skills, evaluate task quality, run meta-evaluations, and capture failure cases for continuous improvement.

Originally developed as yokay-evals for [pokayokay](https://github.com/stevestomp/pokayokay), now a standalone framework usable by any agentic system.

## Why Kaizen?

AI agents are powerful but inconsistent. Kaizen provides the infrastructure to measure and improve agent reliability:

- **Pre-task gates** - Validate task quality before agents start work
- **Post-task grading** - Verify implementations meet requirements
- **Consistency metrics** - Pass^k measures how often agents succeed *every* time, not just sometimes
- **Failure tracking** - 12 categories of documented failure patterns to learn from
- **CI/CD integration** - Quality gates for your deployment pipeline

## Installation

```bash
go install github.com/srstomp/kaizen/cmd/kaizen@latest
```

Or build from source:

```bash
git clone https://github.com/srstomp/kaizen.git
cd kaizen
go build -o bin/kaizen ./cmd/kaizen
```

## Quick Start

```bash
# Check task quality before starting work
kaizen grade-task-quality \
  --task-id "task-123" \
  --task-title "Add user authentication" \
  --task-type "feature" \
  --description "Implement login/logout functionality"

# Run code-based graders on changed files
kaizen grade-task \
  --task-id "task-123" \
  --changed-files "src/auth.go,src/auth_test.go" \
  --format json

# Run meta-evaluations on agents
kaizen meta --suite agents --k 5
```

## CLI Reference

### Commands

| Command | Description |
|---------|-------------|
| `grade-skills` | Grade skills for clarity and completeness |
| `grade-task` | Run code-based graders (file-exists, test-exists) on task changes |
| `grade-task-quality` | Evaluate task metadata quality before work begins |
| `meta` | Run meta-evaluations on agents or skills |
| `eval` | Run eval suite against failure cases |
| `report` | View and analyze evaluation reports |

### grade-skills

Grade skills and generate a clarity report.

```bash
kaizen grade-skills [options]

Options:
  --skills-dir    Path to skills directory
  --output        Output report path (default: reports/skill-clarity-YYYY-MM-DD.md)
```

### grade-task

Run code-based graders on task changes after implementation.

```bash
kaizen grade-task [options]

Options:
  --task-id        Task ID
  --task-type      Task type: feature, bug, test, spike, chore (default: feature)
  --changed-files  Comma-separated list of changed files
  --work-dir       Working directory (default: .)
  --format         Output format: json, text (default: json)
```

**Graders:**
- `file-exists` - Verifies changed files exist in working directory
- `test-exists` - Checks that code files have corresponding test files

### grade-task-quality

Evaluate task quality based on metadata (pre-task gate).

```bash
kaizen grade-task-quality [options]

Options:
  --task-id              Task ID
  --task-title           Task title
  --task-type            Task type: feature, bug, test, spike, chore
  --description          Task description
  --acceptance-criteria  Acceptance criteria (comma-separated or JSON)
  --min-description-length  Minimum description length (default: 100)
  --format               Output format: json, text (default: json)
```

**Quality Checks:**
- Description length (minimum 100 characters)
- Acceptance criteria presence (required for feature/test/spike)
- Ambiguous keywords detection ("investigate", "explore", "figure out")
- Spike type validation

### meta

Run meta-evaluations to test agent consistency.

```bash
kaizen meta [options]

Options:
  --suite      Suite to run: agents, skills
  --agent      Specific agent to run (e.g., yokay-spec-reviewer)
  --k          Number of runs for pass^k consistency (default: 5)
  --meta-dir   Path to meta directory (default: meta)
```

### eval

Run the eval suite against documented failure cases.

```bash
kaizen eval [options]

Options:
  --failures-dir  Path to failures directory (default: failures)
  --category      Filter to specific category (e.g., missing-tests)
  --k             Number of evaluation runs (default: 1)
  --format        Output format: table, json (default: table)
```

### report

View and analyze evaluation reports.

```bash
kaizen report [options]

Options:
  --type         Report type: grade, eval, all (default: grade)
  --format       Output format: markdown, json (default: markdown)
  --list         List available reports without aggregating
  --output       Write output to file instead of stdout
  --reports-dir  Path to reports directory (default: reports/)
```

## Architecture

```
kaizen/
├── cmd/kaizen/          # CLI implementation
│   └── main.go          # Command definitions
├── internal/
│   ├── graders/
│   │   ├── codebased/   # Code-based graders (file-exists, test-exists)
│   │   └── modelbased/  # LLM-based graders (skill-clarity)
│   ├── metrics/         # Pass@k, Pass^k calculations
│   └── harness/         # Test isolation utilities
├── meta/
│   ├── agents/          # Meta-evaluation test cases for agents
│   ├── skills/          # Meta-evaluation test cases for skills
│   └── schema/          # YAML schemas for eval.yaml files
├── failures/            # Documented failure cases by category
│   ├── missed-tasks/
│   ├── missing-tests/
│   ├── wrong-product/
│   └── ...              # 12 categories total
└── reports/             # Generated evaluation reports
```

### Grader Types

**Code-Based Graders** (`internal/graders/codebased/`)
- Run deterministic checks on code artifacts
- Examples: file existence, test file presence, pattern matching
- Fast, no API calls required

**Model-Based Graders** (`internal/graders/modelbased/`)
- Use LLM for semantic evaluation
- Examples: skill clarity, requirement coverage
- More nuanced but slower and non-deterministic

### Metrics

**Pass@k** - Probability that at least one of k attempts succeeds
- Used for: overall success rate across multiple runs
- Higher is better

**Pass^k (Pass-caret-k)** - Probability that ALL k attempts succeed
- Used for: consistency/reliability measurement
- Measures determinism of agent behavior

## Failure Categories

| Category | Prefix | Description |
|----------|--------|-------------|
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

## Development

### Running Tests

```bash
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Building

```bash
go build -o bin/kaizen ./cmd/kaizen
```

### Adding a New Grader

1. Create grader in `internal/graders/codebased/` or `modelbased/`
2. Implement the grader interface
3. Register in `main.go` command handler
4. Add tests

```go
// Code-based grader interface
type CodeGrader interface {
    Name() string
    Grade(input GradeInput) GradeResult
    IsApplicable(input GradeInput) bool
}
```

## Documentation

- [User Guide](docs/user-guide.md) - Complete guide to using Kaizen
- [Dashboard Data Model](docs/dashboard-data-model.md) - Schema for dashboard visualization
- [Meta-Evaluation Baseline](meta/BASELINE.md) - Agent performance baselines

## License

MIT License - See [LICENSE](LICENSE) for details.
