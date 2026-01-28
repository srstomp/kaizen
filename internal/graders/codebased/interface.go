package codebased

// GradeInput contains context for code-based grading
type GradeInput struct {
	TaskID       string   `json:"task_id"`
	TaskType     string   `json:"task_type"` // feature, bug, test, spike, chore
	ChangedFiles []string `json:"changed_files"`
	WorkDir      string   `json:"work_dir"`
}

// GradeResult is the output from a code-based grader
type GradeResult struct {
	GraderName string  `json:"grader_name"`
	Passed     bool    `json:"passed"`
	Score      float64 `json:"score"`   // 0-100
	Details    string  `json:"details"` // Human-readable details
	Skipped    bool    `json:"skipped"` // true if grader not applicable
	SkipReason string  `json:"skip_reason"`
}

// CodeGrader interface for code-based evaluations
type CodeGrader interface {
	Name() string
	Grade(input GradeInput) GradeResult
	IsApplicable(input GradeInput) bool
}
