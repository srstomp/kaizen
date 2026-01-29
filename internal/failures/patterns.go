package failures

// ConfidenceLevel represents the confidence level based on occurrence count
type ConfidenceLevel string

const (
	ConfidenceLow    ConfidenceLevel = "low"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceHigh   ConfidenceLevel = "high"
)

// Action represents the recommended action based on confidence level
type Action string

const (
	ActionLogOnly    Action = "log-only"
	ActionSuggest    Action = "suggest"
	ActionAutoCreate Action = "auto-create"
)

// Confidence represents the confidence score and recommended action
type Confidence struct {
	Level           ConfidenceLevel
	Action          Action
	OccurrenceCount int
}

// CalculateConfidence determines the confidence level and recommended action
// based on the occurrence count.
//
// Thresholds:
//   - 5+ occurrences: high confidence -> auto-create
//   - 2-4 occurrences: medium confidence -> suggest
//   - 0-1 occurrences: low confidence -> log-only
func CalculateConfidence(count int) Confidence {
	conf := Confidence{
		OccurrenceCount: count,
	}

	switch {
	case count >= 5:
		conf.Level = ConfidenceHigh
		conf.Action = ActionAutoCreate
	case count >= 2:
		conf.Level = ConfidenceMedium
		conf.Action = ActionSuggest
	default:
		conf.Level = ConfidenceLow
		conf.Action = ActionLogOnly
	}

	return conf
}
