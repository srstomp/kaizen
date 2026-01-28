package harness

import (
	"testing"
)

func TestNewGraderRegistry(t *testing.T) {
	registry := NewGraderRegistry()

	if registry == nil {
		t.Fatal("NewGraderRegistry() returned nil")
	}
}

func TestGetCodeGrader_ValidGraders(t *testing.T) {
	registry := NewGraderRegistry()

	tests := []struct {
		name       string
		graderName string
		wantNil    bool
	}{
		{
			name:       "file-exists grader exists",
			graderName: "file-exists",
			wantNil:    false,
		},
		{
			name:       "test-exists grader exists",
			graderName: "test-exists",
			wantNil:    false,
		},
		{
			name:       "endpoint-exists grader exists",
			graderName: "endpoint-exists",
			wantNil:    false,
		},
		{
			name:       "test-coverage grader exists",
			graderName: "test-coverage",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grader := registry.GetCodeGrader(tt.graderName)
			if (grader == nil) != tt.wantNil {
				t.Errorf("GetCodeGrader(%q) returned nil=%v, want nil=%v", tt.graderName, grader == nil, tt.wantNil)
			}
			if grader != nil && grader.Name() != tt.graderName {
				t.Errorf("GetCodeGrader(%q) returned grader with Name()=%q, want %q", tt.graderName, grader.Name(), tt.graderName)
			}
		})
	}
}

func TestGetCodeGrader_UnknownGrader(t *testing.T) {
	registry := NewGraderRegistry()

	grader := registry.GetCodeGrader("unknown_grader")
	if grader != nil {
		t.Errorf("GetCodeGrader('unknown_grader') = %v, want nil", grader)
	}
}

func TestGetModelGrader_ValidGraders(t *testing.T) {
	registry := NewGraderRegistry()

	tests := []struct {
		name       string
		graderName string
		wantNil    bool
	}{
		{
			name:       "spec_compliance grader exists",
			graderName: "spec_compliance",
			wantNil:    false,
		},
		{
			name:       "task_quality grader exists",
			graderName: "task_quality",
			wantNil:    false,
		},
		{
			name:       "skill_clarity grader exists",
			graderName: "skill_clarity",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grader := registry.GetModelGrader(tt.graderName)
			if (grader == nil) != tt.wantNil {
				t.Errorf("GetModelGrader(%q) returned nil=%v, want nil=%v", tt.graderName, grader == nil, tt.wantNil)
			}
		})
	}
}

func TestGetModelGrader_UnknownGrader(t *testing.T) {
	registry := NewGraderRegistry()

	grader := registry.GetModelGrader("unknown_grader")
	if grader != nil {
		t.Errorf("GetModelGrader('unknown_grader') = %v, want nil", grader)
	}
}
