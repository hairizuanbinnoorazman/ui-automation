package testprocedure

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	// ErrNameTooLong is returned when the name exceeds the maximum length.
	ErrNameTooLong = errors.New("name exceeds maximum length")

	// ErrDescriptionTooLong is returned when the description exceeds the maximum length.
	ErrDescriptionTooLong = errors.New("description exceeds maximum length")

	// ErrStepsJSONTooLong is returned when the serialized steps exceed the maximum length.
	ErrStepsJSONTooLong = errors.New("steps JSON exceeds maximum length")

	// ErrTooManySteps is returned when the number of steps exceeds the maximum.
	ErrTooManySteps = errors.New("too many steps")

	// ErrInvalidStepStructure is returned when step structure is invalid.
	ErrInvalidStepStructure = errors.New("invalid step structure")

	// ErrSuspiciousContent is returned when content contains suspicious patterns.
	ErrSuspiciousContent = errors.New("content contains suspicious patterns")
)

// ValidationLimits defines the limits for test procedure validation.
type ValidationLimits struct {
	MaxNameLength        int
	MaxDescriptionLength int
	MaxStepsJSONLength   int
	MaxStepsCount        int
}

// DefaultValidationLimits returns the default validation limits.
func DefaultValidationLimits() ValidationLimits {
	return ValidationLimits{
		MaxNameLength:        255,
		MaxDescriptionLength: 5000,
		MaxStepsJSONLength:   50000,
		MaxStepsCount:        200,
	}
}

// ValidateForScriptGeneration performs comprehensive validation of a test procedure
// before it's used for script generation. This includes stricter checks than the
// regular Validate() method to prevent prompt injection attacks.
func ValidateForScriptGeneration(tp *TestProcedure, limits ValidationLimits) error {
	// First, run basic validation
	if err := tp.Validate(); err != nil {
		return err
	}

	// Check length limits
	if len(tp.Name) > limits.MaxNameLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrNameTooLong, len(tp.Name), limits.MaxNameLength)
	}

	if len(tp.Description) > limits.MaxDescriptionLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrDescriptionTooLong, len(tp.Description), limits.MaxDescriptionLength)
	}

	// Validate steps structure
	if err := ValidateStepStructure(tp.Steps, limits); err != nil {
		return err
	}

	// Check for suspicious content patterns
	if err := checkSuspiciousPatterns(tp); err != nil {
		return err
	}

	return nil
}

// ValidateStepStructure validates the structure of test procedure steps.
// Ensures steps contain valid action types and required fields.
func ValidateStepStructure(steps Steps, limits ValidationLimits) error {
	if steps == nil {
		return nil
	}

	// Check steps count
	if len(steps) > limits.MaxStepsCount {
		return fmt.Errorf("%w: %d steps (max %d)", ErrTooManySteps, len(steps), limits.MaxStepsCount)
	}

	// Check serialized length
	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}
	if len(stepsJSON) > limits.MaxStepsJSONLength {
		return fmt.Errorf("%w: %d characters (max %d)", ErrStepsJSONTooLong, len(stepsJSON), limits.MaxStepsJSONLength)
	}

	// Validate known action types
	validActions := map[string]bool{
		"navigate":    true,
		"click":       true,
		"type":        true,
		"wait":        true,
		"assert_text": true,
		"screenshot":  true,
	}

	for i, step := range steps {
		// Check that action field exists and is a string
		action, ok := step["action"].(string)
		if !ok {
			return fmt.Errorf("%w: step %d missing or invalid 'action' field", ErrInvalidStepStructure, i)
		}

		// Validate action type
		if !validActions[action] {
			return fmt.Errorf("%w: step %d has unknown action type '%s'", ErrInvalidStepStructure, i, action)
		}

		// Validate required fields for each action type
		if err := validateStepRequiredFields(action, step, i); err != nil {
			return err
		}

		// Validate field types
		if err := validateStepFieldTypes(step, i); err != nil {
			return err
		}
	}

	return nil
}

// validateStepRequiredFields checks that required fields exist for each action type.
func validateStepRequiredFields(action string, step map[string]interface{}, index int) error {
	switch action {
	case "navigate":
		if _, ok := step["url"]; !ok {
			return fmt.Errorf("%w: step %d (navigate) missing required 'url' field", ErrInvalidStepStructure, index)
		}
	case "click":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("%w: step %d (click) missing required 'selector' field", ErrInvalidStepStructure, index)
		}
	case "type":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("%w: step %d (type) missing required 'selector' field", ErrInvalidStepStructure, index)
		}
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("%w: step %d (type) missing required 'value' field", ErrInvalidStepStructure, index)
		}
	case "assert_text":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("%w: step %d (assert_text) missing required 'selector' field", ErrInvalidStepStructure, index)
		}
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("%w: step %d (assert_text) missing required 'value' field", ErrInvalidStepStructure, index)
		}
	case "screenshot":
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("%w: step %d (screenshot) missing required 'value' field", ErrInvalidStepStructure, index)
		}
	}
	return nil
}

// validateStepFieldTypes validates that step fields have expected types.
func validateStepFieldTypes(step map[string]interface{}, index int) error {
	// Known string fields
	stringFields := map[string]bool{
		"action":   true,
		"url":      true,
		"selector": true,
		"value":    true,
	}

	for key, value := range step {
		if stringFields[key] {
			if _, ok := value.(string); !ok {
				return fmt.Errorf("%w: step %d field '%s' must be a string", ErrInvalidStepStructure, index, key)
			}
		}

		// Special case: timeout can be number or string
		if key == "timeout" {
			switch value.(type) {
			case float64, int, int64, string:
				// Valid types
			default:
				return fmt.Errorf("%w: step %d field 'timeout' must be a number or string", ErrInvalidStepStructure, index)
			}
		}
	}

	return nil
}

// checkSuspiciousPatterns checks for patterns commonly associated with prompt injection.
// This is a heuristic check and may produce false positives, but it's an additional
// layer of defense.
func checkSuspiciousPatterns(tp *TestProcedure) error {
	// Suspicious phrases that might indicate injection attempts
	suspiciousPatterns := []string{
		"ignore previous instructions",
		"ignore all previous",
		"disregard previous",
		"forget all previous",
		"new instructions:",
		"system:",
		"</test_procedure>",
		"</requirements>",
		"<test_procedure>",
		"<requirements>",
		"</test_steps>",
		"<test_steps>",
		"</name>",
		"</description>",
	}

	// Check name
	if err := checkStringForSuspiciousPatterns(tp.Name, "name", suspiciousPatterns); err != nil {
		return err
	}

	// Check description
	if err := checkStringForSuspiciousPatterns(tp.Description, "description", suspiciousPatterns); err != nil {
		return err
	}

	// Check for excessive control characters (potential encoding attacks)
	if hasExcessiveControlCharacters(tp.Name) || hasExcessiveControlCharacters(tp.Description) {
		return fmt.Errorf("%w: content contains excessive control characters", ErrSuspiciousContent)
	}

	// Check all string fields within steps
	if tp.Steps != nil {
		for i, step := range tp.Steps {
			// Check all string values in the step
			for key, value := range step {
				if strValue, ok := value.(string); ok {
					fieldName := fmt.Sprintf("step[%d].%s", i, key)
					if err := checkStringForSuspiciousPatterns(strValue, fieldName, suspiciousPatterns); err != nil {
						return err
					}

					// Check for excessive control characters in step string fields
					if hasExcessiveControlCharacters(strValue) {
						return fmt.Errorf("%w: %s contains excessive control characters", ErrSuspiciousContent, fieldName)
					}
				}
			}
		}
	}

	return nil
}

// checkStringForSuspiciousPatterns checks a string value against a list of suspicious patterns.
func checkStringForSuspiciousPatterns(value, fieldName string, patterns []string) error {
	valueLower := strings.ToLower(value)
	for _, pattern := range patterns {
		if strings.Contains(valueLower, pattern) {
			return fmt.Errorf("%w: %s contains suspicious pattern '%s'", ErrSuspiciousContent, fieldName, pattern)
		}
	}
	return nil
}

// hasExcessiveControlCharacters checks if a string has an unusual number of control characters.
func hasExcessiveControlCharacters(s string) bool {
	if len(s) == 0 {
		return false
	}

	controlCount := 0
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			controlCount++
		}
	}

	// If more than 5% of characters are control characters (excluding common formatting),
	// consider it suspicious
	threshold := len(s) / 20
	if threshold < 5 {
		threshold = 5
	}

	return controlCount > threshold
}
