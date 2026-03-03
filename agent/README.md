# UI Exploration Agent

A standalone Python agent that explores a web application and produces a structured test procedure document. It uses the Claude Agent SDK to orchestrate a three-phase pipeline:

1. **Plan** — Analyze the target URL and decide which pages/sections to visit
2. **Explore** — Navigate the UI via a Playwright MCP server, capturing screenshots along the way
3. **Document** — Write a structured `result.json` with step-by-step instructions and screenshot references

When integrated with the full UI Automation platform the Go backend spawns this script as a subprocess. The instructions below explain how to run it **independently** — no Go backend, MySQL, or frontend required.

## Prerequisites

- Python 3.11+
- A running Playwright MCP server (provides browser automation tools)
- **One of** the following for Claude model access:
  - AWS Bedrock credentials (`CLAUDE_CODE_USE_BEDROCK=1` + AWS credentials)
  - Anthropic API key (`ANTHROPIC_API_KEY`)

## Setting up the Playwright MCP server

The agent connects to a Playwright MCP server over SSE. Pick whichever option suits your setup.

### Option A — Docker (recommended)

Build and run the image from the `playwright-mcp/` directory in this repo:

```bash
docker build -t playwright-mcp ./playwright-mcp
docker run -p 3000:3000 playwright-mcp
```

### Option B — Local npm

```bash
npx @playwright/mcp --port 3000 --headless --browser chrome
```

Either way, the SSE endpoint will be available at `http://localhost:3000/sse`.

## Installing Python dependencies

```bash
pip install claude-agent-sdk anyio
```

Or with `uv`:

```bash
uv pip install claude-agent-sdk anyio
```

## Running the agent

Pipe a JSON config object into `agent_runner.py` via stdin:

```bash
echo '{
  "target_url": "https://example.com",
  "output_dir": "/tmp/agent-output",
  "playwright_mcp_url": "http://localhost:3000/sse"
}' | python3 agent_runner.py
```

## Configuration reference

| Field | Required | Default | Description |
|---|---|---|---|
| `target_url` | yes | — | URL of the web application to explore |
| `output_dir` | yes | — | Directory where screenshots and `result.json` are written |
| `playwright_mcp_url` | no | `http://playwright-mcp:3000/sse` | SSE endpoint of the Playwright MCP server |
| `procedure_name` | no | `"UI Exploration"` | Name used for the generated test procedure |
| `credentials` | no | `[]` | Array of `{"key": "...", "value": "..."}` objects for login |

## Environment variables

When using AWS Bedrock for model access:

| Variable | Description |
|---|---|
| `CLAUDE_CODE_USE_BEDROCK` | Set to `1` to use Bedrock |
| `AWS_REGION` | AWS region (e.g. `us-west-2`) |
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |

When using the Anthropic API directly, set `ANTHROPIC_API_KEY` instead.

## Output

After the agent finishes, the output directory will contain:

```
{output_dir}/
  result.json
  screenshots/
    01_landing_page.png
    02_login_form.png
    ...
```

`result.json` has the following structure:

```json
{
  "procedure_name": "UI Exploration",
  "description": "Brief description of what was explored",
  "steps": [
    {
      "name": "Short step name",
      "instructions": "Detailed instructions a manual tester can follow",
      "image_paths": ["screenshots/01_landing_page.png"]
    }
  ],
  "summary": "Overall summary of the exploration"
}
```

Screenshot paths inside `result.json` are relative to `output_dir`.

## Example with credentials

```bash
echo '{
  "target_url": "https://my-app.example.com",
  "output_dir": "/tmp/my-app-exploration",
  "playwright_mcp_url": "http://localhost:3000/sse",
  "procedure_name": "My App Login Flow",
  "credentials": [
    {"key": "username", "value": "testuser"},
    {"key": "password", "value": "s3cret"}
  ]
}' | python3 agent_runner.py
```
