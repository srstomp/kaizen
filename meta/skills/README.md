# Skill Meta-Evaluation Schema

This directory contains meta-evaluations for pokayokay skills, assessing the quality and effectiveness of skill content.

## Overview

Skill evaluations differ from agent evaluations:
- **Agent evals**: Test behavioral outputs (PASS/FAIL verdicts), use `agent:` field
- **Skill evals**: Test content quality using model-based evaluation with weighted criteria, use `skill:` field

## Schema Structure

### Top-Level Fields

```yaml
skill: skill-name                    # Required: Skill identifier (kebab-case)
evaluator: model-based               # Required: Must be "model-based"
criteria:                            # Required: Array of evaluation criteria (4 items)
  - name: "Criterion name"
    check: "Evaluation question"
    weight: 0.25
threshold: 0.8                       # Required: Minimum quality score (0.0-1.0)
```

### Criteria Fields

Each criterion in the `criteria` array must have:

```yaml
- name: "Short criterion name"       # Required: Human-readable criterion name
  check: "Evaluation question?"      # Required: Clear question for model evaluation
  weight: 0.25                       # Required: Weight (must sum to 1.0 across all criteria)
```

## Validation Rules

### Skill Name
- **Pattern**: `^[a-z][a-z-]*[a-z]$`
- **Example**: `api-design`, `testing-strategy`, `database-design`
- Must be lowercase letters and hyphens, cannot start/end with hyphen

### Evaluator
- **Type**: String
- **Required Value**: `model-based`
- Skill evaluations always use model-based grading

### Criteria
- **Type**: Array
- **Length**: Exactly 4 items
- **Weights**: Must sum to 1.0
- Each criterion evaluates a specific aspect of skill quality

### Threshold
- **Type**: Number
- **Range**: 0.0 to 1.0
- **Recommended**: 0.8 (80% quality target)
- Skills scoring below threshold need improvement

## Evaluation Process

1. **Load Skill**: Read the skill's SKILL.md file
2. **Apply Criteria**: Evaluate each criterion using the model
3. **Calculate Score**: Weighted average across all criteria
4. **Pass/Fail**: Compare score to threshold

Example:
```
Criterion 1 (weight 0.3): 0.9 → 0.27
Criterion 2 (weight 0.3): 0.8 → 0.24
Criterion 3 (weight 0.2): 0.7 → 0.14
Criterion 4 (weight 0.2): 0.9 → 0.18
-------------------------------------------
Total Score: 0.83 > 0.8 threshold → PASS
```

## Common Criteria Patterns

### Instructional Clarity
```yaml
- name: "Clear instructions"
  check: "Does the skill clearly explain what to do and when to use it?"
  weight: 0.3
```

### Actionability
```yaml
- name: "Actionable steps"
  check: "Are the steps concrete, executable, and easy to follow?"
  weight: 0.3
```

### Examples and References
```yaml
- name: "Quality examples"
  check: "Does it include helpful, realistic examples?"
  weight: 0.2
```

### Scope and Coverage
```yaml
- name: "Appropriate scope"
  check: "Is the scope well-defined and comprehensive for the domain?"
  weight: 0.2
```

## Domain-Specific Criteria

Different skill types should have criteria tailored to their domain:

### Technical Skills (API Design, Database Design)
- Covers standard patterns and best practices
- Includes anti-patterns and common mistakes
- References are relevant and up-to-date

### Process Skills (Testing Strategy, Error Handling)
- Decision frameworks are clear
- Trade-offs are explained
- Integration with other practices is documented

### Security Skills (Security Audit)
- OWASP coverage is comprehensive
- Remediation guidance is actionable
- Risk assessment is practical

## Example: API Design Skill

```yaml
skill: api-design
evaluator: model-based

criteria:
  - name: "RESTful patterns coverage"
    check: "Does the skill comprehensively cover RESTful API design patterns and conventions?"
    weight: 0.3
  - name: "Practical examples"
    check: "Are the examples realistic, covering common scenarios and edge cases?"
    weight: 0.3
  - name: "Design process clarity"
    check: "Is the design process clearly explained with actionable steps?"
    weight: 0.2
  - name: "Anti-patterns documented"
    check: "Are common mistakes and anti-patterns clearly identified and explained?"
    weight: 0.2

threshold: 0.8
```

## Example: Security Audit Skill

```yaml
skill: security-audit
evaluator: model-based

criteria:
  - name: "OWASP Top 10 coverage"
    check: "Does the skill comprehensively cover OWASP Top 10 vulnerabilities with detection methods?"
    weight: 0.3
  - name: "Actionable remediation"
    check: "Are remediation steps concrete and implementable?"
    weight: 0.3
  - name: "Tool integration"
    check: "Does it include practical tool usage (npm audit, grep patterns, scanners)?"
    weight: 0.2
  - name: "Severity classification"
    check: "Is there a clear severity classification system with appropriate SLAs?"
    weight: 0.2

threshold: 0.8
```

## Directory Structure

```
meta/skills/
├── README.md                        # This file
├── api-design/
│   └── eval.yaml
├── testing-strategy/
│   └── eval.yaml
├── database-design/
│   └── eval.yaml
├── error-handling/
│   └── eval.yaml
└── security-audit/
    └── eval.yaml
```

## Usage

Skill evaluations are run as part of meta-evaluation to ensure skill quality:

```bash
# Evaluate a specific skill
yokay-evals meta --skill api-design

# Evaluate all skills
yokay-evals meta --skills

# Generate skill quality report
yokay-evals meta --skills --report
```

## Interpreting Results

| Score Range | Quality Level | Action |
|-------------|---------------|--------|
| 0.9-1.0 | Excellent | Maintain quality |
| 0.8-0.89 | Good | Minor improvements |
| 0.7-0.79 | Adequate | Review and enhance |
| < 0.7 | Poor | Significant revision needed |

## Improving Low-Scoring Skills

When a skill scores below threshold:

1. **Review failing criteria**: Which specific aspects scored low?
2. **Gather examples**: Look at high-quality documentation in that domain
3. **Add missing content**: Address gaps identified by evaluation
4. **Refine structure**: Improve clarity, add examples, document anti-patterns
5. **Re-evaluate**: Run eval again to verify improvements

## Differences from Agent Evals

| Aspect | Agent Evals | Skill Evals |
|--------|-------------|-------------|
| **Top-level field** | `agent:` | `skill:` |
| **Evaluation type** | Behavioral (PASS/FAIL) | Quality (weighted score) |
| **Test cases** | Multiple test cases with inputs | Single skill document |
| **Evaluator** | Agent execution | Model-based grading |
| **Metrics** | pass^k consistency | Weighted quality score |
| **Input** | Task specs and implementations | Skill content (SKILL.md) |
| **Output** | Verdict (PASS/FAIL/etc) | Numeric score per criterion |

## Schema Validation

The skill eval schema is validated by the meta-evaluation framework when loading eval.yaml files. Validation checks:

- Required fields are present
- Field types are correct
- Criteria weights sum to 1.0
- Threshold is in valid range (0.0-1.0)
- Evaluator is "model-based"

Validation errors will prevent evaluation from running and provide clear error messages.
