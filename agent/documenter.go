package agent

import (
	"context"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// Documenter is the third agent in the pipeline. It creates test procedures from exploration results.
type Documenter struct {
	config Config
	logger logger.Logger
}

// NewDocumenter creates a new documenter agent.
func NewDocumenter(config Config, log logger.Logger) *Documenter {
	return &Documenter{
		config: config,
		logger: log,
	}
}

// Document creates a test procedure from exploration results.
func (d *Documenter) Document(ctx context.Context, procedureName string, result *ExplorationResult) (*testprocedure.TestProcedure, error) {
	d.logger.Info(ctx, "creating test procedure from exploration", map[string]interface{}{
		"procedure_name": procedureName,
		"interactions":   len(result.Interactions),
	})

	// TODO: Call Claude via Bedrock to generate structured test procedure
	// For now, convert interactions directly to steps

	steps := make(testprocedure.Steps, 0, len(result.Interactions))
	for _, interaction := range result.Interactions {
		var imagePaths []string
		if interaction.ScreenshotPath != "" {
			imagePaths = append(imagePaths, interaction.ScreenshotPath)
		}
		steps = append(steps, testprocedure.TestStep{
			Name:         interaction.ActionType,
			Instructions: interaction.Description,
			ImagePaths:   imagePaths,
		})
	}

	// If no steps from interactions, create a placeholder
	if len(steps) == 0 {
		steps = append(steps, testprocedure.TestStep{
			Name:         "Initial observation",
			Instructions: result.Summary,
			ImagePaths:   []string{},
		})
	}

	tp := &testprocedure.TestProcedure{
		Name:        procedureName,
		Description: "Auto-generated from UI exploration: " + result.Summary,
		Steps:       steps,
	}

	d.logger.Info(ctx, "test procedure created from exploration", map[string]interface{}{
		"procedure_name": procedureName,
		"steps":          len(steps),
	})

	return tp, nil
}
