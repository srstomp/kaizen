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
