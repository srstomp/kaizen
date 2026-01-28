# Meta-Evaluation Baseline Metrics

**Date**: 2026-01-28
**Command**: `yokay-evals meta --suite agents --k 5`
**Status**: Stubbed (agent execution not yet implemented)

## Summary

| Agent | Test Cases | Accuracy | Consistency (pass^k) | Threshold |
|-------|------------|----------|---------------------|-----------|
| yokay-brainstormer | 12 | 100.0%* | 100.0%* | 95% |
| yokay-quality-reviewer | 15 | 100.0%* | 100.0%* | 95% |
| yokay-spec-reviewer | 12 | 100.0%* | 100.0%* | 95% |

*\*Stubbed results - actual agent execution pending implementation*

## Individual Agent Results

### yokay-brainstormer

Tests the brainstormer agent's ability to refine ambiguous tasks.

**Test Cases by Category:**
- Ambiguous tasks (should be REFINED): BR-001, BR-002, BR-003
- Missing acceptance criteria: BR-004, BR-005
- Too vague (should be NEEDS_INPUT): BR-006, BR-007
- Clear tasks (should be SKIP): BR-008, BR-009
- Edge cases: BR-010, BR-011, BR-012

| ID | Name | Expected | Result | Consistency |
|----|------|----------|--------|-------------|
| BR-001 | Refine vague performance task | REFINED | PASS | 5/5 |
| BR-002 | Refine short feature request | REFINED | PASS | 5/5 |
| BR-003 | Refine ambiguous bug description | REFINED | PASS | 5/5 |
| BR-004 | Add criteria to well-described feature | REFINED | PASS | 5/5 |
| BR-005 | Add criteria to spike task | REFINED | PASS | 5/5 |
| BR-006 | Escalate completely ambiguous request | NEEDS_INPUT | PASS | 5/5 |
| BR-007 | Escalate conflicting requirements | NEEDS_INPUT | PASS | 5/5 |
| BR-008 | Skip well-specified task with criteria | SKIP | PASS | 5/5 |
| BR-009 | Skip narrow bugfix with repro steps | SKIP | PASS | 5/5 |
| BR-010 | Refine task with partial criteria | REFINED | PASS | 5/5 |
| BR-011 | Handle extremely detailed but unclear task | REFINED | PASS | 5/5 |
| BR-012 | Clarify task with business decision needed | NEEDS_INPUT | PASS | 5/5 |

### yokay-quality-reviewer

Tests the quality reviewer's ability to identify code quality, test, and security issues.

**Test Cases by Category:**
- Code smells (should be FAIL): QR-001 to QR-004
- Test smells (should be FAIL): QR-005, QR-006
- Security issues (should be FAIL): QR-007, QR-008, QR-009
- Edge cases - missing handling (should be FAIL): QR-010, QR-011
- Good code (should be PASS): QR-012, QR-013, QR-014, QR-015

| ID | Name | Expected | Result | Consistency |
|----|------|----------|--------|-------------|
| QR-001 | Fail long function without abstraction | FAIL | PASS | 5/5 |
| QR-002 | Fail deeply nested conditionals | FAIL | PASS | 5/5 |
| QR-003 | Fail code with magic numbers | FAIL | PASS | 5/5 |
| QR-004 | Fail duplicated validation logic | FAIL | PASS | 5/5 |
| QR-005 | Fail new feature with no tests | FAIL | PASS | 5/5 |
| QR-006 | Fail tests that only cover happy path | FAIL | PASS | 5/5 |
| QR-007 | Fail hardcoded API keys | FAIL | PASS | 5/5 |
| QR-008 | Fail SQL injection vulnerability | FAIL | PASS | 5/5 |
| QR-009 | Fail XSS vulnerability in rendering | FAIL | PASS | 5/5 |
| QR-010 | Fail missing error handling on async operations | FAIL | PASS | 5/5 |
| QR-011 | Fail missing null/undefined checks | FAIL | PASS | 5/5 |
| QR-012 | Pass well-structured code with comprehensive tests | PASS | PASS | 5/5 |
| QR-013 | Pass well-abstracted code following conventions | PASS | PASS | 5/5 |
| QR-014 | Pass code with comprehensive error handling | PASS | PASS | 5/5 |
| QR-015 | Pass code following project conventions and patterns | PASS | PASS | 5/5 |

### yokay-spec-reviewer

Tests the spec-reviewer agent's ability to verify implementations match task specifications.

**Test Cases by Category:**
- Missing requirements (should be FAIL): SR-001, SR-002, SR-003
- Scope creep (should be FAIL): SR-004, SR-005, SR-006
- Misinterpretation (should be FAIL): SR-007, SR-008
- Complete implementations (should be PASS): SR-009, SR-010, SR-011
- Minimal but sufficient (should be PASS): SR-012

| ID | Name | Expected | Result | Consistency |
|----|------|----------|--------|-------------|
| SR-001 | Fail when acceptance criteria partially met | FAIL | PASS | 5/5 |
| SR-002 | Fail when core functionality missing | FAIL | PASS | 5/5 |
| SR-003 | Fail when error handling omitted | FAIL | PASS | 5/5 |
| SR-004 | Fail when unrequested features added | FAIL | PASS | 5/5 |
| SR-005 | Fail when refactoring unrelated code | FAIL | PASS | 5/5 |
| SR-006 | Fail when adding unrequested optimizations | FAIL | PASS | 5/5 |
| SR-007 | Fail when requirement misunderstood | FAIL | PASS | 5/5 |
| SR-008 | Fail when feature implemented in wrong location | FAIL | PASS | 5/5 |
| SR-009 | Pass complete implementation meeting all criteria | PASS | PASS | 5/5 |
| SR-010 | Pass implementation with all edge cases handled | PASS | PASS | 5/5 |
| SR-011 | Pass well-structured implementation following patterns | PASS | PASS | 5/5 |
| SR-012 | Pass minimal implementation with no extras | PASS | PASS | 5/5 |

## Interpreting Results

### Current Status

These baseline metrics are generated using **stubbed agent execution** - the meta-eval framework returns the expected verdict for each test case. This validates:

1. The eval.yaml test cases are well-formed and parseable
2. The metrics calculation logic works correctly
3. All 39 test cases across 3 agents are ready for real evaluation

### Next Steps

To get meaningful baseline metrics:

1. **Implement agent execution** in `meta.go` `stubAgentExecution()` function
2. **Connect to actual agents** via Task tool or direct agent invocation
3. **Re-run baseline** with real agent responses
4. **Track regression** as agents are modified

### Target Thresholds

All agents have a `consistency_threshold: 0.95` (95%) defined in their eval.yaml files. Once real agent execution is implemented:

- **Accuracy** target: ≥95% correct verdicts
- **Consistency** target: ≥95% of test cases have all k runs agree

### Metrics Definitions

- **Accuracy**: Proportion of test cases where majority verdict matches expected verdict
- **Consistency (pass^k)**: Proportion of test cases where all k runs return the same verdict
- **k=5**: Each test case is run 5 times to measure consistency
