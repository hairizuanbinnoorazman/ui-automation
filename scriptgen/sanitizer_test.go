package scriptgen

import (
	"encoding/json"
	"testing"

	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeTestProcedureName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid name unchanged",
			input:    "Test Login Flow",
			expected: "Test Login Flow",
		},
		{
			name:     "name with hyphens and underscores",
			input:    "Test-Login_Flow",
			expected: "Test-Login_Flow",
		},
		{
			name:     "name with parentheses",
			input:    "Test (Login) Flow",
			expected: "Test (Login) Flow",
		},
		{
			name:     "name with special characters replaced",
			input:    "Test@Login#Flow",
			expected: "Test_Login_Flow",
		},
		{
			name:     "name with control characters removed",
			input:    "Test\x00Login\x01Flow",
			expected: "TestLoginFlow",
		},
		{
			name:     "name with multiple spaces normalized",
			input:    "Test    Login     Flow",
			expected: "Test Login Flow",
		},
		{
			name:     "name with leading/trailing whitespace",
			input:    "  Test Login Flow  ",
			expected: "Test Login Flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTestProcedureName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeTestProcedureDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid description unchanged",
			input:    "This is a test description.",
			expected: "This is a test description.",
		},
		{
			name:     "description with newlines preserved",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "description with excessive newlines normalized",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "description with tabs normalized",
			input:    "Test\t\t\tDescription",
			expected: "Test Description",
		},
		{
			name:     "description with control characters removed",
			input:    "Test\x00Description\x01Here",
			expected: "TestDescriptionHere",
		},
		{
			name:     "description with leading/trailing whitespace",
			input:    "  Test Description  ",
			expected: "Test Description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTestProcedureDescription(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeSteps(t *testing.T) {
	tests := []struct {
		name        string
		steps       testprocedure.Steps
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid navigate step",
			steps: testprocedure.Steps{
				{
					"action": "navigate",
					"url":    "https://example.com",
				},
			},
			expectError: false,
		},
		{
			name: "valid click step",
			steps: testprocedure.Steps{
				{
					"action":   "click",
					"selector": "#login-button",
				},
			},
			expectError: false,
		},
		{
			name: "valid type step",
			steps: testprocedure.Steps{
				{
					"action":   "type",
					"selector": "#username",
					"value":    "testuser",
				},
			},
			expectError: false,
		},
		{
			name: "step with invalid action type",
			steps: testprocedure.Steps{
				{
					"action": "invalid_action",
				},
			},
			expectError: true,
			errorMsg:    "invalid action type",
		},
		{
			name: "navigate step missing url",
			steps: testprocedure.Steps{
				{
					"action": "navigate",
				},
			},
			expectError: true,
			errorMsg:    "requires 'url' field",
		},
		{
			name: "click step missing selector",
			steps: testprocedure.Steps{
				{
					"action": "click",
				},
			},
			expectError: true,
			errorMsg:    "requires 'selector' field",
		},
		{
			name: "type step missing value",
			steps: testprocedure.Steps{
				{
					"action":   "type",
					"selector": "#username",
				},
			},
			expectError: true,
			errorMsg:    "requires 'value' field",
		},
		{
			name: "step with control characters in selector",
			steps: testprocedure.Steps{
				{
					"action":   "click",
					"selector": "#button\x00\x01",
				},
			},
			expectError: false, // Should sanitize, not error
		},
		{
			name: "url without protocol gets https prefix",
			steps: testprocedure.Steps{
				{
					"action": "navigate",
					"url":    "example.com",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeSteps(tt.steps)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)

				// If URL without protocol, verify it was prefixed
				if tt.name == "url without protocol gets https prefix" {
					url, _ := result[0]["url"].(string)
					assert.Equal(t, "https://example.com", url)
				}
			}
		})
	}
}

func TestValidateLengthLimits(t *testing.T) {
	config := DefaultValidationConfig()

	tests := []struct {
		name        string
		procedure   *testprocedure.TestProcedure
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid procedure within limits",
			procedure: &testprocedure.TestProcedure{
				Name:        "Test Procedure",
				Description: "A short description",
				Steps: testprocedure.Steps{
					{"action": "navigate", "url": "https://example.com"},
				},
			},
			expectError: false,
		},
		{
			name: "name exceeds max length",
			procedure: &testprocedure.TestProcedure{
				Name:        string(make([]byte, 300)), // 300 chars > 255
				Description: "Description",
				Steps:       testprocedure.Steps{},
			},
			expectError: true,
			errorMsg:    "name exceeds maximum length",
		},
		{
			name: "description exceeds max length",
			procedure: &testprocedure.TestProcedure{
				Name:        "Test",
				Description: string(make([]byte, 6000)), // 6000 chars > 5000
				Steps:       testprocedure.Steps{},
			},
			expectError: true,
			errorMsg:    "description exceeds maximum length",
		},
		{
			name: "too many steps",
			procedure: &testprocedure.TestProcedure{
				Name:        "Test",
				Description: "Description",
				Steps:       makeSteps(201), // 201 steps > 200
			},
			expectError: true,
			errorMsg:    "maximum allowed is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLengthLimits(tt.procedure, config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRemoveControlCharacters(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		preserveFormatting  bool
		expected            string
	}{
		{
			name:               "no control chars",
			input:              "Hello World",
			preserveFormatting: false,
			expected:           "Hello World",
		},
		{
			name:               "with null byte",
			input:              "Hello\x00World",
			preserveFormatting: false,
			expected:           "HelloWorld",
		},
		{
			name:               "with newline preserved",
			input:              "Hello\nWorld",
			preserveFormatting: true,
			expected:           "Hello\nWorld",
		},
		{
			name:               "with newline removed",
			input:              "Hello\nWorld",
			preserveFormatting: false,
			expected:           "HelloWorld",
		},
		{
			name:               "with tab preserved",
			input:              "Hello\tWorld",
			preserveFormatting: true,
			expected:           "Hello\tWorld",
		},
		{
			name:               "with mixed control chars",
			input:              "Hello\x00\x01\n\tWorld",
			preserveFormatting: true,
			expected:           "Hello\n\tWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeControlCharacters(tt.input, tt.preserveFormatting)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveNonPrintable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "printable chars unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "newline preserved",
			input:    "Hello\nWorld",
			expected: "Hello\nWorld",
		},
		{
			name:     "tab preserved",
			input:    "Hello\tWorld",
			expected: "Hello\tWorld",
		},
		{
			name:     "non-printable removed",
			input:    "Hello\x00\x01\x02World",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNonPrintable(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a slice of steps for testing
func makeSteps(count int) testprocedure.Steps {
	steps := make(testprocedure.Steps, count)
	for i := 0; i < count; i++ {
		steps[i] = map[string]interface{}{
			"action": "wait",
		}
	}
	return steps
}

func TestSanitizeStepStringField_URLPrefix(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "URL with https preserved",
			key:      "url",
			value:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with http preserved",
			key:      "url",
			value:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "URL without protocol gets https",
			key:      "url",
			value:    "example.com",
			expected: "https://example.com",
		},
		{
			name:     "non-URL field unchanged",
			key:      "selector",
			value:    "#test",
			expected: "#test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeStepStringField(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateStepFields(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		step        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "navigate with URL valid",
			action:      "navigate",
			step:        map[string]interface{}{"action": "navigate", "url": "https://example.com"},
			expectError: false,
		},
		{
			name:        "navigate without URL invalid",
			action:      "navigate",
			step:        map[string]interface{}{"action": "navigate"},
			expectError: true,
			errorMsg:    "requires 'url' field",
		},
		{
			name:        "click with selector valid",
			action:      "click",
			step:        map[string]interface{}{"action": "click", "selector": "#btn"},
			expectError: false,
		},
		{
			name:        "type with selector and value valid",
			action:      "type",
			step:        map[string]interface{}{"action": "type", "selector": "#input", "value": "text"},
			expectError: false,
		},
		{
			name:        "type without value invalid",
			action:      "type",
			step:        map[string]interface{}{"action": "type", "selector": "#input"},
			expectError: true,
			errorMsg:    "requires 'value' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStepFields(tt.action, tt.step)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSanitizeSteps_ComplexScenario(t *testing.T) {
	// Test a realistic multi-step scenario
	input := testprocedure.Steps{
		{
			"action": "navigate",
			"url":    "example.com", // Missing protocol
		},
		{
			"action":   "type",
			"selector": "#username\x00", // Control character
			"value":    "testuser",
		},
		{
			"action":   "click",
			"selector": "#login-btn",
		},
		{
			"action":   "assert_text",
			"selector": ".welcome",
			"value":    "Welcome!",
		},
		{
			"action": "screenshot",
			"value":  "success.png",
		},
	}

	result, err := SanitizeSteps(input)
	require.NoError(t, err)
	assert.Len(t, result, 5)

	// Verify URL was prefixed
	url, ok := result[0]["url"].(string)
	require.True(t, ok)
	assert.Equal(t, "https://example.com", url)

	// Verify control character was removed from selector
	selector, ok := result[1]["selector"].(string)
	require.True(t, ok)
	assert.NotContains(t, selector, "\x00")

	// Verify all actions are preserved
	actions := []string{}
	for _, step := range result {
		action, ok := step["action"].(string)
		require.True(t, ok)
		actions = append(actions, action)
	}
	assert.Equal(t, []string{"navigate", "type", "click", "assert_text", "screenshot"}, actions)
}

func TestStepsJSONLength(t *testing.T) {
	// Create a procedure with large steps to test JSON serialization length check
	largeSteps := make(testprocedure.Steps, 100)
	for i := 0; i < 100; i++ {
		largeSteps[i] = map[string]interface{}{
			"action":   "type",
			"selector": "#input-field-with-a-very-long-selector-name-to-increase-json-size",
			"value":    "This is a long value that will be repeated many times to make the JSON large",
		}
	}

	stepsJSON, err := json.Marshal(largeSteps)
	require.NoError(t, err)

	// Test that we can detect when steps exceed limit
	config := &ValidationConfig{
		MaxNameLength:        255,
		MaxDescriptionLength: 5000,
		MaxStepsJSONLength:   1000, // Set low limit for test
		MaxStepsCount:        200,
	}

	procedure := &testprocedure.TestProcedure{
		Name:        "Test",
		Description: "Description",
		Steps:       largeSteps,
	}

	err = ValidateLengthLimits(procedure, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "steps JSON exceeds maximum length")

	// Verify actual length
	assert.Greater(t, len(stepsJSON), config.MaxStepsJSONLength)
}
