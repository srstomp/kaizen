package failures

import "testing"

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name          string
		count         int
		wantLevel     ConfidenceLevel
		wantAction    Action
		wantCount     int
	}{
		{
			name:       "zero occurrences - low confidence",
			count:      0,
			wantLevel:  ConfidenceLow,
			wantAction: ActionLogOnly,
			wantCount:  0,
		},
		{
			name:       "one occurrence - low confidence",
			count:      1,
			wantLevel:  ConfidenceLow,
			wantAction: ActionLogOnly,
			wantCount:  1,
		},
		{
			name:       "two occurrences - medium confidence",
			count:      2,
			wantLevel:  ConfidenceMedium,
			wantAction: ActionSuggest,
			wantCount:  2,
		},
		{
			name:       "four occurrences - medium confidence",
			count:      4,
			wantLevel:  ConfidenceMedium,
			wantAction: ActionSuggest,
			wantCount:  4,
		},
		{
			name:       "five occurrences - high confidence",
			count:      5,
			wantLevel:  ConfidenceHigh,
			wantAction: ActionAutoCreate,
			wantCount:  5,
		},
		{
			name:       "ten occurrences - high confidence",
			count:      10,
			wantLevel:  ConfidenceHigh,
			wantAction: ActionAutoCreate,
			wantCount:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConfidence(tt.count)

			if got.Level != tt.wantLevel {
				t.Errorf("CalculateConfidence(%d).Level = %v, want %v", tt.count, got.Level, tt.wantLevel)
			}

			if got.Action != tt.wantAction {
				t.Errorf("CalculateConfidence(%d).Action = %v, want %v", tt.count, got.Action, tt.wantAction)
			}

			if got.OccurrenceCount != tt.wantCount {
				t.Errorf("CalculateConfidence(%d).OccurrenceCount = %v, want %v", tt.count, got.OccurrenceCount, tt.wantCount)
			}
		})
	}
}

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantCat    Category
		wantFound  bool
	}{
		{
			name:      "missing test pattern - lowercase",
			text:      "This code has missing test coverage",
			wantCat:   CategoryMissingTests,
			wantFound: true,
		},
		{
			name:      "missing test pattern - uppercase",
			text:      "MISSING TEST in the module",
			wantCat:   CategoryMissingTests,
			wantFound: true,
		},
		{
			name:      "no test pattern",
			text:      "There is no test for this function",
			wantCat:   CategoryMissingTests,
			wantFound: true,
		},
		{
			name:      "untested pattern",
			text:      "This function is untested",
			wantCat:   CategoryMissingTests,
			wantFound: true,
		},
		{
			name:      "out of scope pattern",
			text:      "This feature is out of scope",
			wantCat:   CategoryScopeCreep,
			wantFound: true,
		},
		{
			name:      "not in spec pattern",
			text:      "This is not in spec",
			wantCat:   CategoryScopeCreep,
			wantFound: true,
		},
		{
			name:      "extra feature pattern",
			text:      "Added an extra feature that wasn't requested",
			wantCat:   CategoryScopeCreep,
			wantFound: true,
		},
		{
			name:      "wrong file pattern",
			text:      "Modified the wrong file",
			wantCat:   CategoryWrongProduct,
			wantFound: true,
		},
		{
			name:      "incorrect implementation pattern",
			text:      "This is an incorrect implementation",
			wantCat:   CategoryWrongProduct,
			wantFound: true,
		},
		{
			name:      "case insensitive - mixed case",
			text:      "We have No TeSt coverage here",
			wantCat:   CategoryMissingTests,
			wantFound: true,
		},
		{
			name:      "no match",
			text:      "Everything looks good",
			wantCat:   "",
			wantFound: false,
		},
		{
			name:      "empty text",
			text:      "",
			wantCat:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCat, gotFound := DetectCategory(tt.text)

			if gotFound != tt.wantFound {
				t.Errorf("DetectCategory(%q) found = %v, want %v", tt.text, gotFound, tt.wantFound)
			}

			if gotCat != tt.wantCat {
				t.Errorf("DetectCategory(%q) category = %v, want %v", tt.text, gotCat, tt.wantCat)
			}
		})
	}
}

func TestDetectAllCategories(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		wantCategories []Category
	}{
		{
			name:          "single category - missing tests",
			text:          "This has missing test coverage",
			wantCategories: []Category{CategoryMissingTests},
		},
		{
			name:          "single category - scope creep",
			text:          "This is out of scope",
			wantCategories: []Category{CategoryScopeCreep},
		},
		{
			name:          "single category - wrong product",
			text:          "Modified the wrong file",
			wantCategories: []Category{CategoryWrongProduct},
		},
		{
			name:          "multiple categories",
			text:          "This has missing test and is out of scope with wrong file changes",
			wantCategories: []Category{CategoryMissingTests, CategoryScopeCreep, CategoryWrongProduct},
		},
		{
			name:          "no matches",
			text:          "Everything is perfect",
			wantCategories: []Category{},
		},
		{
			name:          "empty text",
			text:          "",
			wantCategories: []Category{},
		},
		{
			name:          "duplicate patterns same category",
			text:          "no test and also missing test and untested code",
			wantCategories: []Category{CategoryMissingTests},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectAllCategories(tt.text)

			if len(got) != len(tt.wantCategories) {
				t.Errorf("DetectAllCategories(%q) returned %d categories, want %d", tt.text, len(got), len(tt.wantCategories))
				t.Errorf("  got: %v", got)
				t.Errorf("  want: %v", tt.wantCategories)
				return
			}

			// Check each expected category is present
			for _, want := range tt.wantCategories {
				found := false
				for _, got := range got {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DetectAllCategories(%q) missing category %v", tt.text, want)
					t.Errorf("  got: %v", got)
					t.Errorf("  want: %v", tt.wantCategories)
				}
			}
		})
	}
}
