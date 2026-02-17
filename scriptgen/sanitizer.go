package scriptgen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

var (
	// allowedNameChars matches alphanumeric, spaces, hyphens, underscores, and parentheses
	allowedNameChars = regexp.MustCompile(`^[a-zA-Z0-9 \-_()]+$`)

	// validActionTypes defines the allowed step action types
	validActionTypes = map[string]bool{
		"navigate":    true,
		"click":       true,
		"type":        true,
		"wait":        true,
		"assert_text": true,
		"screenshot":  true,
	}
)

// SanitizeTestProcedureName sanitizes the test procedure name for use in prompts.
// It removes or replaces potentially problematic characters while preserving
// legitimate use cases.
func SanitizeTestProcedureName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove control characters (except newlines which should not be in names)
	name = removeControlCharacters(name, false)

	// Keep only allowed characters
	if !allowedNameChars.MatchString(name) {
		// Replace disallowed characters with underscore
		var result strings.Builder
		for _, r := range name {
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' || r == '_' || r == '(' || r == ')' {
				result.WriteRune(r)
			} else {
				result.WriteRune('_')
			}
		}
		name = result.String()
	}

	// Normalize multiple spaces to single space
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	// Final trim
	return strings.TrimSpace(name)
}

// SanitizeTestProcedureDescription sanitizes the test procedure description.
// Removes control characters and normalizes whitespace while preserving
// legitimate formatting.
func SanitizeTestProcedureDescription(desc string) string {
	// Trim whitespace
	desc = strings.TrimSpace(desc)

	// Remove control characters but preserve newlines and tabs
	desc = removeControlCharacters(desc, true)

	// Remove non-printable characters
	desc = removeNonPrintable(desc)

	// Normalize excessive whitespace (but keep paragraph breaks)
	// Replace 3+ newlines with 2 newlines
	desc = regexp.MustCompile(`\n{3,}`).ReplaceAllString(desc, "\n\n")

	// Normalize spaces and tabs within lines
	lines := strings.Split(desc, "\n")
	for i, line := range lines {
		lines[i] = regexp.MustCompile(`[ \t]+`).ReplaceAllString(line, " ")
		lines[i] = strings.TrimSpace(lines[i])
	}
	desc = strings.Join(lines, "\n")

	// Final trim
	return strings.TrimSpace(desc)
}

// SanitizeSteps validates and sanitizes test procedure steps.
// Returns sanitized steps or error if validation fails.
func SanitizeSteps(steps testprocedure.Steps) (testprocedure.Steps, error) {
	if steps == nil || len(steps) == 0 {
		return steps, nil
	}

	sanitized := make(testprocedure.Steps, 0, len(steps))

	for i, step := range steps {
		// Validate action type
		action, ok := step["action"].(string)
		if !ok {
			return nil, fmt.Errorf("step %d: missing or invalid action field", i)
		}

		if !validActionTypes[action] {
			return nil, fmt.Errorf("step %d: invalid action type '%s'", i, action)
		}

		sanitizedStep := make(map[string]interface{})
		sanitizedStep["action"] = action

		// Sanitize each field based on type
		for key, value := range step {
			if key == "action" {
				continue // Already handled
			}

			switch v := value.(type) {
			case string:
				// Sanitize string fields
				sanitizedStep[key] = sanitizeStepStringField(key, v)
			case float64, int, int64, bool:
				// Numeric and boolean values are safe
				sanitizedStep[key] = v
			default:
				// Skip unknown types to prevent injection
				continue
			}
		}

		// Validate required fields for specific actions
		if err := validateStepFields(action, sanitizedStep); err != nil {
			return nil, fmt.Errorf("step %d: %w", i, err)
		}

		sanitized = append(sanitized, sanitizedStep)
	}

	return sanitized, nil
}

// sanitizeStepStringField sanitizes string fields in test steps.
func sanitizeStepStringField(key, value string) string {
	value = strings.TrimSpace(value)

	// For URLs, basic validation
	if key == "url" {
		// Remove control characters
		value = removeControlCharacters(value, false)
		// Basic URL validation - must start with http:// or https://
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			// Prepend https:// if missing
			value = "https://" + value
		}
		return value
	}

	// For selectors, remove control characters
	if key == "selector" {
		value = removeControlCharacters(value, false)
		return strings.TrimSpace(value)
	}

	// For value fields, preserve some formatting but remove control chars
	if key == "value" {
		value = removeControlCharacters(value, true) // Keep newlines for multi-line text
		value = removeNonPrintable(value)
		return value
	}

	// Default: remove control characters
	return removeControlCharacters(value, false)
}

// validateStepFields validates that required fields exist for specific action types.
func validateStepFields(action string, step map[string]interface{}) error {
	switch action {
	case "navigate":
		if _, ok := step["url"]; !ok {
			return fmt.Errorf("navigate action requires 'url' field")
		}
	case "click":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("click action requires 'selector' field")
		}
	case "type":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("type action requires 'selector' field")
		}
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("type action requires 'value' field")
		}
	case "assert_text":
		if _, ok := step["selector"]; !ok {
			return fmt.Errorf("assert_text action requires 'selector' field")
		}
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("assert_text action requires 'value' field")
		}
	case "screenshot":
		if _, ok := step["value"]; !ok {
			return fmt.Errorf("screenshot action requires 'value' field (filename)")
		}
	}
	return nil
}

// removeControlCharacters removes control characters from a string.
// If preserveFormatting is true, newlines (\n), tabs (\t), and carriage returns (\r) are preserved.
func removeControlCharacters(s string, preserveFormatting bool) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			if preserveFormatting && (r == '\n' || r == '\t' || r == '\r') {
				result.WriteRune(r)
			}
			// Skip other control characters
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// removeNonPrintable removes non-printable characters while preserving
// common formatting characters.
func removeNonPrintable(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' || r == '\r' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateLengthLimits checks if the test procedure content exceeds
// configured length limits. Returns error if any limit is exceeded.
func ValidateLengthLimits(tp *testprocedure.TestProcedure, config *ValidationConfig) error {
	if len(tp.Name) > config.MaxNameLength {
		return fmt.Errorf("name exceeds maximum length of %d characters", config.MaxNameLength)
	}

	if len(tp.Description) > config.MaxDescriptionLength {
		return fmt.Errorf("description exceeds maximum length of %d characters", config.MaxDescriptionLength)
	}

	// Marshal steps to check serialized length
	if tp.Steps != nil {
		stepsJSON, err := json.Marshal(tp.Steps)
		if err != nil {
			return fmt.Errorf("failed to marshal steps: %w", err)
		}
		if len(stepsJSON) > config.MaxStepsJSONLength {
			return fmt.Errorf("steps JSON exceeds maximum length of %d characters", config.MaxStepsJSONLength)
		}
	}

	// Check steps count
	if len(tp.Steps) > config.MaxStepsCount {
		return fmt.Errorf("procedure has %d steps, maximum allowed is %d", len(tp.Steps), config.MaxStepsCount)
	}

	return nil
}

// ValidationConfig holds the configuration for validation limits.
type ValidationConfig struct {
	MaxNameLength        int
	MaxDescriptionLength int
	MaxStepsJSONLength   int
	MaxStepsCount        int
}

// DefaultValidationConfig returns the default validation configuration.
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxNameLength:        255,
		MaxDescriptionLength: 5000,
		MaxStepsJSONLength:   50000,
		MaxStepsCount:        200,
	}
}
