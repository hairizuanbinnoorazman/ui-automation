package scriptgen

import (
	"context"

	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// ScriptGenerator defines the interface for generating automation scripts.
// Implementations can use different backends (AWS Bedrock, OpenAI, local templates, etc.)
type ScriptGenerator interface {
	// Generate creates a Python automation script from a test procedure
	Generate(ctx context.Context, procedure *testprocedure.TestProcedure, framework Framework) ([]byte, error)
}
