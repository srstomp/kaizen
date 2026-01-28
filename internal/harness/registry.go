package harness

import (
	"github.com/srstomp/kaizen/internal/graders/codebased"
	"github.com/srstomp/kaizen/internal/graders/modelbased"
)

// GraderRegistry maintains a registry of available graders
type GraderRegistry struct {
	codeGraders  map[string]codebased.CodeGrader
	modelGraders map[string]modelbased.Grader
}

// NewGraderRegistry creates a new grader registry and initializes all available graders
func NewGraderRegistry() *GraderRegistry {
	registry := &GraderRegistry{
		codeGraders:  make(map[string]codebased.CodeGrader),
		modelGraders: make(map[string]modelbased.Grader),
	}

	// Register code-based graders
	registry.registerCodeGrader(codebased.NewFileExistsGrader())
	registry.registerCodeGrader(codebased.NewTestExistsGrader())
	registry.registerCodeGrader(codebased.NewEndpointExistsGrader())
	registry.registerCodeGrader(codebased.NewTestCoverageGrader())

	// Register model-based graders
	registry.registerModelGrader(modelbased.NewSpecComplianceGrader())
	registry.registerModelGrader(modelbased.NewTaskQualityGrader())
	registry.registerModelGrader(modelbased.NewSkillClarityGrader())

	return registry
}

// registerCodeGrader adds a code-based grader to the registry
func (r *GraderRegistry) registerCodeGrader(grader codebased.CodeGrader) {
	r.codeGraders[grader.Name()] = grader
}

// registerModelGrader adds a model-based grader to the registry
func (r *GraderRegistry) registerModelGrader(grader modelbased.Grader) {
	// Model graders don't have a Name() method, so we need to derive it from the type
	// We'll use the grader type as the key
	switch grader.(type) {
	case *modelbased.SpecComplianceGrader:
		r.modelGraders["spec_compliance"] = grader
	case *modelbased.TaskQualityGrader:
		r.modelGraders["task_quality"] = grader
	case *modelbased.SkillClarityGrader:
		r.modelGraders["skill_clarity"] = grader
	}
}

// GetCodeGrader returns a code-based grader by name, or nil if not found
func (r *GraderRegistry) GetCodeGrader(name string) codebased.CodeGrader {
	return r.codeGraders[name]
}

// GetModelGrader returns a model-based grader by name, or nil if not found
func (r *GraderRegistry) GetModelGrader(name string) modelbased.Grader {
	return r.modelGraders[name]
}
