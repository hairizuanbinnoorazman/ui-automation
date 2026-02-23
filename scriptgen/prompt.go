package scriptgen

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// BuildPrompt constructs a prompt for the LLM to generate an automation script.
// It validates and sanitizes all user-provided content before embedding it in the prompt
// to prevent prompt injection attacks.
func BuildPrompt(procedure *testprocedure.TestProcedure, framework Framework, config *ValidationConfig) (string, error) {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Validate before sanitizing: enforce length limits, step structure, and injection patterns.
	limits := testprocedure.ValidationLimits{
		MaxNameLength:        config.MaxNameLength,
		MaxDescriptionLength: config.MaxDescriptionLength,
		MaxStepsJSONLength:   config.MaxStepsJSONLength,
		MaxStepsCount:        config.MaxStepsCount,
	}
	if err := testprocedure.ValidateForScriptGeneration(procedure, limits); err != nil {
		if errors.Is(err, testprocedure.ErrNameTooLong) || errors.Is(err, testprocedure.ErrDescriptionTooLong) {
			return "", err
		}
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	// Sanitize all user-provided content
	sanitizedName := SanitizeTestProcedureName(procedure.Name)
	sanitizedDescription := SanitizeTestProcedureDescription(procedure.Description)

	// Sanitize and validate steps
	sanitizedSteps, err := SanitizeSteps(procedure.Steps)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize steps: %w", err)
	}

	// Marshal sanitized steps to JSON for readability
	stepsJSON, err := json.MarshalIndent(sanitizedSteps, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal steps: %w", err)
	}

	frameworkName := "Selenium"
	if framework == FrameworkPlaywright {
		frameworkName = "Playwright"
	}

	// Use XML-style tags to create clear boundaries between instructions and user data
	// This follows Anthropic's prompt engineering best practices and makes it harder
	// to "break out" of the user data section.
	prompt := fmt.Sprintf(`Generate a Python automation script using %s for the following test procedure.

<test_procedure>
<name>%s</name>
<version>%d</version>
<description>%s</description>
<test_steps>
%s
</test_steps>
</test_procedure>

<requirements>
- Use Python 3.x syntax
- Include proper error handling and try-except blocks
- Add docstrings for the main test class and methods
- Make the script executable and runnable
- Return ONLY the Python code without markdown formatting or code blocks
- Do not include any explanatory text before or after the code

Action types and their meanings:
- navigate: Open URL in browser (requires "url" field)
- click: Click on element using CSS selector (requires "selector" field)
- type: Enter text into input field (requires "selector" and "value" fields)
- wait: Pause execution (optional "timeout" field in seconds, default 2)
- assert_text: Verify text content of element (requires "selector" and "value" fields)
- screenshot: Capture screenshot (requires "value" field as filename)

%s

The script should:
1. Set up the browser driver
2. Execute each test step in order
3. Handle errors gracefully with meaningful error messages
4. Clean up resources (close browser) in a finally block
5. Print progress messages as it executes each step
6. Exit with appropriate status code (0 for success, non-zero for failure)
</requirements>`,
		frameworkName,
		sanitizedName,
		procedure.Version,
		sanitizedDescription,
		string(stepsJSON),
		getFrameworkSpecificInstructions(framework),
	)

	return prompt, nil
}

func getFrameworkSpecificInstructions(framework Framework) string {
	if framework == FrameworkSelenium {
		return `For Selenium:
- Use selenium.webdriver for browser automation
- Use WebDriverWait for explicit waits
- Use expected_conditions for element interactions
- Create a ChromeDriver instance (or accept browser type as parameter)
- Include proper imports: from selenium import webdriver, from selenium.webdriver.common.by import By, etc.`
	}

	return `For Playwright:
- Use playwright.sync_api for synchronous browser automation
- Use page.wait_for_selector for element waits
- Create a chromium browser instance (or accept browser type as parameter)
- Include proper imports: from playwright.sync_api import sync_playwright
- Use context manager pattern for browser lifecycle`
}
