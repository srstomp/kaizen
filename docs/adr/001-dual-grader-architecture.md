# ADR-001: Dual-Grader Architecture (Code-Based + Model-Based)

**Date:** 2026-01-28
**Status:** Accepted
**Deciders:** Kaizen Core Team

## Context

Kaizen needs to evaluate AI agent outputs across multiple dimensions:

1. **Structural correctness** - Did the agent create the required files? Are tests present?
2. **Semantic quality** - Does the implementation match the requirements? Is the skill documentation clear?

The challenge: structural checks are fast and deterministic, but miss nuance. Semantic evaluation captures intent but is slow, expensive, and non-deterministic.

### Forces

- **Speed**: CI/CD pipelines need fast feedback (< 1s per check)
- **Cost**: LLM API calls cost money; running thousands of evaluations adds up
- **Accuracy**: Simple pattern matching misses semantic errors; only humans (or LLMs) catch "technically correct but wrong approach"
- **Determinism**: Flaky tests erode trust; consistency metrics require reproducible results
- **Coverage**: Some criteria can only be checked with code (file exists), others only with understanding (is this clear?)

## Decision

Implement a **dual-grader architecture** with two distinct grader types:

### Code-Based Graders (`internal/graders/codebased/`)

Deterministic evaluators that run structural checks:

```go
type CodeGrader interface {
    Name() string
    Grade(input GradeInput) GradeResult
    IsApplicable(input GradeInput) bool
}
```

**Characteristics:**
- No external API calls
- Sub-millisecond execution
- 100% reproducible results
- Binary pass/fail for most checks

**Examples:**
- `file-exists` - Verify expected files were created
- `test-exists` - Check test files accompany implementation
- `endpoint-exists` - Validate API routes are defined

### Model-Based Graders (`internal/graders/modelbased/`)

LLM-powered evaluators for semantic assessment:

```go
type Grader interface {
    Grade(input GradeInput) (Result, error)
}
```

**Characteristics:**
- Requires Anthropic API (Claude)
- 1-5 second execution per check
- Non-deterministic (run multiple times for consistency)
- Scored 0-100 with detailed feedback

**Examples:**
- `skill-clarity` - Evaluate skill documentation for clarity and completeness
- `spec-compliance` - Verify implementation matches requirements
- `task-quality` - Assess task definition quality before work begins

### Execution Model

```
┌─────────────────────────────────────────────────────────────┐
│                     Evaluation Pipeline                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Input ──► Code-Based Graders ──► Fast Pass/Fail            │
│                    │                                        │
│                    │ (if passes basic checks)               │
│                    ▼                                        │
│            Model-Based Graders ──► Semantic Score           │
│                    │                                        │
│                    ▼                                        │
│              Final Result                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Code-based graders run first as a filter. If files don't exist, there's no point asking an LLM if the implementation is correct. This saves API costs and provides faster feedback for obvious failures.

## Consequences

### Positive

- **Fast feedback loop**: Code-based graders give instant results for structural issues
- **Cost efficiency**: Only invoke LLM when structural checks pass
- **Deterministic baseline**: Pass^k metrics are meaningful for code-based graders
- **Flexibility**: Teams can choose which graders to run based on their needs
- **Extensibility**: Easy to add new graders of either type

### Negative

- **Two interfaces**: Developers must understand both grader types
- **Complexity**: Pipeline logic needs to coordinate between grader types
- **Inconsistency**: Model-based graders may give different scores on the same input
- **Maintenance**: Two codepaths to maintain and test

### Mitigations

- **Unified input**: Both grader types use the same `GradeInput` structure
- **Consistency metrics**: Run model-based graders k times and report Pass^k
- **Clear documentation**: Each grader documents its type and behavior
- **Fallback heuristics**: Model-based graders have heuristic implementations for offline/testing use

## Alternatives Considered

### 1. LLM-Only Evaluation

Run all evaluations through Claude with detailed prompts.

**Rejected because:**
- Too slow for CI/CD (3-5s per check)
- Too expensive at scale (thousands of daily evaluations)
- Non-deterministic makes Pass^k meaningless
- Overkill for binary checks like "does file exist?"

### 2. Code-Only Evaluation

Pure rule-based evaluation with regex, AST parsing, etc.

**Rejected because:**
- Cannot assess semantic quality ("is this explanation clear?")
- Would miss "technically correct but wrong" implementations
- Skill clarity cannot be measured without understanding

### 3. Single Unified Interface

One interface with optional LLM flag.

**Rejected because:**
- Muddies the contract (what does `IsApplicable` mean for LLM?)
- Error handling differs significantly (API errors vs pure functions)
- Encourages mixing concerns within a single grader

## Related Decisions

- ADR-002: Pass^k Consistency Metrics (pending)
- ADR-003: Failure Category Taxonomy (pending)

## References

- [OpenAI Evals](https://github.com/openai/evals) - Similar dual approach with basic vs model-graded
- [LMSYS Chatbot Arena](https://chat.lmsys.org/) - Human + LLM evaluation patterns
