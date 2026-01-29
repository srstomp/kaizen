package failures

import "strings"

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

// Category represents a failure category
type Category string

const (
	CategoryMissingTests  Category = "missing-tests"
	CategoryScopeCreep    Category = "scope-creep"
	CategoryWrongProduct  Category = "wrong-product"
)

// CategoryPattern represents a category and its associated patterns
type CategoryPattern struct {
	Category Category
	Patterns []string
}

// categoryPatterns defines the patterns for each category
var categoryPatterns = []CategoryPattern{
	{
		Category: CategoryMissingTests,
		Patterns: []string{"missing test", "no test", "untested"},
	},
	{
		Category: CategoryScopeCreep,
		Patterns: []string{"out of scope", "not in spec", "extra feature"},
	},
	{
		Category: CategoryWrongProduct,
		Patterns: []string{"wrong file", "incorrect implementation"},
	},
}

// DetectCategory detects the first matching category from the given text.
// It performs case-insensitive matching against predefined patterns.
// Returns the matched category and true if a match is found, or empty string and false otherwise.
func DetectCategory(text string) (Category, bool) {
	lowerText := strings.ToLower(text)

	for _, cp := range categoryPatterns {
		for _, pattern := range cp.Patterns {
			if strings.Contains(lowerText, pattern) {
				return cp.Category, true
			}
		}
	}

	return "", false
}

// DetectAllCategories detects all matching categories from the given text.
// It performs case-insensitive matching against predefined patterns.
// Returns a slice of all matched categories. Returns an empty slice if no matches are found.
func DetectAllCategories(text string) []Category {
	lowerText := strings.ToLower(text)
	categories := []Category{}
	matched := make(map[Category]bool)

	for _, cp := range categoryPatterns {
		for _, pattern := range cp.Patterns {
			if strings.Contains(lowerText, pattern) {
				if !matched[cp.Category] {
					categories = append(categories, cp.Category)
					matched[cp.Category] = true
				}
				break
			}
		}
	}

	return categories
}
