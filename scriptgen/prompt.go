package scriptgen

import (
	"encoding/json"
	"fmt"

	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// BuildPrompt constructs a prompt for the LLM to generate an automation script.
func BuildPrompt(procedure *testprocedure.TestProcedure, framework Framework) (string, error) {
	// Marshal steps to JSON for readability
	stepsJSON, err := json.MarshalIndent(procedure.Steps, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal steps: %w", err)
	}

	frameworkName := "Selenium"
	if framework == FrameworkPlaywright {
		frameworkName = "Playwright"
	}

	prompt := fmt.Sprintf(`Generate a Python automation script using %s for the following test procedure:

Name: %s
Version: %d
Description: %s

Test Steps:
%s

Requirements:
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
`,
		frameworkName,
		procedure.Name,
		procedure.Version,
		procedure.Description,
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
