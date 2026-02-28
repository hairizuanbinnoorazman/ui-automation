#!/usr/bin/env python3
"""
UI Exploration Agent Runner

Uses claude-agent-sdk to orchestrate a UI exploration pipeline:
1. Planner: Creates exploration strategy
2. Explorer: Navigates UI and captures screenshots
3. Documenter: Creates structured test procedure

Input:  JSON config via stdin
Output: JSON result at {output_dir}/result.json
"""

import json
import os
import sys

import anyio
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, TextBlock


COORDINATOR_SYSTEM_PROMPT = """You are a UI exploration coordinator agent. Your job is to explore a web application and create a structured test procedure document.

You will be given:
- A target URL to explore
- Optional credentials for authentication
- A procedure name
- An output directory for screenshots

You must complete these three phases IN ORDER:

## Phase 1: Planning
Analyze the target URL and plan your exploration strategy. Consider:
- What pages/sections to visit
- What interactions to test (forms, buttons, navigation)
- Whether credentials are needed and when to use them

## Phase 2: Exploration
Navigate the web application using Playwright browser tools:
1. Use `browser_navigate` to go to pages
2. Use `browser_snapshot` to get the accessibility tree and understand the page structure
3. Use `browser_screenshot` to capture visual state
4. Use `browser_click`, `browser_type`, etc. to interact with elements
5. After EACH significant interaction, take a snapshot and screenshot
6. Use the Bash tool to save screenshots: copy them to {output_dir}/screenshots/ with descriptive names like "01_landing_page.png", "02_login_form.png"

For SPAs (Single Page Applications):
- After navigation clicks, use browser_snapshot to verify content has loaded
- Wait for dynamic content by checking snapshots before screenshotting

## Phase 3: Documentation
After exploring, create a structured test procedure. Write the final result as a JSON file to {output_dir}/result.json using the Bash tool.

The JSON format MUST be:
{{
  "procedure_name": "<name>",
  "description": "<brief description of what was explored>",
  "steps": [
    {{
      "name": "<short step name>",
      "instructions": "<detailed instructions for this step>",
      "image_paths": ["screenshots/<filename>.png"]
    }}
  ],
  "summary": "<overall summary of the exploration>"
}}

IMPORTANT:
- You MUST write the result.json file at the end using the Bash tool
- Screenshot paths in result.json should be relative to the output directory (e.g., "screenshots/01_landing.png")
- Each step should have clear, actionable instructions a manual tester can follow
- Group related interactions into logical steps
- Include verification points (what the tester should observe after each action)
"""


async def run_agent(config: dict) -> None:
    target_url = config["target_url"]
    credentials = config.get("credentials", [])
    procedure_name = config.get("procedure_name", "UI Exploration")
    output_dir = config["output_dir"]
    playwright_mcp_url = config.get(
        "playwright_mcp_url", "http://playwright-mcp:3000/sse"
    )

    # Ensure output directories exist
    os.makedirs(os.path.join(output_dir, "screenshots"), exist_ok=True)

    # Build credential instructions
    cred_text = ""
    if credentials:
        cred_lines = [f"  - {c['key']}: {c['value']}" for c in credentials]
        cred_text = "\n\nAvailable credentials:\n" + "\n".join(cred_lines)

    prompt = (
        f'Explore the web application at {target_url} and create a test procedure '
        f'named "{procedure_name}".\n\n'
        f"Output directory: {output_dir}\n"
        f"Screenshots directory: {output_dir}/screenshots/\n"
        f"Result file: {output_dir}/result.json\n"
        f"{cred_text}\n\n"
        f"Begin with Phase 1 (Planning), then Phase 2 (Exploration), "
        f"then Phase 3 (Documentation).\n"
        f"Make sure to write the result.json file when you're done."
    )

    options = ClaudeAgentOptions(
        system_prompt=COORDINATOR_SYSTEM_PROMPT,
        max_turns=100,
        allowed_tools=["Bash", "Task", "mcp__playwright__*"],
        permission_mode="bypassPermissions",
        mcp_servers={
            "playwright": {
                "type": "sse",
                "url": playwright_mcp_url,
            }
        },
    )

    final_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    final_text = block.text

    # Verify result.json was created by the agent
    result_path = os.path.join(output_dir, "result.json")
    if not os.path.exists(result_path):
        # If the agent didn't create the file, write a fallback
        fallback = {
            "procedure_name": procedure_name,
            "description": f"Auto-generated exploration of {target_url}",
            "steps": [
                {
                    "name": "Initial observation",
                    "instructions": (
                        final_text or "Agent did not produce structured output"
                    ),
                    "image_paths": [],
                }
            ],
            "summary": "Exploration completed with fallback output",
        }
        with open(result_path, "w") as f:
            json.dump(fallback, f, indent=2)


def main() -> None:
    # Read config from stdin
    config_data = sys.stdin.read()
    if not config_data.strip():
        print("Error: no config provided on stdin", file=sys.stderr)
        sys.exit(1)

    try:
        config = json.loads(config_data)
    except json.JSONDecodeError as e:
        print(f"Error: invalid JSON config: {e}", file=sys.stderr)
        sys.exit(1)

    # Validate required fields
    required = ["target_url", "output_dir"]
    for field in required:
        if field not in config:
            print(f"Error: missing required field '{field}'", file=sys.stderr)
            sys.exit(1)

    anyio.run(run_agent, config)


if __name__ == "__main__":
    main()
