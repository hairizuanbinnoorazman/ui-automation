package scriptgen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// BedrockGenerator implements ScriptGenerator using AWS Bedrock.
type BedrockGenerator struct {
	client         *bedrockruntime.Client
	modelID        string
	maxTokens      int
	validationCfg  *ValidationConfig
}

// NewBedrockGenerator creates a new Bedrock-based script generator.
func NewBedrockGenerator(region, modelID string, maxTokens int) (*BedrockGenerator, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockGenerator{
		client:        client,
		modelID:       modelID,
		maxTokens:     maxTokens,
		validationCfg: DefaultValidationConfig(),
	}, nil
}

// SetValidationConfig sets the validation configuration for the generator.
func (g *BedrockGenerator) SetValidationConfig(cfg *ValidationConfig) {
	g.validationCfg = cfg
}

// Generate creates a Python automation script using AWS Bedrock.
func (g *BedrockGenerator) Generate(ctx context.Context, procedure *testprocedure.TestProcedure, framework Framework) ([]byte, error) {
	// Build the prompt with validation and sanitization
	prompt, err := BuildPrompt(procedure, framework, g.validationCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// TODO: Add security logging here if logger is available
	// Log if description exceeds warning threshold (2000 chars)
	// Log if suspicious patterns detected but not blocked

	// Prepare the request payload for Claude models
	// Format depends on the model being used
	requestBody := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        g.maxTokens,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock API
	output, err := g.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(g.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	// Parse the response
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract the generated code
	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	generatedCode := strings.TrimSpace(response.Content[0].Text)
	if generatedCode == "" {
		return nil, fmt.Errorf("empty generated code")
	}

	// Strip markdown code fences — LLMs often include these despite prompt instructions.
	if strings.HasPrefix(generatedCode, "```") {
		// Remove opening fence line (e.g. "```python\n" or "```\n")
		if idx := strings.Index(generatedCode, "\n"); idx != -1 {
			generatedCode = generatedCode[idx+1:]
		}
		// Remove closing fence
		generatedCode = strings.TrimSuffix(strings.TrimSpace(generatedCode), "```")
		generatedCode = strings.TrimSpace(generatedCode)
	}

	return []byte(generatedCode), nil
}
